package tests

import (
	"net/http"
	"time"

	"github.com/nimrodshn/cs-load-test/pkg/helpers"
	sdk "github.com/openshift-online/ocm-sdk-go"
	uuid "github.com/satori/go.uuid"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func Run(
	attacker *vegeta.Attacker,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration,
	connection *sdk.Connection) error {

	// testId provides a common value to associate all output data from running
	// the full test suite with a single test run.
	testID := uuid.NewV4().String()

	// Specify Test Cases
	// They are written this way to re-use functionality where possible and
	// hopefully make it easier to modify and/or extend given the declarative
	// style.
	tests := []helpers.TestOptions{

		{
			TestName: "self-access-token",
			Path:     helpers.SelfAccessTokenEndpoint,
			Method:   http.MethodPost,
			Rate:     helpers.SelfAccessTokenRate,
			Handler:  TestGenericEndpoint,
		},

		{
			TestName: "list-subscriptions",
			Path:     helpers.ListSubscriptionsEndpoint,
			Method:   http.MethodGet,
			Rate:     helpers.ListSubscriptionsRate,
			Handler:  TestGenericEndpoint,
		},

		{
			TestName: "access-review",
			Path:     helpers.AccessReviewEndpoint,
			Method:   http.MethodPost,
			Body:     "{\"account_username\": \"rhn-support-tiwillia\", \"action\": \"get\", \"resource_type\": \"Subscription\"}",
			Rate:     helpers.AccessReviewRate,
			Handler:  TestGenericEndpoint,
		},

		{
			TestName: "register-new-cluster",
			Path:     helpers.ClusterRegistrationEndpoint,
			Method:   http.MethodPost,
			Rate:     helpers.RegisterNewClusterRate,
			Handler:  TestRegisterNewCluster,
		},

		{
			TestName: "create-cluster",
			Rate:     helpers.CreateClusterRate,
			Handler:  TestCreateCluster,
		},

		{
			TestName: "list-clusters",
			Rate:     helpers.ListClustersRate,
			Handler:  TestListClusters,
		},
		{
			TestName: "get-current-account",
			Path:     helpers.GetCurrentAccountEndpoint,
			Method:   http.MethodGet,
			Rate:     helpers.GetCurrentAccount,
			Handler:  TestGenericEndpoint,
		},
	}

	for i := range tests {

		// Bind "Test Harness"
		tests[i].ID = testID
		tests[i].Duration = duration
		tests[i].OutputDirectory = outputDirectory
		tests[i].Attacker = attacker
		tests[i].Metrics = metrics
		tests[i].Connection = connection

		// Execute the Test
		test := tests[i]
		err := test.Handler(&test)
		if err != nil {
			return err
		}

	}

	return nil
}
