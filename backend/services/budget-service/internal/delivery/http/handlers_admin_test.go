package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

func TestInternalAdminReconcileBudget_requiresAdminKey(t *testing.T) {
	s := &Server{
		AdminInternalKey: "admin-secret",
		Log:              zap.NewNop(),
		DB:               nil,
	}
	mux := NewRouter(s)
	req := httptest.NewRequest(http.MethodPost, "/internal/admin/reconcile/budget", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without key, got %d", rec.Code)
	}
	req.Header.Set(middlewares.AdminAPIKeyHeader, "wrong")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad key, got %d", rec.Code)
	}
}
