package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	v1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	uuid "github.com/satori/go.uuid"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// TestRegisterNewCluster performs a load test on the endpoint responsible for
// handling Registering New Clusters. This endpoint is typically used by Hive
// and not directly accessed by most clients.
func TestRegisterNewCluster(ctx context.Context, options *types.TestOptions) error {

	testName := options.TestName
	// Fetch the authorization token and create a dynamic Target generator for
	// building valid HTTP Requests
	targeter := generateClusterRegistrationTargeter(ctx, options)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Encoder.Encode(res)
	}

	return nil
}

func TestRegisterExistingCluster(ctx context.Context, options *types.TestOptions) error {

	testName := options.TestName
	quantity := options.Rate.Freq
	targeter := generateClusterReRegistrationTargeter(ctx, quantity, options)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Encoder.Encode(res)
	}

	return nil
}

// getAuthorizationToken will fetch and return the current user's Authorization
//Token which is required by certain endpoints such as Cluster Registration.
func getAuthorizationToken(ctx context.Context, options *types.TestOptions) string {
	result, err := options.Connection.AccountsMgmt().V1().AccessToken().Post().Send()
	if err != nil {
		options.Logger.Error(ctx, "Unable to retrieve authorization token: %s", err)
	}
	body := result.Body().Auths()
	token := body["cloud.openshift.com"].Auth()
	if len(token) == 0 {
		options.Logger.Warn(ctx, "Authorization token appears to be empty. Other requests may not succeed.")
	} else {
		options.Logger.Info(ctx, "Successfully fetched Authorization Token")
	}
	return token
}

// generateClusterReRegistrationTargeter registers fake clusters and then
// returns a targeter which uses those fake clusters.
func generateClusterReRegistrationTargeter(ctx context.Context, qty int, options *types.TestOptions) vegeta.Targeter {

	clusterIds := make([]string, qty)
	var currentTarget = 0

	// Cache the Authorization Token to avoid retrieving it with every request
	var authorizationToken = ""
	if len(authorizationToken) == 0 {
		authorizationToken = getAuthorizationToken(ctx, options)
	}

	// Register multiple mock clusters and store their IDs
	options.Logger.Info(ctx, "Registering %d clusters to use for re-registration test", qty)
	for i := range clusterIds {

		clusterID := uuid.NewV4().String()

		body, err := v1.NewClusterRegistrationRequest().AuthorizationToken(authorizationToken).ClusterID(clusterID).Build()
		if err != nil {
			options.Logger.Fatal(ctx, "Unable to build cluster registration request: %v", err)
		}

		var rawBody bytes.Buffer
		err = v1.MarshalClusterRegistrationRequest(body, &rawBody)
		if err != nil {
			options.Logger.Fatal(ctx, "Unable to serialize cluster registration request body: ", err)
		}

		resp, err := options.Connection.AccountsMgmt().V1().ClusterRegistrations().Post().Request(body).Send()
		if err != nil {
			options.Logger.Fatal(ctx, "Unable to register cluster: ", err)
		}

		options.Logger.Info(ctx, "[%d/%d] Registered Cluster: '%s'. Response: %d\n", i, len(clusterIds), clusterID, resp.Status())
		clusterIds[i] = clusterID

		// Avoid hitting rate limiting
		time.Sleep(time.Second * 1)

	}

	targeter := func(t *vegeta.Target) error {

		clusterId := clusterIds[currentTarget]

		body, err := v1.NewClusterRegistrationRequest().AuthorizationToken(authorizationToken).ClusterID(clusterId).Build()
		if err != nil {
			return err
		}

		var rawBody bytes.Buffer
		err = v1.MarshalClusterRegistrationRequest(body, &rawBody)
		if err != nil {
			return err
		}

		t.Method = http.MethodPost
		t.URL = options.Path
		t.Body = rawBody.Bytes()

		// Loop through Cluster IDs
		currentTarget = currentTarget + 1
		if currentTarget > (qty - 1) {
			currentTarget = 0
		}

		return nil
	}

	return targeter

}

