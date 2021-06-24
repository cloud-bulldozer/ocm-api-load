package handlers

import (
	"bytes"
	"fmt"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestCreateCluster(options *types.TestOptions) error {

	testName := options.TestName
	targeter := generateCreateClusterTargeter(options.ID, options.Method, options.Path)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Encoder.Encode(res)
	}

	return nil
}

// Generates a targeter for the "POST /api/clusters_mgmt/v1/clusters" endpoint
// with monotonic increasing indexes.
// The clusters created are "fake clusters", that is, do not consume any cloud-provider infrastructure.
func generateCreateClusterTargeter(ID, method, url string) vegeta.Targeter {
	idx := 0

	// This will take the first 4 characters of the UUID
	// Cluster Names must match the following regex:
	// ^[a-z]([-a-z0-9]*[a-z0-9])?$
	id := ID[:4]

	// CCS is used to create fake clusters within the AWS
	// environment supplied by the user executing this test.
	// TODO: Fetch these values from the configuration file.
	ccsRegion := "us-east-1"
	ccsAccessKey := "example-access-key"
	ccsSecretKey := "example-private-key"
	ccsAccountID := "example-account-id"

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
