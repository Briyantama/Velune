package httpapi

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	constx "github.com/moon-eye/velune/shared/constx"
	"go.uber.org/zap"
)

func TestNotificationPingRequiresAuth(t *testing.T) {
	h := NewRouter(&Server{Log: zap.NewNop(), JWTSecret: "secret"})
	req := httptest.NewRequest(constx.MethodGet, "/api/v1/notifications/ping", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code !=constx.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rec.Code)
	}
}

func TestNotificationPingWithInternalUserHeader(t *testing.T) {
	h := NewRouter(&Server{Log: zap.NewNop(), JWTSecret: "secret"})
	req := httptest.NewRequest(constx.MethodGet, "/api/v1/notifications/ping", nil)
	req.Header.Set("X-User-ID", uuid.NewString())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code !=constx.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}
