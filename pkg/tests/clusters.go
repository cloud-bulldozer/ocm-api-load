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

func TestListClusters(attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {
	testName := "list-clusters"
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: http.MethodGet,
		URL:    helpers.ClustersEndpoint,
	})
	fileName := fmt.Sprintf("list-clusters-results-%s", time.Now().Local().Format("2006-01-02"))
	resFile, err := createFile(fileName, outputDirectory)
	if err != nil {
		return err
	}

	metrics[testName] = new(vegeta.Metrics)
	for res := range attacker.Attack(targeter, rate, duration, testName) {
		metrics[testName].Add(res)
		result.Write(res, resFile)
	}
	metrics[testName].Close()

	return report.Write("list-clusters-report",
		outputDirectory,
		metrics[testName])
}

func TestCreateCluster(attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {
	testName := "create-cluster"
	targeter := generateCreateClusterTargeter()
	fileName := fmt.Sprintf("create-cluster-results-%s", time.Now().Local().Format("2006-01-02"))
	resFile, err := createFile(fileName, outputDirectory)
	if err != nil {
		return err
	}

	metrics[testName] = new(vegeta.Metrics)
	for res := range attacker.Attack(targeter, rate, duration, testName) {
		metrics[testName].Add(res)
		result.Write(res, resFile)
	}
	metrics[testName].Close()

	return report.Write("create-cluster-report",
		outputDirectory,
		metrics[testName])
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
