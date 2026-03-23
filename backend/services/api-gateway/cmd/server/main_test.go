package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReportsProxyWithFallback_On404(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer primary.Close()
	fallback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("legacy"))
	}))
	defer fallback.Close()

	h := reportsProxyWithFallback(primary.URL, fallback.URL)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2026&month=3", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReportsProxyWithFallback_NoFallback(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	h := reportsProxyWithFallback(primary.URL, "")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/monthly?year=2026&month=3", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}
