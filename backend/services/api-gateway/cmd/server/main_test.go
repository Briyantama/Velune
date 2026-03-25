package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sharedconfig "github.com/moon-eye/velune/shared/config"
	constx "github.com/moon-eye/velune/shared/constx"
)

func TestReportsProxyWithFallback_On404(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(constx.StatusNotFound)
	}))
	defer primary.Close()
	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(constx.StatusOK)
		_, _ = w.Write([]byte("legacy"))
	}))
	defer fallback.Close()

	h := reportsProxyWithFallback(primary.URL, fallback.URL)
	req := httptest.NewRequest(constx.MethodGet, "/api/v1/reports/monthly?year=2026&month=3", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != constx.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReportsProxyWithFallback_NoFallback(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(constx.StatusInternalServerError)
	}))
	defer primary.Close()

	h := reportsProxyWithFallback(primary.URL, "")
	req := httptest.NewRequest(constx.MethodGet, "/api/v1/reports/monthly?year=2026&month=3", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != constx.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

func TestPickProxy_NotificationsRoute(t *testing.T) {
	cfg := &sharedconfig.Service{
		NotificationServiceURL: "http://127.0.0.1:8086",
	}
	h := pickProxy(cfg, "/api/v1/notifications/ping")
	if h == nil {
		t.Fatal("expected notification proxy handler")
	}
}
