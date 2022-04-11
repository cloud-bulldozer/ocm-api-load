package elastic

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	opensearch "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/spf13/viper"
)

type ESIndexer struct {
	BulkIndexer opensearchutil.BulkIndexer
}

func NewESIndexer(ctx context.Context, logger logging.Logger) (*ESIndexer, error) {
	bulkIndexer, err := newBulkIndexer(ctx, logger)
	if err != nil {
		return nil, err
	}
	return &ESIndexer{
		BulkIndexer: bulkIndexer,
	}, nil

}

func newClient(ctx context.Context, logger logging.Logger) (*opensearch.Client, error) {
	logger.Info(ctx, "Building ES configuration")
	logger.Debug(ctx, "Using server: %s", viper.GetString("elastic.server"))
	cfg := opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("elastic.insecure-skip-verify")},
		},
		Addresses: []string{
			viper.GetString("elastic.server"),
		},
		Username: viper.GetString("elastic.user"),
		Password: viper.GetString("elastic.password"),
	}
	return opensearch.NewClient(cfg)

}

func newBulkIndexer(ctx context.Context, logger logging.Logger) (opensearchutil.BulkIndexer, error) {
	cli, err := newClient(ctx, logger)
	if err != nil {
		return nil, err
	}

	bulkConfig := opensearchutil.BulkIndexerConfig{
		Index:  viper.GetString("elastic.index"),
		Client: cli,
		OnError: func(ctx context.Context, err error) {
			logger.Error(ctx, "%s", err)
		},
		ErrorTrace: true,
	}
	bulkIndexer, err := opensearchutil.NewBulkIndexer(bulkConfig)
	if err != nil {
		return nil, err
	}
	return bulkIndexer, nil
}

func (in *ESIndexer) IndexFile(ctx context.Context, testID string, version string, fileName string, logger logging.Logger) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	fileReader := bufio.NewReader(file)

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
		m, err := json.Marshal(_doc)
		if err != nil {
			errors = fmt.Sprintf("%s\n%s", errors, err)
			continue
		}
		err = in.BulkIndexer.Add(ctx, opensearchutil.BulkIndexerItem{
			Body:   bytes.NewReader(m),
			Action: "index",
		})
		if err != nil {
			errors = fmt.Sprintf("%s\n%s", errors, err)
		}
	}

	in.BulkIndexer.Close(ctx)
	logger.Info(ctx,
		"BulkIndexer Stats:\nNumAdded: %d\t\tNumCreate: %d\t\tNumFailed: %d",
		in.BulkIndexer.Stats().NumAdded,
		in.BulkIndexer.Stats().NumCreated,
		in.BulkIndexer.Stats().NumFailed)

	if errors != "" {
		return fmt.Errorf("BulkIndexer Error: %s", errors)
	}
	return nil
}
