package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/httpx"
	"github.com/moon-eye/velune/shared/jwt"
	"go.uber.org/zap"
)

func JWTAuth(secret string, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if internalUID := strings.TrimSpace(r.Header.Get("X-User-ID")); internalUID != "" {
				if uid, err := uuid.Parse(internalUID); err == nil && uid != uuid.Nil {
					ctx := httpx.WithUserID(r.Context(), uid)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				httpx.WriteError(w, errs.ErrUnauthorized)
				return
			}
			raw := strings.TrimSpace(h[7:])
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
		ctx := context.WithValue(r.Context(), httpx.RequestIDKey, rid)
		w.Header().Set("X-Request-ID", rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
