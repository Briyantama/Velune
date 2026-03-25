//go:build integration

package postgres

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	db "github.com/moon-eye/velune/shared/sqlc/generated"
)

// TestDedupeRepo_SeenOrMark_Concurrent requires INTEGRATION_DATABASE_URL (e.g. postgres://postgres:postgres@localhost:5432/velune_notification?sslmode=disable).
func TestDedupeRepo_SeenOrMark_Concurrent(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("INTEGRATION_DATABASE_URL not set")
	}
	ctx := context.Background()
	store, err := NewStore(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := db.New(store.Pool).EventDedupeTruncate(ctx); err != nil {
		t.Fatal(err)
	}

	key := "chaos-concurrent:" + uuid.New().String()
	var firstWins int32
	const workers = 64
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			repo := NewDedupeRepo(store)
			seen, err := repo.SeenOrMark(ctx, key, uuid.New())
			if err != nil {
				t.Error(err)
				return
			}
			if !seen {
				atomic.AddInt32(&firstWins, 1)
			}
		}()
	}
	wg.Wait()
	if firstWins != 1 {
		t.Fatalf("expected exactly one successful first insert, got %d", firstWins)
	}
}
