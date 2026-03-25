package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/notification-service/internal/domain"
	"github.com/moon-eye/velune/services/notification-service/internal/repository"
	"github.com/moon-eye/velune/shared/contracts"
)

type mockChannel struct{ delivered int }

func (m *mockChannel) Name() string { return "mock" }
func (m *mockChannel) Deliver(_ context.Context, _ domain.OverspendAlert) error {
	m.delivered++
	return nil
}

type mockDedupe struct {
	keys map[string]struct{}
}

func (m *mockDedupe) SeenOrMark(_ context.Context, key string, _ uuid.UUID) (bool, error) {
	if m.keys == nil {
		m.keys = map[string]struct{}{}
	}
	_, ok := m.keys[key]
	if ok {
		return true, nil
	}
	m.keys[key] = struct{}{}
	return false, nil
}

type mockJobs struct{ jobs []repository.NotificationJob }

func (m *mockJobs) Enqueue(_ context.Context, job *repository.NotificationJob) error {
	m.jobs = append(m.jobs, *job)
	return nil
}
func (m *mockJobs) FetchDue(_ context.Context, _ int) ([]repository.NotificationJob, error) { return nil, nil }
func (m *mockJobs) MarkSent(_ context.Context, _ uuid.UUID) error                             { return nil }
func (m *mockJobs) MarkRetry(_ context.Context, _ uuid.UUID, _ int, _ time.Time) error        { return nil }
func (m *mockJobs) MarkFailed(_ context.Context, _ uuid.UUID) error                            { return nil }

func TestOverspendChannelsByThreshold(t *testing.T) {
	inApp := &mockChannel{}
	email := &mockChannel{}
	uid := uuid.New()
	bid := uuid.New()
	base := contracts.OverspendAlertRequested{
		BudgetID:         bid,
		UserID:           uid,
		Currency:         "USD",
		LimitAmountMinor: 10000,
		SpentMinor:       9500,
		UsagePercent:     95,
		IsOverspent:      false,
	}
	p, _ := json.Marshal(base)
	jobs := &mockJobs{}
	s := &OverspendService{InApp: inApp, Email: email, Dedupe: &mockDedupe{}, Jobs: jobs}
	err := s.HandleEnvelope(context.Background(), contracts.EventEnvelope{
		EventID:     uuid.New(),
		EventType:   contracts.EventOverspendAlertRequested,
		OccurredAt:  time.Now().UTC(),
		Idempotency: "k1",
		Payload:     p,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(jobs.jobs) != 1 {
		t.Fatalf("expected one job for <100 usage")
	}

	base.UsagePercent = 110
	base.IsOverspent = true
	p, _ = json.Marshal(base)
	err = s.HandleEnvelope(context.Background(), contracts.EventEnvelope{
		EventID:     uuid.New(),
		EventType:   contracts.EventOverspendAlertRequested,
		OccurredAt:  time.Now().UTC(),
		Idempotency: "k2",
		Payload:     p,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(jobs.jobs) != 3 {
		t.Fatalf("expected in-app + email jobs at >=100 usage")
	}
}

func TestOverspendInvalidPayload(t *testing.T) {
	s := &OverspendService{Dedupe: &mockDedupe{}, Jobs: &mockJobs{}}
	err := s.HandleEnvelope(context.Background(), contracts.EventEnvelope{
		EventID:     uuid.New(),
		EventType:   contracts.EventOverspendAlertRequested,
		OccurredAt:  time.Now().UTC(),
		Idempotency: "k1",
		Payload:     []byte(`{`),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
