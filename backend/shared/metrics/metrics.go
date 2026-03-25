package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	GatewayRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_requests_total",
		Help: "API gateway routed requests",
	}, []string{"route_group"})

	GatewayFallbackHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_fallback_hits_total",
		Help: "Gateway fallback or legacy catch-all usage",
	}, []string{"reason"})

	OutboxPending = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_pending_total",
		Help: "Approximate pending/failed outbox rows eligible for dispatch",
	})

	OutboxRetryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_retry_total",
		Help: "Outbox publish retries scheduled",
	}, []string{"service"})

	DLQMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dlq_messages_total",
		Help: "Messages routed to DLQ exchange",
	}, []string{"service"})

	NotificationSentTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_sent_total",
		Help: "Notification jobs completed successfully",
	})

	NotificationFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_failed_total",
		Help: "Notification jobs marked failed after max retries",
	})

	NotificationRetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_retry_total",
		Help: "Notification job retry schedules",
	})

	OverspendEventsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "overspend_events_total",
		Help: "Overspend alert events emitted from budget evaluation",
	})

	ReconciliationMismatchTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reconciliation_mismatch_total",
		Help: "Reconciliation detected a mismatch",
	}, []string{"type"})
)

// Handler serves Prometheus metrics on /metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}
