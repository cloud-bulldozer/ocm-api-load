package handlers

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/report"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/result"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestCreateCluster(options *helpers.TestOptions) error {

	testName := options.TestName
	targeter := generateCreateClusterTargeter(options.ID, options.Method, options.Path)
	fileName := fmt.Sprintf("%s_%s-%s.json", options.ID, options.TestName, time.Now().Local().Format("2006-01-02"))
	resFile, err := helpers.CreateFile(fileName, options.OutputDirectory)
	if err != nil {
		return err
	}

	options.Metrics[testName] = new(vegeta.Metrics)
	defer options.Metrics[testName].Close()
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Metrics[testName].Add(res)
		result.Write(res, resFile)
	}

	log.Printf("Results written to: %s", fileName)

	if options.WriteReport {
		err = report.Write(fmt.Sprintf("%s_%s-report", options.ID, options.TestName), options.ReportsDirectory, options.Metrics[testName])
		if err != nil {
			return err
		}
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

	targeter := func(t *vegeta.Target) error {
		fakeClusterProps := map[string]string{
			"fake_cluster": "true",
		}
		body, err := v1.NewCluster().
			Name(fmt.Sprintf("perf-%s-%d", id, idx)).
			Properties(fakeClusterProps).
			MultiAZ(true).
			Region(v1.NewCloudRegion().ID(helpers.DefaultAWSRegion)).
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
