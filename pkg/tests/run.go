package tests

import (
	"net/http"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/tests/handlers"
	sdk "github.com/openshift-online/ocm-sdk-go"
	uuid "github.com/satori/go.uuid"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func Run(
	attacker *vegeta.Attacker,
	metrics map[string]*vegeta.Metrics,
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
			Handler:  handlers.TestStaticEndpoint,
		},
		{
			TestName: "list-subscriptions",
			Path:     helpers.ListSubscriptionsEndpoint,
			Method:   http.MethodGet,
			Rate:     helpers.ListSubscriptionsRate,
			Handler:  handlers.TestStaticEndpoint,
		},
		{
			TestName: "access-review",
			Path:     helpers.AccessReviewEndpoint,
			Method:   http.MethodPost,
			Body:     "{\"account_username\": \"rhn-support-tiwillia\", \"action\": \"get\", \"resource_type\": \"Subscription\"}",
			Rate:     helpers.AccessReviewRate,
			Handler:  handlers.TestStaticEndpoint,
		},
		{
			TestName: "register-new-cluster",
			Path:     helpers.ClusterRegistrationEndpoint,
			Method:   http.MethodPost,
			Rate:     helpers.RegisterNewClusterRate,
			Handler:  handlers.TestRegisterNewCluster,
		},
		{
			TestName: "create-cluster",
			Rate:     helpers.CreateClusterRate,
			Handler:  handlers.TestCreateCluster,
		},
		{
			TestName: "list-clusters",
			Rate:     helpers.ListClustersRate,
			Handler:  handlers.TestListClusters,
		},
		{
			TestName: "get-current-account",
			Path:     helpers.GetCurrentAccountEndpoint,
			Method:   http.MethodGet,
			Rate:     helpers.GetCurrentAccountRate,
			Handler:  handlers.TestStaticEndpoint,
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
