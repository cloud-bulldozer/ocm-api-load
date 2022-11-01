package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	v1 "github.com/openshift-online/ocm-sdk-go/servicemgmt/v1"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestCreateService(ctx context.Context, options *types.TestOptions) error {
	testName := options.TestName
	targeter := generateCreateServiceTargeter(ctx, options.ID, options.Method, options.Path, options.Logger)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Encoder.Encode(res)
	}

	helpers.Cleanup(ctx, options.Connection)
	return nil
}

// Generates a targeter for the "POST /api/service_mgmt/v1/services" endpoint
// with monotonic increasing indexes.
// The clusters created are "fake clusters", that is, do not consume any cloud-provider infrastructure.
func generateCreateServiceTargeter(ctx context.Context, ID, method, url string, log logging.Logger) vegeta.Targeter {
	idx := 0

	// This will take the first 4 characters of the UUID
	// Cluster Names must match the following regex:
	// ^[a-z]([-a-z0-9]*[a-z0-9])?$
	id := ID[:4]

	awsCreds := viper.Get("aws").([]interface{})
	if len(awsCreds) < 1 {
		log.Fatal(ctx, "No aws credentials found")
	}

	// CCS is used to create fake clusters within the AWS
	// environment supplied by the user executing this test.
	// Not fully supporting multi account now, so using first accaunt always
	ccsRegion := awsCreds[0].(map[string]interface{})["region"].(string)
	ccsAccessKey := awsCreds[0].(map[string]interface{})["access-key"].(string)
	ccsSecretKey := awsCreds[0].(map[string]interface{})["secret-access-key"].(string)
	ccsAccountID := awsCreds[0].(map[string]interface{})["account-id"].(string)
	ccsAccountName := awsCreds[0].(map[string]interface{})["account-name"].(string)

	targeter := func(t *vegeta.Target) error {
		fakeClusterProps := map[string]string{
			"fake_cluster": "true",
		}
		arn := strings.Replace("arn:aws:iam::{acctID}:user/{acctName}", "{acctID}", ccsAccountID, -1)
		arn = strings.Replace(arn, "{acctName}", ccsAccountName, -1)
		rosaCreatorProps := map[string]string{
			"rosa_creator_arn": arn,
		}

		body, err := v1.NewManagedService().
			Service("ocm-addon-test-operator").
			Parameters(v1.NewServiceParameter().ID("has-external-resources").Value("false")).
			Cluster(v1.NewCluster().
				Name(fmt.Sprintf("perf-%s-%d", id, idx)).
				AWS(
					v1.NewAWS().
						AccessKeyID(ccsAccessKey).
						SecretAccessKey(ccsSecretKey).
						AccountID(ccsAccountID),
				).
				Nodes(v1.NewClusterNodes().AvailabilityZones(fmt.Sprintf("%sa", ccsRegion))).
				Properties(fakeClusterProps).
				Properties(rosaCreatorProps).
				Region(v1.NewCloudRegion().ID(ccsRegion))).
			Build()

		if err != nil {
			return err
		}

		var raw bytes.Buffer
		err = v1.MarshalManagedService(body, &raw)
		if err != nil {
			return err
		}
		t.Method = method
		t.URL = url
		t.Body = raw.Bytes()

		idx += 1
		return nil
	}
	return targeter
}

