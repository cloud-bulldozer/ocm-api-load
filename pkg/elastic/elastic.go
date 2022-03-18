package elastic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	opensearch "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/spf13/viper"
)

func newClient(logger logging.Logger, ctx context.Context) (*opensearch.Client, error) {
	logger.Info(ctx, "Building ES configuration")
	logger.Debug(ctx, "Using server: %s", viper.GetString("elastic.server"))
	cfg := opensearch.Config{
		Addresses: []string{
			viper.GetString("elastic.server"),
		},
		Username: viper.GetString("elastic.user"),
		Password: viper.GetString("elastic.password"),
	}
	return opensearch.NewClient(cfg)

}

func IndexFile(testID string, fileName string, logger logging.Logger, ctx context.Context) error {
	cli, err := newClient(logger, ctx)
	if err != nil {
		return err
	}

	var errors string
	bulkConfig := opensearchutil.BulkIndexerConfig{
		Index:  viper.GetString("elastic.index"),
		Client: cli,
		OnError: func(ctx context.Context, err error) {
			errors = fmt.Sprintf("%s\n%s", errors, err)
		},
		OnFlushEnd: func(ctx context.Context) {
			fmt.Printf("%v", ctx)
		},
		ErrorTrace: true,
	}
	bulkIndexer, err := opensearchutil.NewBulkIndexer(bulkConfig)
	if err != nil {
		return err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		_doc := doc{}
		json.Unmarshal(fileScanner.Bytes(), &_doc)
		if _doc.Error != "" {
			_doc.HasError = true
		}
		if _doc.Body != "" {
			_doc.HasBody = true
		}
		_doc.Uuid = testID
		m, _ := json.Marshal(_doc)
		bulkIndexer.Add(ctx, opensearchutil.BulkIndexerItem{
			Body:   bytes.NewReader(m),
			Action: "index",
		})
	}

	bulkIndexer.Close(ctx)
	logger.Info(ctx,
		"BulkIndexer Stats:\nNumAdded: %d\t\tNumCreate: %d\t\tNumFailed: %d",
		bulkIndexer.Stats().NumAdded,
		bulkIndexer.Stats().NumCreated,
		bulkIndexer.Stats().NumFailed)

	if errors != "" {
		return fmt.Errorf("BulkIndexer Error: %s", errors)
	}
	return nil
}
