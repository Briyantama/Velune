package logger

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/shared/httpx"
)

func TestFieldsFromContext_populatesIDs(t *testing.T) {
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ctx := context.Background()
	ctx = httpx.WithCorrelationID(ctx, "corr-1")
	ctx = httpx.WithRequestID(ctx, "req-1")
	ctx = httpx.WithUserID(ctx, uid)
	fields := FieldsFromContext(ctx)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
}
