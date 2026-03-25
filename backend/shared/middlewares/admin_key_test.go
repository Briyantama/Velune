package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestAdminAPIKeyAuth_rejectsMissingKey(t *testing.T) {
	h := AdminAPIKeyAuth("secret", zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rec.Code)
	}
}

func TestAdminAPIKeyAuth_acceptsKey(t *testing.T) {
	var saw bool
	h := AdminAPIKeyAuth("correct", zap.NewNop())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		saw = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(AdminAPIKeyHeader, "correct")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !saw || rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d saw=%v", rec.Code, saw)
	}
}
