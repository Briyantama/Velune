package httpx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const UserIDKey ctxKey = 1

type ridKey int

const RequestIDKey ridKey = 1

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
