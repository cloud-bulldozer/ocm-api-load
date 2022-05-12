package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/golang/mock/gomock"
	opensearch "github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/spf13/viper"
)

const (
	OKFileContentWithError    = `{"attack":"access-review","seq":0,"code":0,"timestamp":"2022-03-17T17:12:32.077719034+01:00","latency":788624775,"bytes_out":0,"bytes_in":0,"error":"Post \"https://api.integration.openshift.com/api/authorizations/v1/access_review\": can't send request: Post \"https://api.integration.openshift.com/api/authorizations/v1/access_review\": can't get access token: invalid_grant: Offline user session not found","body":null,"method":"POST","url":"/api/authorizations/v1/access_review","headers":null}`
	OKFileContentNoError      = `{"attack":"access-review","seq":0,"code":200,"timestamp":"2022-03-17T17:12:32.077719034+01:00","latency":788624775,"bytes_out":0,"bytes_in":0,"error":"","body":"Hello World","method":"POST","url":"/api/authorizations/v1/access_review","headers":null}`
	OKFileContentRetunrnError = `{"attack":"access-review","seq":0,"code":300,"timestamp":"2022-03-17T17:12:32.077719034+01:00","latency":788624775,"bytes_out":0,"bytes_in":0,"error":"","body":"Hello World","method":"POST","url":"/api/authorizations/v1/access_review","headers":null}`
	ErrorFileContent          = `{"attack":"access-review","seq":0,"code":200,"timestamp":"2022-03-17T17:12:32.077719034+01:00","latency":788624775,"bytes_out":0,"bytes_in":0,"error":"","body":"Hello World","method":"POST","url":"/api/authorizations/v1/access_review",`
	TetsID                    = "4059b202-dc97-4473-9b90-b8bc0e14be1b"
	Version                   = "version1"
)

func initConfig() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigFile("test.yaml")
	viper.AutomaticEnv()
}

func Test_newClient(t *testing.T) {
	lBuilder := logging.NewGoLoggerBuilder().Debug(true)
	logger, _ := lBuilder.Build()
	ctx := context.TODO()

	t.Run("NoRunningElastic", func(t *testing.T) {
		initConfig()
		config := map[string]interface{}{
			"server": "http://localhost:9200",
			"index":  "ocm-requests-test",
		}
		viper.Set("elastic", config)
		_, err := newClient(ctx, logger)
		if err != nil {
			t.Errorf("newClient() error = %v", err)
			return
		}
	})
}

