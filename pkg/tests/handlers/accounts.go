package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/report"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/result"
	sdk "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	uuid "github.com/satori/go.uuid"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// TestRegisterNewCluster performs a load test on the endpoint responsible for
// handling Registering New Clusters. This endpoint is typically used by Hive
// and not directly accessed by most clients.
func TestRegisterNewCluster(options *helpers.TestOptions) error {

	testName := options.TestName
	// Fetch the authorization token and create a dynamic Target generator for
	// building valid HTTP Requests
	targeter := generateClusterRegistrationTargeter(options.Path, options.Connection)

	// Create a file to store results
	fileName := fmt.Sprintf("%s_%s.json", options.ID, testName)
	resultFile, err := helpers.CreateFile(fileName, options.OutputDirectory)
	if err != nil {
		return err
	}
	defer resultFile.Close()

	// Store Metrics from load test
	options.Metrics[testName] = new(vegeta.Metrics)
	defer options.Metrics[testName].Close()

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		result.Write(res, resultFile)
		options.Metrics[testName].Add(res)
	}

	log.Printf("Results written to: %s", fileName)

	if options.WriteReport {
		err = report.Write(fmt.Sprintf("%s_%s-report", options.ID, options.TestName), options.OutputDirectory, options.Metrics[testName])
		if err != nil {
			return err
		}
	}

	return nil
}

// getAuthorizationToken will fetch and return the current user's Authorization
//Token which is required by certain endpoints such as Cluster Registration.
func getAuthorizationToken(connection *sdk.Connection) string {
	result, err := connection.AccountsMgmt().V1().AccessToken().Post().Send()
	if err != nil {
		log.Fatalf("Unable to retrieve authorization token: %s", err)
	}
	body := result.Body().Auths()
	token := body["cloud.openshift.com"].Auth()
	if len(token) == 0 {
		log.Println("Authorization token appears to be empty. Other requests may not succeed.")
	} else {
		log.Println("Successfully fetched Authorization Token")
	}
	return token
}

// generateClusterRegistrationTargeter returns a targeter which will create a
// unique Cluster Registration request body each time using a valid auth token
// and a UUID for the Cluster's ID to ensure uniqueness.
func generateClusterRegistrationTargeter(url string, connection *sdk.Connection) vegeta.Targeter {

	// Cache the Authorization Token to avoid retrieving it with every request
	var authorizationToken = ""
	if len(authorizationToken) == 0 {
		authorizationToken = getAuthorizationToken(connection)
	}

	targeter := func(t *vegeta.Target) error {

		// Each Cluster uses a UUID to ensure uniqueness
		clusterId := uuid.NewV4().String()
		body, err := v1.NewClusterRegistrationRequest().AuthorizationToken(authorizationToken).ClusterID(clusterId).Build()
		if err != nil {
			return err
		}

		var raw bytes.Buffer
		err = v1.MarshalClusterRegistrationRequest(body, &raw)
		if err != nil {
			return err
		}

		t.Method = http.MethodPost
		t.URL = url
		t.Body = raw.Bytes()

		return nil
	}

	return targeter
}

// Test quota cost
func TestQuotaCost(options *helpers.TestOptions) error {

	conn := options.Connection

	orgs, err := conn.AccountsMgmt().V1().Organizations().List().Send()
	if err != nil {
		return err
	}

	if orgs.Total() == 0 {
		return fmt.Errorf("no organizations where found for this account")
	}

	orgsIds := []string{}
	for _, o := range orgs.Items().Slice() {
		orgsIds = append(orgsIds, o.ID())
	}

	log.Printf("Using Organization id: %s.", orgsIds[0])
	options.Path = strings.Replace(options.Path, "{orgId}", orgsIds[0], 1)

	return TestStaticEndpoint(options)

}

// Test Cluster Authorizations
func TestClusterAuthorizations(options *helpers.TestOptions) error {
	testName := options.TestName
	// Vegeta Results File
	fileName := fmt.Sprintf("%s_%s.json", options.ID, testName)
	resultFile, err := helpers.CreateFile(fileName, options.OutputDirectory)
	if err != nil {
		return err
	}
	defer resultFile.Close()

	targeter := func(t *vegeta.Target) error {

		// Each Cluster uses a UUID to ensure uniqueness
		clusterId := uuid.NewV4().String()
		t.Method = http.MethodPost
		t.URL = options.Path
		t.Body = clusterAuthorizationsBody(clusterId)

		return nil
	}

	// Create a Metrics bucket for this test run
	options.Metrics[testName] = new(vegeta.Metrics)
	defer options.Metrics[testName].Close()

	// Execute the HTTP Requests; repeating as needed to meet the specified duration
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, options.TestName) {
		result.Write(res, resultFile)
		options.Metrics[testName].Add(res)
	}

	log.Printf("Results written to: %s", fileName)

	if options.WriteReport {
		err = report.Write(fmt.Sprintf("%s_%s-report", options.ID, options.TestName), options.OutputDirectory, options.Metrics[testName])
		if err != nil {
			return err
		}
	}

	return nil
}

func clusterAuthorizationsBody(clusterID string) []byte {
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
		log.Printf("building `cluster-authorizations` request: %s", err)
		return buff.Bytes()
	}
	err = v1.MarshalClusterAuthorizationRequest(clusterAuthReq, buff)
	if err != nil {

		log.Printf("marshaling `cluster-authorizations` request: %s", err)
	}
	return buff.Bytes()
}
