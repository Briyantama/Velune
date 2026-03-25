package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMustProxy_forwardsCorrelationAndRequestID(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Correlation-ID") != "corr-x" {
			t.Errorf("correlation header got %q", r.Header.Get("X-Correlation-ID"))
		}
		if r.Header.Get("X-Request-ID") != "req-x" {
			t.Errorf("request id header got %q", r.Header.Get("X-Request-ID"))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	proxy := MustProxy(upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := WithCorrelationID(req.Context(), "corr-x")
	ctx = WithRequestID(ctx, "req-x")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
}
