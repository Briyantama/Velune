package logger

import (
	"context"

	"github.com/moon-eye/velune/shared/httpx"
	"go.uber.org/zap"
)

// FieldsFromContext returns zap fields for structured request/event tracing.
func FieldsFromContext(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}
	var fs []zap.Field
	if cid, ok := httpx.CorrelationID(ctx); ok {
		fs = append(fs, zap.String("correlation_id", cid))
	}
	if rid, ok := httpx.RequestIDFromCtx(ctx); ok {
		fs = append(fs, zap.String("request_id", rid))
	}
	if uid, ok := httpx.UserID(ctx); ok {
		fs = append(fs, zap.String("user_id", uid.String()))
	}
	return fs
}
