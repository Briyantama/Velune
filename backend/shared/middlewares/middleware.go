package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/jwt"
	"github.com/moon-eye/velune/shared/stringx"
	"go.uber.org/zap"
)

func JWTAuth(secret string, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if internalUID := stringx.TrimSpace(r.Header.Get("X-User-ID")); internalUID != "" {
				if uid, err := uuid.Parse(internalUID); err == nil && uid != uuid.Nil {
					ctx := httpx.WithUserID(r.Context(), uid)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			h := r.Header.Get("Authorization")
			if h == "" || !stringx.HasPrefix(stringx.Lower(h), "bearer ") {
				httpx.WriteError(w, errs.ErrUnauthorized)
				return
			}
			raw := stringx.TrimSpace(h[7:])
			claims, err := jwt.Parse(raw, secret)
			if err != nil {
				httpx.WriteError(w, errs.ErrUnauthorized)
				return
			}
			uid := claims.UserID
			if uid == uuid.Nil {
				var err error
				uid, err = uuid.Parse(claims.Subject)
				if err != nil {
					log.Debug("jwt subject parse", zap.Error(err))
					httpx.WriteError(w, errs.ErrUnauthorized)
					return
				}
			}
			ctx := httpx.WithUserID(r.Context(), uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDHeader ensures X-Request-ID exists for tracing.
func RequestIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		ctx := httpx.WithRequestID(r.Context(), rid)
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CorrelationIDHeader ensures X-Correlation-ID exists and stores it in context with X-Request-ID.
func CorrelationIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Used by shared/httpx response helpers to automatically populate the envelope `path`.
		w.Header().Set("X-Request-Path", r.URL.Path)

		cid := stringx.TrimSpace(r.Header.Get("X-Correlation-ID"))
		if cid == "" {
			cid = uuid.New().String()
		}
		w.Header().Set("X-Correlation-ID", cid)
		ctx := httpx.WithCorrelationID(r.Context(), cid)

		rid := stringx.TrimSpace(r.Header.Get("X-Request-ID"))
		if rid == "" {
			rid = uuid.New().String()
			w.Header().Set("X-Request-ID", rid)
		} else {
			w.Header().Set("X-Request-ID", rid)
		}
		ctx = httpx.WithRequestID(ctx, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
