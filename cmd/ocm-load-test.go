package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/cmd"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/tests"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	configFile string
)

const (
	longHelp = `
	A set of load tests for OCM's clusters-service, based on vegeta.
	For example:

	ocm-load-test --test-id=foo --ocm-token=$OCM_TOKEN --duration=20m --rate=5/s --output-path=./results/$TEST_ID_$TEST_NAME.json <test_name>

	Or

	ocm-load-test --config-file=config.yaml
`
)

var rootCmd = &cobra.Command{
	Use:   "ocm-api-load",
	Short: "A set of load tests for OCM's clusters-service, based on vegeta.",
	Long:  longHelp,
	RunE:  run,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVar(&configFile, "config-file", "config.yaml", "config file")
	rootCmd.Flags().String("ocm-token-url", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token", "Token URL")
	rootCmd.Flags().String("ocm-token", "", "OCM Authorization token")
	rootCmd.Flags().String("gateway-url", "https://api.integration.openshift.com", "Gateway url to perform the test against")
	rootCmd.Flags().String("test-id", uuid.NewV4().String(), "Unique ID to identify the test run. UUID is recommended")
	rootCmd.Flags().String("output-path", "results", "Output directory for result and report files")
	rootCmd.Flags().Int("duration", 1, "Duration of each individual run in minutes.")
	rootCmd.Flags().String("rate", "1/s", "Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h')")
	rootCmd.Flags().StringSlice("test-names", []string{}, "Names for the tests to be run.")
	rootCmd.Flags().BoolP("verbose", "v", false, "set this flag to activate verbose logging.")
	rootCmd.Flags().Int("cooldown", 10, "Cooldown time between tests in seconds.")
	// AWS config
	// If needed to use multiple AWS account, use the config file
	rootCmd.Flags().String("aws-region", "us-west-1", "AWS region")
	rootCmd.Flags().String("aws-access-key", "", "AWS access key")
	rootCmd.Flags().String("aws-access-secret", "", "AWS access secret")
	rootCmd.Flags().String("aws-account-id", "", "AWS Account ID, is the 12-digit account number.")
	rootCmd.AddCommand(cmd.NewVersionCommand())
}

func initConfig() {
	viper.SetConfigType("yaml")
	if configFile == "" {
		viper.AddConfigPath(".")
	}
	viper.SetConfigFile(configFile)
	viper.BindPFlags(rootCmd.Flags())

	viper.AutomaticEnv()

	if _, err := os.Stat(viper.GetString("config-file")); err != nil {
		viper.WriteConfig()
	} else {
		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}

// configTests decides wether to use Flags values or config file
func configTests() {
	// Flag overrides config
	// Selecting test passed in the Flag
	if len(viper.GetStringSlice("test-names")) > 0 {
		viper.Set("tests", map[string]interface{}{})
		tests := viper.GetStringSlice("test-names")
		for _, t := range tests {
			viper.Set(fmt.Sprintf("tests.%s", t), map[string]interface{}{})
		}
	}

	// If no Flag or Config is passed all test should run
	if len(viper.GetStringSlice("test-names")) == 0 && len(viper.GetStringMap("tests")) == 0 {
		viper.Set("tests.all", map[string]interface{}{})
	}
}

// configAWS decides wether to use Flags values or config file
func configAWS(ctx context.Context, logger *logging.GoLogger) {
	// Flag overrides config
	// Selecting test passed in the Flag
	if viper.GetString("aws-account-id") != "" {
		if viper.GetString("aws-access-key") == "" || viper.GetString("aws-access-secret") == "" {
			logger.Fatal(ctx, "AWS configuration not complete")
		}
		config := []interface{}{map[interface{}]interface{}{
			"region":            viper.GetString("aws-region"),
			"access-key":        viper.GetString("aws-access-key"),
			"secret-access-key": viper.GetString("aws-access-secret"),
			"account-id":        viper.GetString("aws-account-id"),
		}}
		viper.Set("aws", config)
	}

	// If no Flag or Config is passed all test should run
	if len(viper.Get("aws").([]interface{})) < 1 {
		logger.Fatal(ctx, "AWS configuration not provided")
	}
}

func run(cmd *cobra.Command, args []string) error {
	logger, err := logging.NewGoLoggerBuilder().
		Debug(viper.GetBool("verbose")).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't build logger: %v\n", err)
		os.Exit(1)
	}

	if viper.GetString("ocm-token") == "" {
		logger.Fatal(cmd.Context(), "ocm-token is a necessary configuration")
	}
	err = helpers.CreateFolder(viper.GetString("output-path"), logger)
	if err != nil {
		logger.Fatal(cmd.Context(), "creating api connection: %v", err)
	}
	logger.Info(cmd.Context(), "Using output directory: %s", viper.GetString("output-path"))

	connection, err := helpers.BuildConnection(viper.GetString("gateway-url"),
		viper.GetString("client.id"),
		viper.GetString("client.secret"),
		viper.GetString("ocm-token"),
		logger,
		cmd.Context())
	if err != nil {
		logger.Fatal(cmd.Context(), "creating api connection: %v", err)
	}
	defer helpers.Cleanup(connection)

	vegetaRate, err := helpers.ParseRate(viper.GetString("rate"))
	if err != nil {
		logger.Fatal(cmd.Context(), "parsing rate: %v", err)
	}

	configTests()

	configAWS(cmd.Context(), logger)

	testConfig := types.TestConfiguration{
		TestID:          viper.GetString("test-id"),
		OutputDirectory: viper.GetString("output-path"),
		Duration:        time.Duration(viper.GetInt("duration")) * time.Minute,
		Cooldown:        time.Duration(viper.GetInt("cooldown")) * time.Second,
		Rate:            vegetaRate,
		Connection:      connection,
		Logger:          logger,
		Ctx:             cmd.Context(),
	}

	if err := tests.Run(testConfig); err != nil {
		logger.Fatal(cmd.Context(), "running load test: %v", err)
	}

	return nil
}

func main() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
