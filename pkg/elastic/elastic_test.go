package elastic

import (
	"context"
	"testing"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/spf13/viper"
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
		_, err := newClient(logger, ctx)
		if err != nil {
			t.Errorf("newClient() error = %v", err)
			return
		}
	})
}
