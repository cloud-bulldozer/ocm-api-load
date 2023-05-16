package elastic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cloud-bulldozer/go-commons/indexers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/spf13/viper"
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
		// Username: viper.GetString("elastic.user"),
		// Password: viper.GetString("elastic.password"),
	}
	client, err := indexers.NewIndexer(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func IndexFile(ctx context.Context, testID string, version string, fileName string, logger logging.Logger) error {
	indexer, err := newClient(ctx, logger)
	if err != nil {
		logger.Error(ctx, "obtaining indexer: %s", err)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	fileReader := bufio.NewReader(file)

	docs := []interface{}{}
	var errors string
	for {
		line, pref, err := fileReader.ReadLine()
		if err == io.EOF {
			break
		}

		fullLine := bytes.Join([][]byte{line}, []byte(""))
		if pref {
			for {
				l, p, err := fileReader.ReadLine()
				if err == io.EOF {
					break
				}
				if err != nil {
					errors = fmt.Sprintf("%s\n%s", errors, err)
					break
				}
				fullLine = bytes.Join([][]byte{fullLine, l}, []byte(""))

				if !p {
					break
				}
			}
		}

		_doc := doc{}
		err = json.Unmarshal(fullLine, &_doc)
		if err != nil {
			errors = fmt.Sprintf("%s\n%s", errors, err)
			continue
		}
		if _doc.Error != "" {
			_doc.HasError = true
		}
		if _doc.Body != "" {
			_doc.HasBody = true
		}
		_doc.Uuid = testID
		_doc.Version = version

		docs = append(docs, _doc)
	}
	resp, err := (*indexer).Index(docs, indexers.IndexingOpts{
		JobName: testID,
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