func TestPatchService(ctx context.Context, options *types.TestOptions) error {
	idx := 0

	// This will take the first 4 characters of the UUID
	// Cluster Names must match the following regex:
	// ^[a-z]([-a-z0-9]*[a-z0-9])?$
	id := options.ID[:4]

	awsCreds := viper.Get("aws").([]interface{})
	if len(awsCreds) < 1 {
		log.Fatal(ctx, "No aws credentials found")
	}

	// CCS is used to create fake clusters within the AWS
	// environment supplied by the user executing this test.
	// Not fully supporting multi account now, so using first accaunt always
	ccsRegion := awsCreds[0].(map[string]interface{})["region"].(string)
	ccsAccessKey := awsCreds[0].(map[string]interface{})["access-key"].(string)
	ccsSecretKey := awsCreds[0].(map[string]interface{})["secret-access-key"].(string)
	ccsAccountID := awsCreds[0].(map[string]interface{})["account-id"].(string)
	ccsAccountName := awsCreds[0].(map[string]interface{})["account-name"].(string)
	serviceIds := make([]string, 2)

	fakeClusterProps := map[string]string{
		"fake_cluster": "true",
	}
	arn := strings.Replace("arn:aws:iam::{acctID}:user/{acctName}", "{acctID}", ccsAccountID, -1)
	arn = strings.Replace(arn, "{acctName}", ccsAccountName, -1)
	rosaCreatorProps := map[string]string{
		"rosa_creator_arn": arn,
	}

	// Register multiple mock Services and store their IDs
	options.Logger.Info(ctx, "Registering 2 Services to use for patch requests test")
	for i := range serviceIds {

		clusterID := uuid.NewV4().String()

		body, err := v1.NewManagedService().
			Service("ocm-addon-test-operator").
			Parameters(v1.NewServiceParameter().ID("has-external-resources").Value("false")).
			Cluster(v1.NewCluster().
				Name(fmt.Sprintf("perf-%s-%d", id, idx)).
				AWS(
					v1.NewAWS().
						AccessKeyID(ccsAccessKey).
						SecretAccessKey(ccsSecretKey).
						AccountID(ccsAccountID),
				).
				Nodes(v1.NewClusterNodes().AvailabilityZones(ccsRegion)).
				Properties(fakeClusterProps).
				Properties(rosaCreatorProps).
				Region(v1.NewCloudRegion().ID(ccsRegion))).
			Build()
		if err != nil {
			options.Logger.Fatal(ctx, "Unable to build Service request: %v", err)
		}

		var rawBody bytes.Buffer
		err = v1.MarshalManagedService(body, &rawBody)
		if err != nil {
			options.Logger.Fatal(ctx, "Unable to serialize Service request body: ", err)
		}

		resp, err := options.Connection.ServiceMgmt().V1().Services().Add().Body(body).Send()
		if err != nil {
			options.Logger.Fatal(ctx, "Unable to create Service: ", err)
		}
		serviceID, ok := resp.Body().GetID()
		if !ok {
			options.Logger.Info(ctx, "Unable to get Service ID")
		}

		options.Logger.Info(ctx, "[%d/%d] Created Service: '%s'. Response: %d\n", i, len(serviceIds), serviceID, resp.Status())
		serviceIds[i] = clusterID

		// Avoid hitting rate limiting
		time.Sleep(time.Second * 1)
	}

	testName := options.TestName
	targeter := generatePatchServiceTargeter(ctx, options.ID, options.Method, options.Path, options.Logger, serviceIds)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Encoder.Encode(res)
	}

	helpers.Cleanup(ctx, options.Connection)
	return nil
}

func generatePatchServiceTargeter(ctx context.Context, ID, method, url string, log logging.Logger, ids []string) vegeta.Targeter {
	idx := 0
	var currentTarget = 0

	// This will take the first 4 characters of the UUID
	// Cluster Names must match the following regex:
	// ^[a-z]([-a-z0-9]*[a-z0-9])?$
	id := ID[:4]

	awsCreds := viper.Get("aws").([]interface{})
	if len(awsCreds) < 1 {
		log.Fatal(ctx, "No aws credentials found")
	}

	// CCS is used to create fake clusters within the AWS
	// environment supplied by the user executing this test.
	// Not fully supporting multi account now, so using first accaunt always
	ccsRegion := awsCreds[0].(map[string]interface{})["region"].(string)
	ccsAccessKey := awsCreds[0].(map[string]interface{})["access-key"].(string)
	ccsSecretKey := awsCreds[0].(map[string]interface{})["secret-access-key"].(string)
	ccsAccountID := awsCreds[0].(map[string]interface{})["account-id"].(string)
	ccsAccountName := awsCreds[0].(map[string]interface{})["account-name"].(string)

	targeter := func(t *vegeta.Target) error {
		fakeClusterProps := map[string]string{
			"fake_cluster": "true",
		}
		arn := strings.Replace("arn:aws:iam::{acctID}:user/{acctName}", "{acctID}", ccsAccountID, -1)
		arn = strings.Replace(arn, "{acctName}", ccsAccountName, -1)
		rosaCreatorProps := map[string]string{
			"rosa_creator_arn": arn,
		}
		updateProps := map[string]string{
			"random_prop": fmt.Sprintf("prop-%d", idx),
		}

		body, err := v1.NewManagedService().
			Service("ocm-addon-test-operator").
			Parameters(v1.NewServiceParameter().ID("has-external-resources").Value("false")).
			Cluster(v1.NewCluster().
				Name(fmt.Sprintf("perf-%s-%d", id, idx)).
				AWS(
					v1.NewAWS().
						AccessKeyID(ccsAccessKey).
						SecretAccessKey(ccsSecretKey).
						AccountID(ccsAccountID),
				).
				Nodes(v1.NewClusterNodes().AvailabilityZones(ccsRegion)).
				Properties(fakeClusterProps).
				Properties(rosaCreatorProps).
				Properties(updateProps).
				Region(v1.NewCloudRegion().ID(ccsRegion))).
			Build()

		if err != nil {
			return err
		}

		var raw bytes.Buffer
		err = v1.MarshalManagedService(body, &raw)
		if err != nil {
			return err
		}
		t.Method = method
		t.URL = strings.Replace(url, "{srvcId}", ids[currentTarget], -1)
		t.Body = raw.Bytes()

		idx += 1
		if idx >= len(ids) {
			idx = 0
		}
		return nil
	}
	return targeter
}