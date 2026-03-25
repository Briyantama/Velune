package httpapi

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/report-service/internal/usecase"
	constx "github.com/moon-eye/velune/shared/constx"
	"go.uber.org/zap"
)

func TestMonthlyValidation_MissingYear(t *testing.T) {
	srv := &Server{
		Reports:   &usecase.ReportService{Transactions: nil},
		Log:       zap.NewNop(),
		JWTSecret: "secret",
	}
	handler := NewRouter(srv)
	req := httptest.NewRequest(constx.MethodGet, "/api/v1/reports/monthly?month=3", nil)
	req.Header.Set("X-User-ID", uuid.NewString())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != constx.StatusBadRequest {
		t.Fatalf("expected %d got %d",constx.StatusBadRequest, rec.Code)
	}
}

func TestMonthlyValidation_MissingAuth(t *testing.T) {
	srv := &Server{
		Reports:   &usecase.ReportService{Transactions: nil},
		Log:       zap.NewNop(),
		JWTSecret: "secret",
	}
	handler := NewRouter(srv)
	req := httptest.NewRequest(constx.MethodGet, "/api/v1/reports/monthly?year=2026&month=3", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != constx.StatusUnauthorized {
		t.Fatalf("expected %d got %d",constx.StatusUnauthorized, rec.Code)
	}
}
