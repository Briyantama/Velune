package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/middlewares"
)

func decodeResponseBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var obj map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &obj); err != nil {
		t.Fatalf("decode json: %v body=%s", err, rec.Body.String())
	}
	return obj
}

func TestResponseEnvelope_successShape(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("X-Request-ID", "req-x")

	rec := httptest.NewRecorder()
	h := middlewares.CorrelationIDHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, constx.StatusOK, map[string]any{"hello": "world"})
	}))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d body=%s", rec.Code, rec.Body.String())
	}

	obj := decodeResponseBody(t, rec)

	ts, ok := obj["timestamp"].(string)
	if !ok || ts == "" {
		t.Fatal("expected timestamp string")
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000Z07:00", ts); err != nil {
		t.Fatalf("timestamp not RFC3339Milli: %v (%q)", err, ts)
	}
	if obj["path"] != "/api/v1/test" {
		t.Fatalf("path mismatch: %#v", obj["path"])
	}
	if obj["status"] != float64(http.StatusOK) {
		t.Fatalf("status mismatch: %#v", obj["status"])
	}
	if obj["requestId"] != "req-x" {
		t.Fatalf("requestId mismatch: %#v", obj["requestId"])
	}

	if _, hasError := obj["error"]; hasError {
		t.Fatalf("did not expect error field: %#v", obj["error"])
	}

	data, ok := obj["data"].(map[string]any)
	if !ok || data["hello"] != "world" {
		t.Fatalf("unexpected data: %#v", obj["data"])
	}
}

func TestResponseEnvelope_errorShape(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil)
	req.Header.Set("X-Request-ID", "req-404")

	rec := httptest.NewRecorder()
	h := middlewares.CorrelationIDHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, constx.StatusNotFound, map[string]any{"ignored": "payload"})
	}))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 got %d body=%s", rec.Code, rec.Body.String())
	}

	obj := decodeResponseBody(t, rec)

	ts, ok := obj["timestamp"].(string)
	if !ok || ts == "" {
		t.Fatal("expected timestamp string")
	}
	if _, err := time.Parse("2006-01-02T15:04:05.000Z07:00", ts); err != nil {
		t.Fatalf("timestamp not RFC3339Milli: %v (%q)", err, ts)
	}
	if obj["path"] != "/api/v1/missing" {
		t.Fatalf("path mismatch: %#v", obj["path"])
	}
	if obj["status"] != float64(http.StatusNotFound) {
		t.Fatalf("status mismatch: %#v", obj["status"])
	}
	if obj["requestId"] != "req-404" {
		t.Fatalf("requestId mismatch: %#v", obj["requestId"])
	}

	if obj["error"] != http.StatusText(http.StatusNotFound) {
		t.Fatalf("error mismatch: %#v", obj["error"])
	}
	if _, hasData := obj["data"]; hasData {
		t.Fatalf("did not expect data field: %#v", obj["data"])
	}
}

func TestResponseEnvelope_WriteError_usesAppErrorStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/unauthorized", nil)
	req.Header.Set("X-Request-ID", "req-401")

	rec := httptest.NewRecorder()
	h := middlewares.CorrelationIDHeader(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, errs.ErrUnauthorized)
	}))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 got %d body=%s", rec.Code, rec.Body.String())
	}

	obj := decodeResponseBody(t, rec)
	if obj["error"] != http.StatusText(http.StatusUnauthorized) {
		t.Fatalf("error mismatch: %#v", obj["error"])
	}
}

