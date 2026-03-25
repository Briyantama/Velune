package sim

import (
	"slices"
	"testing"
)

func TestLoadFromEnv_ReproducibleWithSeed(t *testing.T) {
	t.Setenv("SIMULATE_SEED", "424242")
	t.Setenv("SIMULATE_PUBLISH_FAIL_RATE", "0.37")
	t.Setenv("SIMULATE_EMAIL_FAIL_RATE", "0.51")
	// Clear other sim vars
	t.Setenv("SIMULATE_BROKER_DOWN", "")
	t.Setenv("SIMULATE_CONSUMER_PANIC", "")

	seqPublish := func(c *Config) []bool {
		out := make([]bool, 48)
		for i := range out {
			out[i] = c.SimulatePublishFailure()
		}
		return out
	}
	seqEmail := func(c *Config) []bool {
		out := make([]bool, 48)
		for i := range out {
			out[i] = c.SimulateEmailFailure()
		}
		return out
	}

	a := LoadFromEnv()
	gotP := seqPublish(a)
	gotE := seqEmail(a)

	t.Setenv("SIMULATE_SEED", "424242")
	t.Setenv("SIMULATE_PUBLISH_FAIL_RATE", "0.37")
	t.Setenv("SIMULATE_EMAIL_FAIL_RATE", "0.51")
	b := LoadFromEnv()

	if !slices.Equal(gotP, seqPublish(b)) {
		t.Fatal("publish failure sequence must match for same seed and rate")
	}
	if !slices.Equal(gotE, seqEmail(b)) {
		t.Fatal("email failure sequence must match for same seed and rate")
	}
}

func TestLoadFromEnv_DifferentSeedDiffers(t *testing.T) {
	t.Setenv("SIMULATE_SEED", "1")
	t.Setenv("SIMULATE_PUBLISH_FAIL_RATE", "0.4")
	first := loadSeq(t)

	t.Setenv("SIMULATE_SEED", "2")
	second := loadSeq(t)

	equal := true
	for i := range first {
		if first[i] != second[i] {
			equal = false
			break
		}
	}
	if equal {
		t.Fatal("expected differing sequences for different seeds")
	}
}

func loadSeq(t *testing.T) []bool {
	t.Helper()
	c := LoadFromEnv()
	out := make([]bool, 32)
	for i := range out {
		out[i] = c.SimulatePublishFailure()
	}
	return out
}

func TestLoadFromEnv_BrokerDown(t *testing.T) {
	t.Setenv("SIMULATE_BROKER_DOWN", "true")
	c := LoadFromEnv()
	if !c.BrokerDown {
		t.Fatal("expected BrokerDown")
	}
}

func TestLoadFromEnv_RatesClamped(t *testing.T) {
	t.Setenv("SIMULATE_PUBLISH_FAIL_RATE", "99")
	c := LoadFromEnv()
	if c.PublishFailRate != 1 {
		t.Fatalf("got %v", c.PublishFailRate)
	}
}
