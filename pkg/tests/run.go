package tests

import (
	"time"

	uuid "github.com/satori/go.uuid"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type testCase func(attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error

func Run(
	attacker *vegeta.Attacker,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {

	// testId is an identifier used to associate all tests in this test suite with each
	// other
	testID := uuid.NewV4().String()

	testCases := []testCase{
		TestCreateCluster,
		TestListClusters,
		TestSelfAccessToken,
		TestListSubscriptions,
		TestAccessReview,
		TestRegisterNewCluster,
	}

	for _, testCase := range testCases {
		err := testCase(attacker,
			testID,
			metrics,
			rate,
			outputDirectory,
			duration)
		if err != nil {
			return err
		}
	}
	return nil
}
