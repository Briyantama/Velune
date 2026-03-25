package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moon-eye/velune/services/admin-service/internal/config"
	"github.com/moon-eye/velune/services/admin-service/internal/dlq"
	"github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/events"
	"github.com/moon-eye/velune/shared/helper"
	"github.com/moon-eye/velune/shared/httpx"
	sharedlog "github.com/moon-eye/velune/shared/logger"
	"github.com/moon-eye/velune/shared/metrics"
	"github.com/moon-eye/velune/shared/middlewares"
	"github.com/moon-eye/velune/shared/otelx"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
	stringx "github.com/moon-eye/velune/shared/stringx"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Handlers implements /internal/admin APIs.
type Handlers struct {
	cfg    *config.Config
	log    *zap.Logger
	txPool *pgxpool.Pool
	bdPool *pgxpool.Pool
	nfPool *pgxpool.Pool
	pub    *events.RabbitPublisher
	httpc  *http.Client
}

func NewHandlers(cfg *config.Config, log *zap.Logger, txPool, bdPool, nfPool *pgxpool.Pool, pub *events.RabbitPublisher) *Handlers {
	return &Handlers{
		cfg:    cfg,
		log:    log,
		txPool: txPool,
		bdPool: bdPool,
		nfPool: nfPool,
		pub:    pub,
		httpc:  otelx.TracedHTTPClient(&http.Client{Timeout: 8 * time.Second}),
	}
}

func (h *Handlers) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CorrelationIDHeader)

	r.Get("/health", h.publicHealth)
	r.Handle("/metrics", metrics.Handler())

	r.Route("/internal/admin", func(r chi.Router) {
		r.Use(middlewares.AdminAPIKeyAuth(h.cfg.AdminAPIKey, h.log))
		r.Get("/health", h.dashboardHealth)
		r.Get("/dlq", h.getDLQ)
		r.Post("/dlq/replay", h.postDLQReplay)
		r.Get("/outbox", h.getOutbox)
		r.Post("/outbox/retry", h.postOutboxRetry)
		r.Get("/notifications/jobs", h.getNotificationJobs)
		r.Post("/notifications/jobs/retry", h.postNotificationJobRetry)
		r.Post("/reconcile/balance", h.postReconcileBalance)
		r.Post("/reconcile/budget", h.postReconcileBudget)
		r.Get("/reconcile/logs", h.getReconcileLogs)
		r.Post("/events/replay", h.postEventsReplay)
	})
	return r
}

func (h *Handlers) publicHealth(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{"status": "ok", "service": "admin-service"})
}

func (h *Handlers) adminFields(r *http.Request, action string) []zap.Field {
	return append(sharedlog.FieldsFromContext(r.Context()),
		zap.String("admin_action", action),
		zap.String("admin_actor", "api_key"),
	)
}

func (h *Handlers) dashboardHealth(w http.ResponseWriter, r *http.Request) {
	svcs := map[string]string{
		"auth":         h.probe(r.Context(), h.cfg.AuthServiceURL+"/health"),
		"transaction":  h.probe(r.Context(), h.cfg.TransactionServiceURL+"/health"),
		"budget":       h.probe(r.Context(), h.cfg.BudgetServiceURL+"/health"),
		"report":       h.probe(r.Context(), h.cfg.ReportServiceURL+"/health"),
		"notification": h.probe(r.Context(), h.cfg.NotificationServiceURL+"/health"),
		"category":     h.probe(r.Context(), h.cfg.CategoryServiceURL+"/api/v1/meta"),
	}
	outboxPending := 0
	if h.txPool != nil {
		n, _ := db.New(h.txPool).OutboxCountPendingOrFailed(r.Context())
		outboxPending += int(n)
	}
	if h.bdPool != nil {
		n, _ := db.New(h.bdPool).OutboxCountPendingOrFailed(r.Context())
		outboxPending += int(n)
	}

	dlqCount := 0
	if h.cfg.BrokerDLQ != "" {
		conn, err := amqp.Dial(h.cfg.BrokerURL)
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				dlqCount, _ = dlq.Depth(ch, h.cfg.BrokerDLQ)
				_ = ch.Close()
			}
			_ = conn.Close()
		}
	}
	notifFail := 0
	if h.nfPool != nil {
		n, _ := db.New(h.nfPool).NotificationJobsCountFailedOrPending(r.Context())
		notifFail = int(n)
	}
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{
		"services":              svcs,
		"outbox_pending":        outboxPending,
		"dlq_messages":          dlqCount,
		"notification_failures": notifFail,
	})
}

