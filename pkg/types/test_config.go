package types

import (
	"context"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// TestConfiguration
type TestConfiguration struct {
	TestID          string
	OutputDirectory string
	Duration        time.Duration
	Cooldown        time.Duration
	Rate            vegeta.Rate
	Logger          logging.Logger
	Ctx             context.Context
}
