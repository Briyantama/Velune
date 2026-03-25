package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/notification-service/internal/domain"
	"github.com/moon-eye/velune/services/notification-service/internal/repository"
	"github.com/moon-eye/velune/shared/contracts"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
)

type OverspendService struct {
	Dedupe    repository.DedupeStore
	Jobs      repository.JobRepository
	InApp     repository.DeliveryChannel
	Email     repository.DeliveryChannel
	Publisher repository.EventPublisher
	MaxRetry  int
	BaseDelay time.Duration
}

func (s *OverspendService) HandleEnvelope(ctx context.Context, env contracts.EventEnvelope) error {
	if env.EventType != contracts.EventOverspendAlertRequested {
		return nil
	}
	if env.Idempotency == "" {
		env.Idempotency = fmt.Sprintf("%s:%s", env.EventType, env.EventID.String())
	}
	if s.Dedupe != nil {
		seen, err := s.Dedupe.SeenOrMark(ctx, env.Idempotency, env.EventID)
		if err != nil {
			return err
		}
		if seen {
			return nil
		}
	}
	var req contracts.OverspendAlertRequested
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		return errs.New("VALIDATION_ERROR", "invalid overspend payload",constx.StatusBadRequest)
	}
	payload, _ := json.Marshal(req)
	if s.Jobs == nil {
		return errs.New("INTERNAL_ERROR", "jobs repository not configured",constx.StatusInternalServerError)
	}
	if err := s.Jobs.Enqueue(ctx, &repository.NotificationJob{
		ID:      uuid.New(),
		UserID:  req.UserID,
		Channel: "in_app",
		Payload: payload,
	}); err != nil {
		return err
	}
	if req.UsagePercent >= 100 {
		if err := s.Jobs.Enqueue(ctx, &repository.NotificationJob{
			ID:      uuid.New(),
			UserID:  req.UserID,
			Channel: "email",
			Payload: payload,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *OverspendService) ProcessJob(ctx context.Context, j repository.NotificationJob) error {
	var req contracts.OverspendAlertRequested
	if err := json.Unmarshal(j.Payload, &req); err != nil {
		return err
	}
	alert := domain.OverspendAlert{
		BudgetID:         req.BudgetID,
		UserID:           req.UserID,
		Currency:         req.Currency,
		LimitAmountMinor: req.LimitAmountMinor,
		SpentMinor:       req.SpentMinor,
		UsagePercent:     req.UsagePercent,
		IsOverspent:      req.IsOverspent,
	}
	switch j.Channel {
	case "email":
		if s.Email == nil {
			return errs.New("INTERNAL_ERROR", "email channel not configured",constx.StatusInternalServerError)
		}
		if err := s.Email.Deliver(ctx, alert); err != nil {
			return err
		}
	default:
		if s.InApp == nil {
			return errs.New("INTERNAL_ERROR", "in-app channel not configured",constx.StatusInternalServerError)
		}
		if err := s.InApp.Deliver(ctx, alert); err != nil {
			return err
		}
	}
	if s.Publisher != nil {
		eventID := uuid.New()
		dispatched := contracts.NotificationDispatched{
			EventID:      eventID,
			Kind:         contracts.EventOverspendAlertRequested,
			Channel:      s.channelName(req.UsagePercent),
			Status:       "sent",
			DispatchedAt: time.Now().UTC(),
		}
		payload, _ := json.Marshal(dispatched)
		_ = s.Publisher.Publish(ctx, contracts.EventEnvelope{
			EventID:     eventID,
			EventType:   contracts.EventNotificationDispatched,
			Source:      "notification-service",
			OccurredAt:  time.Now().UTC(),
			UserID:      &req.UserID,
			Idempotency: fmt.Sprintf("%s:%s", contracts.EventNotificationDispatched, eventID.String()),
			Payload:     payload,
		})
	}
	return nil
}

func (s *OverspendService) Backoff(retry int) time.Duration {
	base := s.BaseDelay
	if base <= 0 {
		base = 2 * time.Second
	}
	if retry < 0 {
		retry = 0
	}
	return time.Duration(1<<retry) * base
}

func (s *OverspendService) channelName(usage float64) string {
	if usage >= 100 {
		return "inapp,email"
	}
	return "inapp"
}