func (h *Handlers) probe(ctx context.Context, base string) string {
	if stringx.TrimSpace(base) == "" {
		return "unknown"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
	if err != nil {
		return "error"
	}
	resp, err := h.httpc.Do(req)
	if err != nil {
		return "down"
	}
	defer resp.Body.Close()
	if resp.StatusCode < 300 {
		return "ok"
	}
	return "degraded"
}

func (h *Handlers) getDLQ(w http.ResponseWriter, r *http.Request) {
	if h.cfg.BrokerDLQ == "" {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "DLQ_NOT_CONFIGURED", "message": "BROKER_DLQ is empty"})
		return
	}
	limit := config.AtoiDefault(r.URL.Query().Get("limit"), 50)
	conn, err := amqp.Dial(h.cfg.BrokerURL)
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "BROKER_ERROR", "message": err.Error()})
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "CHANNEL_ERROR", "message": err.Error()})
		return
	}
	defer ch.Close()
	msgs, err := dlq.Peek(r.Context(), ch, h.cfg.BrokerDLQ, limit)
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "DLQ_PEEK", "message": err.Error()})
		return
	}
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{"messages": msgs})
}

type dlqReplayReq struct {
	EventID string `json:"event_id"`
	Target  string `json:"target"`
}

func (h *Handlers) postDLQReplay(w http.ResponseWriter, r *http.Request) {
	var req dlqReplayReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	req.Target = stringx.TrimSpace(stringx.Lower(req.Target))
	if req.Target != "" && req.Target != "original" {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "UNSUPPORTED_TARGET", "message": "only target=original is supported"})
		return
	}
	eid := stringx.TrimSpace(req.EventID)
	if eid == "" {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "event_id required"})
		return
	}
	if _, err := uuid.Parse(eid); err != nil {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "event_id must be a UUID"})
		return
	}
	if h.cfg.BrokerDLQ == "" || h.pub == nil {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "DLQ_NOT_CONFIGURED", "message": "broker not configured"})
		return
	}
	conn, err := amqp.Dial(h.cfg.BrokerURL)
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "BROKER_ERROR", "message": err.Error()})
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "CHANNEL_ERROR", "message": err.Error()})
		return
	}
	defer ch.Close()

	h.log.Info("admin_dlq_replay_start", append(h.adminFields(r, "dlq_replay"), zap.String("event_id", eid))...)

	err = dlq.Replay(r.Context(), ch, h.cfg.BrokerDLQ, eid, func(ctx context.Context, env contracts.EventEnvelope) error {
		return h.pub.Publish(ctx, env)
	})
	if err != nil {
		h.log.Error("admin_dlq_replay_failed", append(h.adminFields(r, "dlq_replay"), zap.Error(err), zap.String("event_id", eid))...)
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "REPLAY_FAILED", "message": err.Error()})
		return
	}
	h.log.Info("admin_dlq_replay_ok", append(h.adminFields(r, "dlq_replay"), zap.String("event_id", eid))...)
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{"status": "replayed", "event_id": eid})
}

