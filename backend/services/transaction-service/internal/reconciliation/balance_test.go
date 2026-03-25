package reconciliation

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/shared/contracts"
)

func TestBalanceMismatchDetected_JSON(t *testing.T) {
	id := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	uid := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	p := contracts.BalanceMismatchDetected{
		AccountID:          id,
		UserID:             uid,
		StoredBalanceMinor: 100,
		LedgerSumMinor:     99,
		Currency:           "USD",
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var out contracts.BalanceMismatchDetected
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.LedgerSumMinor != 99 || out.StoredBalanceMinor != 100 {
		t.Fatalf("%+v", out)
	}
}
