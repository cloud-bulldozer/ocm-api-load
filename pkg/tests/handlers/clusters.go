package handlers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	"github.com/spf13/viper"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestCreateCluster(ctx context.Context, options *types.TestOptions) error {

	testName := options.TestName
	targeter := generateCreateClusterTargeter(ctx, options.ID, options.Method, options.Path, options.Logger)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Encoder.Encode(res)
	}

	return nil
}

// Generates a targeter for the "POST /api/clusters_mgmt/v1/clusters" endpoint
// with monotonic increasing indexes.
// The clusters created are "fake clusters", that is, do not consume any cloud-provider infrastructure.
func generateCreateClusterTargeter(ctx context.Context, ID, method, url string, log logging.Logger) vegeta.Targeter {
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
	ccsRegion := awsCreds[0].(map[interface{}]interface{})["region"].(string)
	ccsAccessKey := awsCreds[0].(map[interface{}]interface{})["access-key"].(string)
	ccsSecretKey := awsCreds[0].(map[interface{}]interface{})["secret-access-key"].(string)
	ccsAccountID := awsCreds[0].(map[interface{}]interface{})["account-id"].(string)

	targeter := func(t *vegeta.Target) error {
		fakeClusterProps := map[string]string{
			"fake_cluster": "true",
		}
		body, err := v1.NewCluster().
			Name(fmt.Sprintf("perf-%s-%d", id, idx)).
			Properties(fakeClusterProps).
			MultiAZ(true).
			Region(v1.NewCloudRegion().ID(ccsRegion)).
			CCS(v1.NewCCS().Enabled(true)).
			AWS(
				v1.NewAWS().
					AccessKeyID(ccsAccessKey).
					SecretAccessKey(ccsSecretKey).
					AccountID(ccsAccountID),
			).
			Build()
		if err != nil {
			return err
		}

		var raw bytes.Buffer
		err = v1.MarshalCluster(body, &raw)
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
