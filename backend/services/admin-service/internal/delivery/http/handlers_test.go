package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moon-eye/velune/services/admin-service/internal/config"
	"github.com/moon-eye/velune/shared/middlewares"
	"go.uber.org/zap"
)

func TestAdminRoutes_requireAPIKey(t *testing.T) {
	cfg := &config.Config{AdminAPIKey: "api-secret"}
	h := NewHandlers(cfg, zap.NewNop(), nil, nil, nil, nil)
	mux := h.Routes()
	req := httptest.NewRequest(http.MethodGet, "/internal/admin/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rec.Code)
	}
}

func TestAdminRoutes_dashboardWithKey(t *testing.T) {
	cfg := &config.Config{
		AdminAPIKey: "api-secret",
	}
	h := NewHandlers(cfg, zap.NewNop(), nil, nil, nil, nil)
	mux := h.Routes()
	req := httptest.NewRequest(http.MethodGet, "/internal/admin/health", nil)
	req.Header.Set(middlewares.AdminAPIKeyHeader, "api-secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d body=%s", rec.Code, rec.Body.String())
	}
}
