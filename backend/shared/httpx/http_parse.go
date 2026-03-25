package httpx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/otelx"
)

func ParsePageLimit(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	return page, limit
}

func DecodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errs.New("VALIDATION_ERROR", err.Error(), constx.StatusBadRequest)
	}
	return nil
}

func ParseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, name)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errs.New("VALIDATION_ERROR", "invalid id", constx.StatusBadRequest)
	}
	return id, nil
}

func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, errs.New("VALIDATION_ERROR", "invalid id", constx.StatusBadRequest)
	}
	return id, nil
}

func ParseTimeQuery(r *http.Request, key string) *time.Time {
	if v := r.URL.Query().Get(key); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

func ParseInt64Query(r *http.Request, key string) (int64, bool) {
	v := r.URL.Query().Get(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func ValidateStruct(s any) error {
	if err := validator.New().Struct(s); err != nil {
		return errs.New("VALIDATION_ERROR", err.Error(), constx.StatusBadRequest)
	}
	return nil
}

func MustUserID(r *http.Request) (uuid.UUID, error) {
	uid, ok := UserID(r.Context())
	if !ok || uid == uuid.Nil {
		return uuid.Nil, errs.ErrUnauthorized
	}
	return uid, nil
}

func singleHostReverseProxyWithTracing(u *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(u)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		if cid, ok := CorrelationID(req.Context()); ok {
			req.Header.Set("X-Correlation-ID", cid)
		}
		if rid, ok := RequestIDFromCtx(req.Context()); ok {
			req.Header.Set("X-Request-ID", rid)
		}
		otelx.InjectHTTP(req.Context(), req.Header)
	}
	return proxy
}

// MustProxy returns a reverse proxy that forwards X-Correlation-ID and X-Request-ID from the inbound request context.
func MustProxy(origin string) http.Handler {
	u, err := url.Parse(origin)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad upstream URL", constx.StatusInternalServerError)
		})
	}
	return singleHostReverseProxyWithTracing(u)
}

// SingleHostReverseProxy returns a proxy with correlation and request ID propagation on the outbound request.
func SingleHostReverseProxy(primary *url.URL) *httputil.ReverseProxy {
	return singleHostReverseProxyWithTracing(primary)
}

func MustProxyWithFallback(primaryURL, fallbackURL string) http.Handler {
	primary, err := url.Parse(primaryURL)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad report-service URL", constx.StatusInternalServerError)
		})
	}
	primaryProxy := httputil.NewSingleHostReverseProxy(primary)
	d := primaryProxy.Director
	primaryProxy.Director = func(req *http.Request) {
		d(req)
		if cid, ok := CorrelationID(req.Context()); ok {
			req.Header.Set("X-Correlation-ID", cid)
		}
		if rid, ok := RequestIDFromCtx(req.Context()); ok {
			req.Header.Set("X-Request-ID", rid)
		}
		otelx.InjectHTTP(req.Context(), req.Header)
	}
	var fallbackProxy http.Handler
	if fallbackURL != "" {
		fallbackProxy = MustProxy(fallbackURL)
	}
	primaryProxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == constx.StatusNotFound || resp.StatusCode >= constx.StatusInternalServerError {
			return errs.New("UPSTREAM_FALLBACK_TRIGGER", "report upstream fallback trigger", constx.StatusBadGateway)
		}
		return nil
	}
	primaryProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, _ error) {
		if fallbackProxy != nil {
			fallbackProxy.ServeHTTP(w, r)
			return
		}
		http.Error(w, "report upstream unavailable", constx.StatusBadGateway)
	}
	return primaryProxy
}

func NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}