package handlers

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	"github.com/spf13/viper"

	sdk "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	osl "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestCreateCluster(ctx context.Context, options *types.TestOptions) error {

	testName := options.TestName
	targeter := generateCreateClusterTargeter(ctx, options.ID, options.Method, options.Path, options.Logger)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Metrics.Add(res)
	}

	helpers.Cleanup(ctx, options.Connection)
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
	ccsRegion := awsCreds[0].(map[string]interface{})["region"].(string)
	ccsAccessKey := awsCreds[0].(map[string]interface{})["access-key"].(string)
	ccsSecretKey := awsCreds[0].(map[string]interface{})["secret-access-key"].(string)
	ccsAccountID := awsCreds[0].(map[string]interface{})["account-id"].(string)

	targeter := func(t *vegeta.Target) error {
		fakeClusterProps := map[string]string{
			"fake_cluster": "true",
		}
		awsTags := map[string]string{
			"User": "pocm-perf",
		}
		body, err := v1.NewCluster().
			Name(fmt.Sprintf("pocm-%s-%d", id, idx)).
			Properties(fakeClusterProps).
			MultiAZ(true).
			Region(v1.NewCloudRegion().ID(ccsRegion)).
			CCS(v1.NewCCS().Enabled(true)).
			AWS(
				v1.NewAWS().
					AccessKeyID(ccsAccessKey).
					SecretAccessKey(ccsSecretKey).
					AccountID(ccsAccountID).
					Tags(awsTags),
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

func buildFakeClusters(ctx context.Context, prefix string, quantity int, ID string, conn *sdk.Connection, log logging.Logger) ([]string, error) {

	// This will take the first 4 characters of the UUID
	// Cluster Names must match the following regex:
	// ^[a-z]([-a-z0-9]*[a-z0-9])?$
	id := ID[:4]

	awsCreds := viper.Get("aws").([]any)
	if len(awsCreds) < 1 {
		log.Fatal(ctx, "No aws credentials found")
	}
	aws := awsCreds[0].(map[any]any)

	// CCS is used to create fake clusters within the AWS
	// environment supplied by the user executing this test.
	// Not fully supporting multi account now, so using first accaunt always
	ccsRegion := aws["region"].(string)
	ccsAccessKey := aws["access-key"].(string)
	ccsSecretKey := aws["secret-access-key"].(string)
	ccsAccountID := aws["account-id"].(string)

	fakeClusterProps := map[string]string{
		"fake_cluster": "true",
	}
	awsTags := map[string]string{
		"User": "pocm-perf",
	}

	clusterIDs := []string{}

	for idx := 0; idx < quantity; idx++ {
		clusterName := fmt.Sprintf("pocm-%s-%s-%d", prefix, id, idx)
		body, err := v1.NewCluster().
			Name(clusterName).
			Properties(fakeClusterProps).
			MultiAZ(true).
			Region(v1.NewCloudRegion().ID(ccsRegion)).
			CCS(v1.NewCCS().Enabled(true)).
			AWS(
				v1.NewAWS().
					AccessKeyID(ccsAccessKey).
					SecretAccessKey(ccsSecretKey).
					AccountID(ccsAccountID).
					Tags(awsTags),
			).
			Build()
		if err != nil {
			return nil, err
		}

		var raw bytes.Buffer
		err = v1.MarshalCluster(body, &raw)
		if err != nil {
			return nil, err
		}

		cID, _, err := helpers.CreateCluster(ctx, raw.String(), conn)
		if err != nil {
			return nil, err
		}
		clusterIDs = append(clusterIDs, cID)
	}

	return clusterIDs, nil
}

// Test GET with a replace of cluster_ID
func TestGETReplaceClusterID(ctx context.Context, options *types.TestOptions) error {
	clusterID := viper.GetString("cluster-id")
	if clusterID == "" {
		clusterIDs, err := buildFakeClusters(ctx, "mp", 1, options.ID, options.Connection, options.Logger)
		if err != nil {
			options.Logger.Error(ctx, "CreatingFakeclusters for test: %s", err)
			return nil
		}
		var currentTarget = 0
		clusterID = clusterIDs[currentTarget]
	}

	options.Logger.Info(ctx, "Using cluster id: %s.", clusterID)
	options.Path = strings.Replace(options.Path, "{cluster_id}", clusterID, 1)

	return TestStaticEndpoint(ctx, options)
}

// Test GET with a replace of cluster_ID
func TestSearchCluster(ctx context.Context, options *types.TestOptions) error {
	clusterID := viper.GetString("cluster-id")
	if clusterID == "" {
		clusterIDs, err := buildFakeClusters(ctx, "mp", 1, options.ID, options.Connection, options.Logger)
		if err != nil {
			options.Logger.Error(ctx, "CreatingFakeclusters for test: %s", err)
			return nil
		}
		var currentTarget = 0
		clusterID = clusterIDs[currentTarget]
	}
	options.Logger.Info(ctx, "Using cluster id: %s.", clusterID)
	options.Path = fmt.Sprintf("%s?search=id%%3D%%27%s%%27", options.Path, clusterID)

	return TestStaticEndpoint(ctx, options)
}

// Generates a targeter for the "POST /api/clusters_mgmt/v1/clusters/{cluster_id}/limited_support_reasons" endpoint
// with monotonic increasing indexes.
func generateLimitedSupportReassonsTargeter(ctx context.Context, method, url string, log logging.Logger) vegeta.Targeter {
	idx := 0
	log.Info(ctx, url)
	targeter := func(t *vegeta.Target) error {
		body, err := v1.NewLimitedSupportReason().
			CreationTimestamp(time.Now()).
			Details("test").
			ID(fmt.Sprintf("%d", idx)).
			DetectionType(v1.DetectionTypeManual).
			HREF("test").
			Summary("test").
			Build()
		if err != nil {
			return err
		}
		log.Info(ctx, "%v", body)

		var raw bytes.Buffer
		err = v1.MarshalLimitedSupportReason(body, &raw)
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

// Test Cluster Limited support reassons
func TestClusterLimitedSupportReasons(ctx context.Context, options *types.TestOptions) error {
	clusterID := viper.GetString("cluster-id")
	if clusterID == "" {
		clusterIDs, err := buildFakeClusters(ctx, "mp", 1, options.ID, options.Connection, options.Logger)
		if err != nil {
			options.Logger.Error(ctx, "CreatingFakeclusters for test: %s", err)
			return nil
		}
		var currentTarget = 0
		clusterID = clusterIDs[currentTarget]
	}
	testName := options.TestName
	options.Logger.Info(ctx, "Using cluster id: %s.", clusterID)
	options.Path = strings.Replace(options.Path, "{cluster_id}", clusterID, 1)
	targeter := generateLimitedSupportReassonsTargeter(ctx, options.Method, options.Path, options.Logger)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Metrics.Add(res)
	}

	return nil
}

// Test OSL CLuster Logs cluster_ID
func TestOSLGetClusterLogs(ctx context.Context, options *types.TestOptions) error {
	clusterID := viper.GetString("cluster-id")
	if clusterID == "" {
		clusterIDs, err := buildFakeClusters(ctx, "mp", 1, options.ID, options.Connection, options.Logger)
		if err != nil {
			options.Logger.Error(ctx, "CreatingFakeclusters for test: %s", err)
			return nil
		}
		var currentTarget = 0
		clusterID = clusterIDs[currentTarget]
	}

	options.Logger.Info(ctx, "Using cluster id: %s.", clusterID)
	body := fmt.Sprintf("{\"cluster_id\": \"%s\"}", clusterID)
	options.Body = []byte(body)
	options.Path = fmt.Sprintf("%s?cluster_id=%s", options.Path, clusterID)

	return TestStaticEndpoint(ctx, options)
}

// Generates a targeter for the "POST /api/service_logs/v1/cluster_logs" endpoint
// with monotonic increasing indexes.
func generateTestOSLPostClusterLogsTargeter(ctx context.Context, method, clusterID string, url string, log logging.Logger) vegeta.Targeter {
	idx := 0

	targeter := func(t *vegeta.Target) error {
		logEntry, err := osl.NewLogEntry().
			HREF("test").
			ID(fmt.Sprintf("%d", idx)).
			CreatedAt(time.Now()).
			CreatedBy("test").
			ClusterID(clusterID).
			Description(fmt.Sprintf("this is a log %d", idx)).
			Severity(osl.SeverityInfo).
			DocReferences("test").
			EventStreamID(fmt.Sprintf("stream-%d", idx)).
			LogType(osl.LogTypeClusterCreateHighLevel).
			ServiceName("test").
			Timestamp(time.Now()).
			Username("test").
			Summary("test").
			Link(false).
			Build()
		if err != nil {
			return err
		}
		var raw bytes.Buffer
		err = osl.MarshalLogEntry(logEntry, &raw)
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

// Test OSL Post CLuster Logs
func TestOSLPostClusterLogs(ctx context.Context, options *types.TestOptions) error {
	clusterID := viper.GetString("cluster-id")
	if clusterID == "" {
		clusterIDs, err := buildFakeClusters(ctx, "mp", 1, options.ID, options.Connection, options.Logger)
		if err != nil {
			options.Logger.Error(ctx, "CreatingFakeclusters for test: %s", err)
			return nil
		}
		var currentTarget = 0
		clusterID = clusterIDs[currentTarget]
	}

	testName := options.TestName
	targeter := generateTestOSLPostClusterLogsTargeter(ctx, options.Method, clusterID, options.Path, options.Logger)

	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, testName) {
		options.Metrics.Add(res)
	}

	return nil
}