// generateClusterRegistrationTargeter returns a targeter which will create a
// unique Cluster Registration request body each time using a valid auth token
// and a UUID for the Cluster's ID to ensure uniqueness.
func generateClusterRegistrationTargeter(ctx context.Context, options *types.TestOptions) vegeta.Targeter {

	// Cache the Authorization Token to avoid retrieving it with every request
	var authorizationToken = ""
	if len(authorizationToken) == 0 {
		authorizationToken = getAuthorizationToken(ctx, options)
	}

	targeter := func(t *vegeta.Target) error {

		// Each Cluster uses a UUID to ensure uniqueness
		clusterId := uuid.NewV4().String()
		body, err := v1.NewClusterRegistrationRequest().AuthorizationToken(authorizationToken).ClusterID(clusterId).Build()
		if err != nil {
			return err
		}

		var rawBody bytes.Buffer
		err = v1.MarshalClusterRegistrationRequest(body, &rawBody)
		if err != nil {
			return err
		}

		t.Method = http.MethodPost
		t.URL = options.Path
		t.Body = rawBody.Bytes()

		return nil
	}

	return targeter
}

// Test quota cost
func TestQuotaCost(ctx context.Context, options *types.TestOptions) error {

	conn := options.Connection

	acct, err := conn.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		return err
	}

	org, ok := acct.Body().GetOrganization()
	if !ok {
		return fmt.Errorf("no organizations where found for this account")
	}

	orgID, ok := org.GetID()
	if !ok {
		return fmt.Errorf("no organizations where found for this account")
	}

	options.Logger.Info(ctx, "Using Organization id: %s.", orgID)
	options.Path = strings.Replace(options.Path, "{orgId}", orgID, 1)

	return TestStaticEndpoint(ctx, options)

}

// Test Cluster Authorizations
func TestClusterAuthorizations(ctx context.Context, options *types.TestOptions) error {

	targeter := func(t *vegeta.Target) error {

		// Each Cluster uses a UUID to ensure uniqueness
		clusterId := uuid.NewV4().String()
		t.Method = http.MethodPost
		t.URL = options.Path
		t.Body = clusterAuthorizationsBody(ctx, clusterId, options)

		return nil
	}

	// Execute the HTTP Requests; repeating as needed to meet the specified duration
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, options.TestName) {
		options.Encoder.Encode(res)
	}

	return nil
}

func clusterAuthorizationsBody(ctx context.Context, clusterID string, options *types.TestOptions) []byte {
	buff := &bytes.Buffer{}
	reservedResource := v1.NewReservedResource().
		ResourceName(helpers.M5XLargeResource).
		ResourceType(helpers.AWSComputeNodeResourceType).
		Count(helpers.DefaultClusterCount).
		BillingModel(helpers.StandardBillingModel)

	clusterAuthReq, err := v1.NewClusterAuthorizationRequest().
		ClusterID(clusterID).
		ProductID(helpers.OsdProductID).
		CloudProviderID(helpers.AWSCloudProvider).
		AccountUsername(helpers.ClusterAuthAccountUsername).
		Managed(helpers.ClusterAuthManaged).
		Reserve(helpers.ClusterAuthReserve).
		BYOC(helpers.ClusterAuthBYOC).
		AvailabilityZone(helpers.SingleAvailabilityZone).
		Resources(reservedResource).
		Build()
	if err != nil {
		options.Logger.Info(ctx, "building `cluster-authorizations` request: %s", err)
		return buff.Bytes()
	}
	err = v1.MarshalClusterAuthorizationRequest(clusterAuthReq, buff)
	if err != nil {

		options.Logger.Error(ctx, "marshaling `cluster-authorizations` request: %s", err)
	}
	return buff.Bytes()
}
