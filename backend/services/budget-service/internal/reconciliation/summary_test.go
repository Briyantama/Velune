package reconciliation

import "testing"

func TestAlertImpliedSpent(t *testing.T) {
	if g := alertImpliedSpent(10_000, 50); g != 5_000 {
		t.Fatalf("got %d", g)
	}
	if g := alertImpliedSpent(0, 99); g != 0 {
		t.Fatalf("got %d", g)
	}
}

func TestSpentMinorExceedsTolerance(t *testing.T) {
	if spentMinorExceedsTolerance(100, 100) {
		t.Fatal("equal should not exceed")
	}
	if spentMinorExceedsTolerance(100, 101) {
		t.Fatal("diff 1 should be within tolerance")
	}
	if !spentMinorExceedsTolerance(100, 102) {
		t.Fatal("diff 2 should exceed")
	}
}
