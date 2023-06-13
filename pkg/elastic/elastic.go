package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/cloud-bulldozer/go-commons/indexers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func newClient(ctx context.Context, logger logging.Logger) (*indexers.Indexer, error) {
	logger.Info(ctx, "Building ES configuration")
	logger.Debug(ctx, "Using server: %s", viper.GetString("elastic.server"))
	config := indexers.IndexerConfig{
		Type: indexers.ElasticIndexer,
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

func IndexFile(ctx context.Context, testID string, version string, attack string, fileName string, metrics vegeta.Metrics, logger logging.Logger) error {
	indexer, err := newClient(ctx, logger)
	if err != nil {
		logger.Error(ctx, "obtaining indexer: %s", err)
	}

	var errors string
	_doc := doc{}
	_doc.Metrics = metrics
	_doc.Uuid = testID
	_doc.Version = version
	_doc.Attack = attack

	resp, err := (*indexer).Index([]interface{}{_doc}, indexers.IndexingOpts{
		JobName: testID,
	})
	if err != nil {
		errors = fmt.Sprintf("%s\n%s", errors, err)
	}

	file := path.Join(viper.GetString("output-path"), fileName)

	logger.Info(ctx, "Writing results to: %s", file)
	data, err := json.Marshal(_doc)
	if err != nil {
		logger.Error(ctx, "Error during json marshal: %s", err)
	} else {
		err = os.WriteFile(file, data, 0666)
		if err != nil {
			logger.Error(ctx, "Error writing file: %s", err)
		}
	}

	logger.Info(ctx, resp)
	if errors != "" {
		return fmt.Errorf("BulkIndexer Error: %s", errors)
	}
	return nil
}