func (h *Handlers) getOutbox(w http.ResponseWriter, r *http.Request) {
	svc := stringx.Lower(stringx.TrimSpace(r.URL.Query().Get("service")))
	if svc == "" {
		svc = "all"
	}
	status := stringx.TrimSpace(r.URL.Query().Get("status"))
	limit := config.AtoiDefault(r.URL.Query().Get("limit"), 50)
	if limit > 200 {
		limit = 200
	}

	type row struct {
		Service    string    `json:"service"`
		ID         string    `json:"id"`
		EventType  string    `json:"event_type"`
		Status     string    `json:"status"`
		RetryCount int       `json:"retry_count"`
		NextRetry  time.Time `json:"next_retry_at"`
		CreatedAt  time.Time `json:"created_at"`
		Preview    string    `json:"payload_preview"`
	}

	out := map[string][]row{}
	appendRows := func(name string, pool *pgxpool.Pool) error {
		if pool == nil {
			return nil
		}
		dbrows, err := db.New(pool).OutboxListAdmin(r.Context(), db.OutboxListAdminParams{
			Column1: status,
			Limit:   int32(limit),
		})
		if err != nil {
			return err
		}
		list := make([]row, 0, len(dbrows))
		for _, dr := range dbrows {
			list = append(list, row{
				Service:    name,
				ID:         dr.ID,
				EventType:  dr.EventType,
				Status:     dr.Status,
				RetryCount: int(dr.RetryCount),
				NextRetry:  timestamptzToTime(dr.NextRetryAt),
				CreatedAt:  timestamptzToTime(dr.CreatedAt),
				Preview:    dr.PayloadPreview,
			})
		}
		out[name] = list
		return nil
	}
	switch svc {
	case "transaction":
		if err := appendRows("transaction", h.txPool); err != nil {
			httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "QUERY", "message": err.Error()})
			return
		}
	case "budget":
		if err := appendRows("budget", h.bdPool); err != nil {
			httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "QUERY", "message": err.Error()})
			return
		}
	case "all":
		_ = appendRows("transaction", h.txPool)
		_ = appendRows("budget", h.bdPool)
	default:
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "service must be transaction, budget, or all"})
		return
	}
	httpx.WriteJSON(w, constx.StatusOK, out)
}

type outboxRetryReq struct {
	Service  string `json:"service"`
	OutboxID string `json:"outbox_id"`
}

func (h *Handlers) postOutboxRetry(w http.ResponseWriter, r *http.Request) {
	var req outboxRetryReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(stringx.TrimSpace(req.OutboxID))
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "invalid outbox_id"})
		return
	}
	var pool *pgxpool.Pool
	switch stringx.Lower(stringx.TrimSpace(req.Service)) {
	case "transaction":
		pool = h.txPool
	case "budget":
		pool = h.bdPool
	default:
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "service must be transaction or budget"})
		return
	}
	if pool == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_DB", "message": "database pool not configured"})
		return
	}
	affected, err := db.New(pool).OutboxRetryReset(r.Context(), helper.ToPgUUID(id))
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "UPDATE", "message": err.Error()})
		return
	}
	if affected == 0 {
		httpx.WriteJSON(w, constx.StatusNotFound, map[string]string{"code": "NOT_FOUND", "message": "outbox row not found"})
		return
	}
	h.log.Info("admin_outbox_retry", append(h.adminFields(r, "outbox_retry"),
		zap.String("outbox_id", id.String()),
		zap.String("service", req.Service),
	)...)
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{"status": "scheduled", "outbox_id": id.String()})
}

func (h *Handlers) getNotificationJobs(w http.ResponseWriter, r *http.Request) {
	if h.nfPool == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_DB", "message": "notification database not configured"})
		return
	}
	status := stringx.TrimSpace(r.URL.Query().Get("status"))
	limit := config.AtoiDefault(r.URL.Query().Get("limit"), 50)
	dbrows, err := db.New(h.nfPool).NotificationJobsListAdmin(r.Context(), db.NotificationJobsListAdminParams{
		Column1: status,
		Limit:   int32(limit),
	})
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "QUERY", "message": err.Error()})
		return
	}
	type jr struct {
		ID          string    `json:"id"`
		UserID      string    `json:"user_id"`
		Channel     string    `json:"channel"`
		Status      string    `json:"status"`
		RetryCount  int       `json:"retry_count"`
		NextRetryAt time.Time `json:"next_retry_at"`
		CreatedAt   time.Time `json:"created_at"`
		Preview     string    `json:"payload_preview"`
	}
	jobs := make([]jr, 0, len(dbrows))
	for _, dr := range dbrows {
		jobs = append(jobs, jr{
			ID:          dr.ID,
			UserID:      dr.UserID,
			Channel:     dr.Channel,
			Status:      dr.Status,
			RetryCount:  int(dr.RetryCount),
			NextRetryAt: timestamptzToTime(dr.NextRetryAt),
			CreatedAt:   timestamptzToTime(dr.CreatedAt),
			Preview:     dr.PayloadPreview,
		})
	}
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{"jobs": jobs})
}

