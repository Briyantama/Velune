package delivery

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type MemoryDedupe struct {
	mu   sync.Mutex
	keys map[string]struct{}
}

func NewMemoryDedupe() *MemoryDedupe {
	return &MemoryDedupe{
		keys: make(map[string]struct{}),
	}
}

func (m *MemoryDedupe) SeenOrMark(_ context.Context, key string, _ uuid.UUID) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.keys[key]
	if ok {
		return true, nil
	}
	m.keys[key] = struct{}{}
	return false, nil
}
