package helpers

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var (
	// Default Rate
	DefaultRate = vegeta.Rate{Freq: 5, Per: time.Second}

	// Authorization Services
	RegisterExistingClusterRate = vegeta.Rate{Freq: 25, Per: time.Second}
)