type jobRetryReq struct {
	JobID string `json:"job_id"`
}

func (h *Handlers) postNotificationJobRetry(w http.ResponseWriter, r *http.Request) {
	var req jobRetryReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	id, err := uuid.Parse(stringx.TrimSpace(req.JobID))
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "invalid job_id"})
		return
	}
	if h.nfPool == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_DB", "message": "notification database not configured"})
		return
	}
	affected, err := db.New(h.nfPool).NotificationJobRetryReset(r.Context(), helper.ToPgUUID(id))
	if err != nil {
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "UPDATE", "message": err.Error()})
		return
	}
	if affected == 0 {
		httpx.WriteJSON(w, constx.StatusNotFound, map[string]string{"code": "NOT_FOUND", "message": "job not found or not retryable"})
		return
	}
	h.log.Info("admin_notification_job_retry", append(h.adminFields(r, "notification_job_retry"), zap.String("job_id", id.String()))...)
	httpx.WriteJSON(w, constx.StatusOK, map[string]string{"status": "scheduled", "job_id": id.String()})
}

func (h *Handlers) postReconcileBalance(w http.ResponseWriter, r *http.Request) {
	url := stringx.TrimRight(h.cfg.TransactionServiceURL, "/") + "/internal/admin/reconcile/balance"
	h.proxyInternal(w, r, http.MethodPost, url, "reconcile_balance")
}

func (h *Handlers) postReconcileBudget(w http.ResponseWriter, r *http.Request) {
	url := stringx.TrimRight(h.cfg.BudgetServiceURL, "/") + "/internal/admin/reconcile/budget"
	h.proxyInternal(w, r, http.MethodPost, url, "reconcile_budget")
}

func (h *Handlers) proxyInternal(w http.ResponseWriter, r *http.Request, method, url, action string) {
	if h.cfg.AdminInternalKey == "" {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "MISCONFIGURED", "message": "ADMIN_INTERNAL_KEY not set on admin-service"})
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), method, url, nil)
	if err != nil {
		httpx.WriteJSON(w, constx.StatusInternalServerError, map[string]string{"code": "REQUEST", "message": err.Error()})
		return
	}
	req.Header.Set(middlewares.AdminAPIKeyHeader, h.cfg.AdminInternalKey)
	if cid, ok := httpx.CorrelationID(r.Context()); ok {
		req.Header.Set("X-Correlation-ID", cid)
	}
	h.log.Info("admin_proxy_reconcile", append(h.adminFields(r, action), zap.String("url", url))...)
	resp, err := h.httpc.Do(req)
	if err != nil {
		h.log.Error("admin_proxy_error", append(h.adminFields(r, action), zap.Error(err))...)
		httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "UPSTREAM", "message": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func (h *Handlers) getReconcileLogs(w http.ResponseWriter, r *http.Request) {
	typ := stringx.TrimSpace(r.URL.Query().Get("type"))
	limit := config.AtoiDefault(r.URL.Query().Get("limit"), 50)
	svc := stringx.Lower(stringx.TrimSpace(r.URL.Query().Get("service")))
	if svc == "" {
		svc = "all"
	}
	type logRow struct {
		Service   string          `json:"service"`
		ID        string          `json:"id"`
		Type      string          `json:"type"`
		Status    string          `json:"status"`
		Details   json.RawMessage `json:"details"`
		CreatedAt time.Time       `json:"created_at"`
	}
	var all []logRow
	appendLogs := func(name string, pool *pgxpool.Pool) error {
		if pool == nil {
			return nil
		}
		dbrows, err := db.New(pool).AuditLogsList(r.Context(), db.AuditLogsListParams{
			Column1: typ,
			Limit:   int32(limit),
		})
		if err != nil {
			return err
		}
		for _, dr := range dbrows {
			all = append(all, logRow{
				Service:   name,
				ID:        dr.ID,
				Type:      dr.Type,
				Status:    dr.Status,
				Details:   json.RawMessage(dr.Details),
				CreatedAt: timestamptzToTime(dr.CreatedAt),
			})
		}
		return nil
	}
	switch svc {
	case "transaction":
		if err := appendLogs("transaction", h.txPool); err != nil {
			httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "QUERY", "message": err.Error()})
			return
		}
	case "budget":
		if err := appendLogs("budget", h.bdPool); err != nil {
			httpx.WriteJSON(w, constx.StatusBadGateway, map[string]string{"code": "QUERY", "message": err.Error()})
			return
		}
	case "all":
		_ = appendLogs("transaction", h.txPool)
		_ = appendLogs("budget", h.bdPool)
	default:
		httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "service must be transaction, budget, or all"})
		return
	}
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{"logs": all})
}

