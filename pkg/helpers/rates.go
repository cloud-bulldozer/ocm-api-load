package helpers

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var (

	// Cluster Services
	CreateClusterRate = vegeta.Rate{Freq: 10, Per: time.Second}
	ListClustersRate  = vegeta.Rate{Freq: 10, Per: time.Second}

	// Account Services
	SelfAccessTokenRate   = vegeta.Rate{Freq: 1000, Per: time.Hour}
	ListSubscriptionsRate = vegeta.Rate{Freq: 2000, Per: time.Hour}

	// Authorization Services
	AccessReviewRate            = vegeta.Rate{Freq: 100, Per: time.Second}
	RegisterNewClusterRate      = vegeta.Rate{Freq: 1000, Per: time.Hour}
	RegisterExistingClusterRate = vegeta.Rate{Freq: 25, Per: time.Second}
)
