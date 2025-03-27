package elastic

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloud-bulldozer/go-commons/indexers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func newClient(ctx context.Context, logger logging.Logger) (*indexers.Indexer, error) {
	logger.Info(ctx, "Building Indexer configuration")
	if viper.GetString("elastic.server") == "" {
		logger.Debug(ctx, "Using local server")
		config := indexers.IndexerConfig{
			Type:             indexers.LocalIndexer,
			MetricsDirectory: viper.GetString("output-path"),
		}
		client, err := indexers.NewIndexer(config)
		if err != nil {
			return nil, err
		}
		return client, nil
	} else {
		logger.Debug(ctx, "Using server: %s", viper.GetString("elastic.server"))
		config := indexers.IndexerConfig{
			Type: indexers.OpenSearchIndexer,
			Servers: []string{
				viper.GetString("elastic.server"),
			},
			Index:              viper.GetString("elastic.index"),
			InsecureSkipVerify: viper.GetBool("elastic.insecure-skip-verify"),
		}
		client, err := indexers.NewIndexer(config)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
}

func IndexFile(ctx context.Context, testID string, version string, attack string, fileName string, metrics vegeta.Metrics, logger logging.Logger) error {
	indexer, err := newClient(ctx, logger)
	if err != nil {
		logger.Error(ctx, "obtaining indexer: %s", err)
	}

	env := extractENVContent()

	var errors string
	_doc := doc{}
	_doc.Metrics = metrics
	_doc.Uuid = testID
	_doc.Version = version
	_doc.Attack = attack
	_doc.BuildURL = env["buildUrl"]
	_doc.CiSystem = env["CiSystem"]

	resp, err := (*indexer).Index([]interface{}{_doc}, indexers.IndexingOpts{
		MetricName: strings.Join([]string{testID, attack}, "-"),
	})
	if err != nil {
		errors = fmt.Sprintf("%s\n%s", errors, err)
	}

	logger.Info(ctx, resp)
	if errors != "" {
		return fmt.Errorf("BulkIndexer Error: %s", errors)
	}
	return nil
}

func extractENVContent() map[string]string {
	result := make(map[string]string)
	buildURL := os.Getenv("BUILD_URL")
	if buildURL != "" {
		result["buildUrl"] = buildURL
	} else {
		result["buildUrl"] = "Manual run"
	}
	dagID := os.Getenv("AIRFLOW_CTX_DAG_ID")
	jenkinsURL := os.Getenv("JENKINS_URL")
	if jenkinsURL != "" {
		result["CiSystem"] = "Jenkins"
	} else if dagID != "" {
		result["CiSystem"] = "Airflow"
	} else {
		result["CiSystem"] = "Local"
	}
	return result
}
