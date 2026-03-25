package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moon-eye/velune/shared/httpx"
)

func TestCorrelationIDHeader_generatesWhenAbsent(t *testing.T) {
	var saw string
	h := CorrelationIDHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid, ok := httpx.CorrelationID(r.Context())
		if !ok || cid == "" {
			t.Fatal("expected correlation id in context")
		}
		saw = cid
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if saw == "" {
		t.Fatal("empty correlation id")
	}
	if got := rec.Header().Get("X-Correlation-ID"); got != saw {
		t.Fatalf("response header %q, context %q", got, saw)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID")
	}
}

func TestCorrelationIDHeader_preservesInbound(t *testing.T) {
	h := CorrelationIDHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cid, _ := httpx.CorrelationID(r.Context())
		if cid != "fixed-cid" {
			t.Fatalf("got %q", cid)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "fixed-cid")
	req.Header.Set("X-Request-ID", "fixed-rid")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Correlation-ID") != "fixed-cid" {
		t.Fatal()
	}
	if rec.Header().Get("X-Request-ID") != "fixed-rid" {
		t.Fatal()
	}
}
