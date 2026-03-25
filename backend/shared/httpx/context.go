package httpx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const UserIDKey ctxKey = 1

type ridKey int

const RequestIDKey ridKey = 1

type cidKey int

const CorrelationIDKey cidKey = 1

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(UserIDKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// WithCorrelationID stores the distributed trace correlation id (typically X-Correlation-ID).
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return ctx
	}
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// CorrelationID returns the correlation id when present.
func CorrelationID(ctx context.Context) (string, bool) {
	v := ctx.Value(CorrelationIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// WithRequestID stores X-Request-ID in context (distinct from correlation id).
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// RequestIDFromCtx returns the per-request id when present.
func RequestIDFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(RequestIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}
