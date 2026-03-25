package sim

import "errors"

// ErrSimulatedPublishFailure is returned when SIMULATE_PUBLISH_FAIL_RATE triggers (main exchange only).
var ErrSimulatedPublishFailure = errors.New("simulated publish failure")
