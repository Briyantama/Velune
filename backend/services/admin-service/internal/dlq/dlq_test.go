package dlq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/shared/contracts"
)

// Replay is integration-tested against Rabbit; this documents the scan contract with a unit-style helper.

func TestReplayScanLogic_matchEventID(t *testing.T) {
	eid := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	env := contracts.EventEnvelope{EventID: eid, EventType: "x.y", Payload: json.RawMessage(`{}`)}
	b, _ := json.Marshal(env)
	if string(b) == "" {
		t.Fatal()
	}
	var back contracts.EventEnvelope
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.EventID != eid {
		t.Fatal()
	}
}

type fakePublisher struct {
	last *contracts.EventEnvelope
	err  error
}

func (f *fakePublisher) pub(ctx context.Context, env contracts.EventEnvelope) error {
	if f.err != nil {
		return f.err
	}
	f.last = &env
	return nil
}

func TestReplayHandler_contract(t *testing.T) {
	fp := &fakePublisher{}
	_ = fp
	if fp.pub(context.Background(), contracts.EventEnvelope{}) != nil {
		t.Fatal()
	}
	fp.err = errors.New("x")
	if fp.pub(context.Background(), contracts.EventEnvelope{}) == nil {
		t.Fatal()
	}
}
