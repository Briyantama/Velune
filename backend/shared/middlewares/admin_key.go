package middlewares

import (
	"crypto/subtle"
	"net/http"

	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/httpx"
	"go.uber.org/zap"
)

const AdminAPIKeyHeader = "X-Admin-Key"

// AdminAPIKeyAuth requires a non-empty secret and matching header (constant-time).
func AdminAPIKeyAuth(secret string, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				if log != nil {
					log.Warn("admin_auth_misconfigured_empty_secret")
				}
				httpx.WriteJSON(w, constx.StatusInternalServerError, map[string]string{"code": "ADMIN_AUTH_MISCONFIGURED", "message": "admin secret not set"})
				return
			}
			got := r.Header.Get(AdminAPIKeyHeader)
			if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
				if log != nil {
					log.Warn("admin_auth_failed", zap.String("path", r.URL.Path))
				}
				httpx.WriteJSON(w, constx.StatusUnauthorized, map[string]string{"code": "UNAUTHORIZED", "message": "invalid admin key"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
