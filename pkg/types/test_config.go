package types

import (
	"context"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	sdk "github.com/openshift-online/ocm-sdk-go"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// TestConfiguration
type TestConfiguration struct {
	TestID          string
	OutputDirectory string
	Duration        time.Duration
	Cooldown        time.Duration
	Rate            vegeta.Rate
	Connection      *sdk.Connection
	Logger          logging.Logger
	Ctx             context.Context
}
