package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/report-service/internal/usecase"
	"go.uber.org/zap"
)

func TestMonthlyValidation_MissingYear(t *testing.T) {
	srv := &Server{
		Reports:   &usecase.ReportService{Transactions: nil},
		Log:       zap.NewNop(),
		JWTSecret: "secret",
	}
	handler := NewRouter(srv)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?month=3", nil)
	req.Header.Set("X-User-ID", uuid.NewString())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMonthlyValidation_MissingAuth(t *testing.T) {
	srv := &Server{
		Reports:   &usecase.ReportService{Transactions: nil},
		Log:       zap.NewNop(),
		JWTSecret: "secret",
	}
	handler := NewRouter(srv)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2026&month=3", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d got %d", http.StatusUnauthorized, rec.Code)
	}
}