func TestIndexFile(t *testing.T) {
	initConfig()
	lBuilder := logging.NewGoLoggerBuilder().Debug(true)
	logger, _ := lBuilder.Build()
	ctx := context.TODO()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dir := t.TempDir()

	// Mock general definitions
	mock := NewMockBulkIndexer(ctrl)
	mock.EXPECT().Close(ctx).AnyTimes().Return(nil)
	mock.EXPECT().Stats().Return(opensearchutil.BulkIndexerStats{
		NumAdded:    1,
		NumFlushed:  1,
		NumFailed:   0,
		NumIndexed:  1,
		NumCreated:  0,
		NumUpdated:  1,
		NumDeleted:  0,
		NumRequests: 1,
	}).AnyTimes()

	//OKFileContentWithError
	testfile001 := fmt.Sprintf("%s/%s", dir, "testfile001.txt")
	os.WriteFile(testfile001, []byte(OKFileContentWithError), 0o0777)
	_doc := doc{}
	json.Unmarshal([]byte(OKFileContentWithError), &_doc)
	if _doc.Error != "" {
		_doc.HasError = true
	}
	if _doc.Body != "" {
		_doc.HasBody = true
	}
	_doc.Uuid = TetsID
	_doc.Version = Version
	OKFileContentWithError, _ := json.Marshal(_doc)

	mock.EXPECT().Add(ctx, gomock.Eq(
		opensearchutil.BulkIndexerItem{
			Body:   bytes.NewReader(OKFileContentWithError),
			Action: "index",
		})).AnyTimes().Return(nil)

	//OKFileContentNoError
	testfile002 := fmt.Sprintf("%s/%s", dir, "testfile002.txt")
	os.WriteFile(testfile002, []byte(OKFileContentNoError), 0o0777)
	_doc = doc{}
	json.Unmarshal([]byte(OKFileContentNoError), &_doc)
	if _doc.Error != "" {
		_doc.HasError = true
	}
	if _doc.Body != "" {
		_doc.HasBody = true
	}
	_doc.Uuid = TetsID
	_doc.Version = Version
	OKFileContentNoError, _ := json.Marshal(_doc)

	mock.EXPECT().Add(ctx, gomock.Eq(
		opensearchutil.BulkIndexerItem{
			Body:   bytes.NewReader(OKFileContentNoError),
			Action: "index",
		})).AnyTimes().Return(nil)

	//OKFileContentRetunrnError
	testfile003 := fmt.Sprintf("%s/%s", dir, "testfile003.txt")
	os.WriteFile(testfile003, []byte(OKFileContentRetunrnError), 0o0777)
	_doc = doc{}
	json.Unmarshal([]byte(OKFileContentRetunrnError), &_doc)
	if _doc.Error != "" {
		_doc.HasError = true
	}
	if _doc.Body != "" {
		_doc.HasBody = true
	}
	_doc.Uuid = TetsID
	_doc.Version = Version
	OKFileContentRetunrnError, _ := json.Marshal(_doc)

	mock.EXPECT().Add(ctx, gomock.Eq(
		opensearchutil.BulkIndexerItem{
			Body:   bytes.NewReader(OKFileContentRetunrnError),
			Action: "index",
		})).AnyTimes().Return(fmt.Errorf("Error"))

	//OKFileContentNoError
	testfile004 := fmt.Sprintf("%s/%s", dir, "testfile004.txt")
	os.WriteFile(testfile004, []byte(ErrorFileContent), 0o0777)

	indexer := &ESIndexer{
		BulkIndexer: mock,
	}

	tests := []struct {
		name     string
		testID   string
		version  string
		fileName string
		wantErr  bool
	}{
		{"OKFileContentWithError", TetsID, Version, testfile001, false},
		{"OKFileContentNoError", TetsID, Version, testfile002, false},
		{"OKFileContentNoError", TetsID, Version, testfile003, true},
		{"ErrorFileContent", TetsID, Version, testfile004, true},
		{"FileDoesNotExist", TetsID, Version, path.Join("tmp", "filename.txt"), true},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if err := indexer.IndexFile(ctx, tt.testID, tt.version, tt.fileName, logger); (err != nil) != tt.wantErr {
				t.Errorf("IndexFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

	}
}

func TestNewESIndexer(t *testing.T) {
	lBuilder := logging.NewGoLoggerBuilder().Debug(true)
	logger, _ := lBuilder.Build()
	ctx := context.TODO()

	cfg := opensearch.Config{
		Addresses: []string{
			viper.GetString("elastic.server"),
		},
		Username: viper.GetString("elastic.user"),
		Password: viper.GetString("elastic.password"),
	}
	cli, _ := opensearch.NewClient(cfg)
	bulkConfig := opensearchutil.BulkIndexerConfig{
		Index:  viper.GetString("elastic.index"),
		Client: cli,
		OnError: func(ctx context.Context, err error) {
			logger.Error(ctx, "%s", err)
		},
		ErrorTrace: true,
	}
	indexer, _ := opensearchutil.NewBulkIndexer(bulkConfig)

	want := &ESIndexer{BulkIndexer: indexer}

	t.Run("New ESIndexer", func(t *testing.T) {
		got, err := NewESIndexer(ctx, logger)
		if (err != nil) != false {
			t.Errorf("NewESIndexer() error = %v, wantErr %v", err, false)
			return
		}

		if reflect.TypeOf(got.BulkIndexer) != reflect.TypeOf(want.BulkIndexer) {
			t.Errorf("NewESIndexer() want = %v, got %v", want, got)
			return
		}
	})

}
