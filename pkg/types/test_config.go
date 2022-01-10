package types

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// TestConfiguration
type TestConfiguration struct {
	Duration time.Duration
	Cooldown time.Duration
	Rate     vegeta.Rate
}
