package sim

import (
	"os"
	"strconv"

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
		PublishFailRate: parsePercentEnvFloat("SIMULATE_PUBLISH_FAIL_RATE"),
		EmailFailRate:   parsePercentEnvFloat("SIMULATE_EMAIL_FAIL_RATE"),
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

	c.rngSource = parseEnvInt64("SIMULATE_SEED")
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

func parseEnvInt64(key string) int64 {
	v := stringx.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// parsePercentEnvFloat parses a float env var and returns 0 when missing/invalid.
// Values are expected to already be in [0,1] (e.g. 0.37).
func parsePercentEnvFloat(key string) float64 {
	v := stringx.TrimSpace(os.Getenv(key))
	if v == "" {
		return 0
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
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
