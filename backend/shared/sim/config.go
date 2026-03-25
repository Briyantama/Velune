package sim

import (
	"os"

	"github.com/moon-eye/velune/shared/helper"
	"github.com/moon-eye/velune/shared/stringx"
)

// Config holds env-driven chaos simulation flags (all off by default).
type Config struct {
	BrokerDown      bool
	PublishFailRate float64
	EmailFailRate   float64
	ConsumerPanic   bool
	DLQSnoop        bool
	rngSource       int64
	randFloat       func() float64
}

// LoadFromEnv reads SIMULATE_* variables once. Safe to call from service main.
func LoadFromEnv() *Config {
	c := &Config{
		BrokerDown:      stringx.StringsEqualTrue(os.Getenv("SIMULATE_BROKER_DOWN")),
		ConsumerPanic:   stringx.StringsEqualTrue(stringx.TrimSpace(os.Getenv("SIMULATE_CONSUMER_PANIC"))),
		DLQSnoop:        stringx.StringsEqualTrue(stringx.TrimSpace(os.Getenv("SIMULATE_DLQ_SNOOP"))),
		PublishFailRate: helper.ToFloat64(os.Getenv("SIMULATE_PUBLISH_FAIL_RATE")),
		EmailFailRate:   helper.ToFloat64(os.Getenv("SIMULATE_EMAIL_FAIL_RATE")),
	}
	if c.PublishFailRate < 0 {
		c.PublishFailRate = 0
	}
	if c.PublishFailRate > 1 {
		c.PublishFailRate = 1
	}
	if c.EmailFailRate < 0 {
		c.EmailFailRate = 0
	}
	if c.EmailFailRate > 1 {
		c.EmailFailRate = 1
	}

	seed, err := helper.ToInt64(os.Getenv("SIMULATE_SEED"))
	if err != nil {
		c.rngSource = 0
	} else {
		c.rngSource = seed
	}
	src := c.rngSource
	// Mulberry64: deterministic, fast, good enough for test injection.
	var state uint64
	if src != 0 {
		state = uint64(src)
	} else {
		state = 0x123456789abcdef
	}
	c.randFloat = func() float64 {
		state += 0x9e3779b97f4a7c15
		z := state
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		z = z ^ (z >> 31)
		// [0,1)
		return float64(z>>11) / float64(1<<53)
	}
	return c
}

// RNGSeed returns the effective seed (after env or crypto fallback) for logging/tests.
func (c *Config) RNGSeed() int64 {
	if c == nil {
		return 0
	}
	return c.rngSource
}

// SimulatePublishFailure returns true with probability PublishFailRate.
func (c *Config) SimulatePublishFailure() bool {
	if c == nil || c.PublishFailRate <= 0 {
		return false
	}
	return c.randFloat() < c.PublishFailRate
}

// SimulateEmailFailure returns true with probability EmailFailRate.
func (c *Config) SimulateEmailFailure() bool {
	if c == nil || c.EmailFailRate <= 0 {
		return false
	}
	return c.randFloat() < c.EmailFailRate
}