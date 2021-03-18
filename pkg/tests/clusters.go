package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nimrodshn/cs-load-test/pkg/helpers"
	"github.com/nimrodshn/cs-load-test/pkg/report"
	"github.com/nimrodshn/cs-load-test/pkg/result"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	defaultAWSRegion = "us-east-1"
)

func TestListClusters(options *helpers.TestOptions) error {
	testName := options.TestName
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodGet,
		URL:    helpers.ClustersEndpoint,
	})
	fileName := fmt.Sprintf("list-clusters-results-%s", time.Now().Local().Format("2006-01-02"))
	resFile, err := createFile(fileName, options.OutputDirectory)
	if err != nil {
		return err
	}

	options.Metrics[testName] = new(vegeta.Metrics)
	defer options.Metrics[testName].Close()
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Metrics[testName].Add(res)
		result.Write(res, resFile)
	}

	err = report.Write("list-clusters-report", options.OutputDirectory, options.Metrics[testName])
	if err != nil {
		return err
	}

	return nil
}

func TestCreateCluster(options *helpers.TestOptions) error {

	testName := options.TestName
	targeter := generateCreateClusterTargeter()
	// TODO: Consistent filename with other test results
	fileName := fmt.Sprintf("create-cluster-results-%s", time.Now().Local().Format("2006-01-02"))
	resFile, err := createFile(fileName, options.OutputDirectory)
	if err != nil {
		return err
	}

	options.Metrics[testName] = new(vegeta.Metrics)
	defer options.Metrics[testName].Close()
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Metrics[testName].Add(res)
		result.Write(res, resFile)
	}

	// TODO: Consistency among all tests
	err = report.Write("create-cluster-report", options.OutputDirectory, options.Metrics[testName])
	if err != nil {
		return err
	}

	return nil
}

// Generates a targeter for the "POST /api/clusters_mgmt/v1/clusters" endpoint
// with monotonic increasing indexes.
// The clusters created are "fake clusters", that is, do not consume any cloud-provider infrastructure.
func generateCreateClusterTargeter() vegeta.Targeter {
	idx := 0

	targeter := func(t *vegeta.Target) error {
		fakeClusterProps := map[string]string{
			"fake_cluster": "true",
		}
		body, err := v1.NewCluster().
			Name(fmt.Sprintf("test-cluster-%d", idx)).
			Properties(fakeClusterProps).
			MultiAZ(false).
			Region(v1.NewCloudRegion().ID(defaultAWSRegion)).
			Build()
		if err != nil {
			return err
		}

		var raw bytes.Buffer
		err = v1.MarshalCluster(body, &raw)
		if err != nil {
			return err
		}

		t.Method = http.MethodPost
		t.URL = helpers.ClustersEndpoint
		t.Body = raw.Bytes()

		idx += 1
		return nil
	}
	return targeter
}

func createFile(name, path string) (*os.File, error) {
	resultPath := filepath.Join(path, fmt.Sprintf("%s.json", name))
	out, err := os.Create(resultPath)
	if err != nil {
		// Silently ignore pre-existing file.
		if err == os.ErrExist {
			return out, nil
		}
		return nil, fmt.Errorf("Error while writing result: %v", err)
	}
	return out, nil
}