type eventsReplayReq struct {
	EventType string `json:"event_type"`
	From      string `json:"from"`
	To        string `json:"to"`
}

func (h *Handlers) postEventsReplay(w http.ResponseWriter, r *http.Request) {
	dry := r.URL.Query().Get("dry_run") == "true" || r.URL.Query().Get("dry_run") == "1"
	var req eventsReplayReq
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	fromT := time.Now().UTC().Add(-24 * time.Hour)
	toT := time.Now().UTC()
	if req.From != "" {
		t, err := time.Parse(time.RFC3339, req.From)
		if err != nil {
			httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "invalid from RFC3339"})
			return
		}
		fromT = t
	}
	if req.To != "" {
		t, err := time.Parse(time.RFC3339, req.To)
		if err != nil {
			httpx.WriteJSON(w, constx.StatusBadRequest, map[string]string{"code": "VALIDATION", "message": "invalid to RFC3339"})
			return
		}
		toT = t
	}
	if h.pub == nil {
		httpx.WriteJSON(w, constx.StatusServiceUnavailable, map[string]string{"code": "NO_PUBLISHER", "message": "rabbit publisher not configured"})
		return
	}
	et := stringx.TrimSpace(req.EventType)
	replayArg := db.OutboxPayloadsForReplayParams{
		CreatedAt:   helper.ToPgTS(fromT),
		CreatedAt_2: helper.ToPgTS(toT),
		Column3:     et,
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var replayed []string
	var errs []string
	runPool := func(name string, pool *pgxpool.Pool) {
		if pool == nil {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			payloads, err := db.New(pool).OutboxPayloadsForReplay(r.Context(), replayArg)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s: %v", name, err))
				mu.Unlock()
				return
			}
			for _, payload := range payloads {
				var env contracts.EventEnvelope
				if err := json.Unmarshal([]byte(payload), &env); err != nil {
					continue
				}
				if dry {
					mu.Lock()
					replayed = append(replayed, env.EventID.String()+":"+name+":dry_run")
					mu.Unlock()
					continue
				}
				if err := h.pub.Publish(r.Context(), env); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Sprintf("%s %s: %v", name, env.EventID, err))
					mu.Unlock()
					continue
				}
				mu.Lock()
				replayed = append(replayed, env.EventID.String()+":"+name)
				mu.Unlock()
			}
		}()
	}
	runPool("transaction", h.txPool)
	runPool("budget", h.bdPool)
	wg.Wait()
	h.log.Info("admin_events_replay", append(h.adminFields(r, "events_replay"),
		zap.Bool("dry_run", dry),
		zap.Int("published", len(replayed)),
		zap.Strings("errors", errs),
	)...)
	httpx.WriteJSON(w, constx.StatusOK, map[string]any{
		"dry_run":    dry,
		"replayed":   replayed,
		"errors":     errs,
		"window":     map[string]string{"from": fromT.Format(time.RFC3339), "to": toT.Format(time.RFC3339)},
		"event_type": et,
	})
}

func timestamptzToTime(v pgtype.Timestamptz) time.Time {
	if !v.Valid {
		return time.Time{}
	}
	return v.Time
}
