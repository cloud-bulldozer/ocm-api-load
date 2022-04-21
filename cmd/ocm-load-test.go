package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/cmd"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/tests"
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
	//Flags with defaults
	rootCmd.Flags().StringVar(&configFile, "config-file", "config.yaml", "config file")
	rootCmd.Flags().String("ocm-token-url", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token", "Token URL")
	rootCmd.Flags().String("gateway-url", "https://api.integration.openshift.com", "Gateway url to perform the test against")
	rootCmd.Flags().String("test-id", uuid.NewV4().String(), "Unique ID to identify the test run. UUID is recommended")
	rootCmd.Flags().String("output-path", "results", "Output directory for result and report files")
	rootCmd.Flags().Int("duration", 1, "Duration of each individual run in minutes.")
	rootCmd.Flags().String("rate", "1/s", "Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h')")
	rootCmd.Flags().BoolP("verbose", "v", false, "set this flag to activate verbose logging.")
	rootCmd.Flags().Int("cooldown", 10, "Cooldown time between tests in seconds.")
	rootCmd.Flags().StringSlice("test-names", []string{}, "Names for the tests to be run.")
	//Elasticsearch Flags
	rootCmd.Flags().String("elastic-server", "", "Elasticsearch cluster URL")
	rootCmd.Flags().String("elastic-user", "", "Elasticsearch User for authentication")
	rootCmd.Flags().Bool("elastic-insecure-skip-verify", false, "Elasticsearch skip tls verifcation during authentication")
	rootCmd.Flags().String("elastic-password", "", "Elasticsearch Password for authentication")
	rootCmd.Flags().String("elastic-index", "", "Elasticsearch index to store the documents")
	//Ramping Flags
	rootCmd.Flags().String("ramp-type", "", "Type of ramp to use for all tests. (linear, exponential)")
	rootCmd.Flags().Int("start-rate", 0, "Starting request per second rate. (E.g.: 5 would be 5 req/s)")
	rootCmd.Flags().Int("end-rate", 0, "Ending request per second rate. (E.g.: 5 would be 5 req/s)")
	rootCmd.Flags().Int("ramp-steps", 0, "Number of stepts to get from start rate to end rate. (Minimum 2 steps)")
	rootCmd.Flags().Int("ramp-duration", 0, "Duration of ramp in minutes, before normal execution")

	//Required flags
	rootCmd.Flags().String("ocm-token", "", "OCM Authorization token")
	// AWS config
	// If multiple AWS account are needed use the config file
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
func configAWS() error {
	// Flag overrides config
	// Selecting aws config passed by flags
	if viper.GetString("aws-account-id") != "" {
		if viper.GetString("aws-access-key") == "" || viper.GetString("aws-access-secret") == "" {
			return fmt.Errorf("AWS configuration not complete")
		}
		config := []interface{}{map[interface{}]interface{}{
			"region":            viper.GetString("aws-region"),
			"access-key":        viper.GetString("aws-access-key"),
			"secret-access-key": viper.GetString("aws-access-secret"),
			"account-id":        viper.GetString("aws-account-id"),
		}}
		viper.Set("aws", config)
	}

	// If no Flag or Config is passed test should fail
	if !viper.IsSet("aws") || len(viper.Get("aws").([]interface{})) < 1 {
		return fmt.Errorf("AWS configuration not provided")
	}

	// If multiple accounts are passed.
	if len(viper.Get("aws").([]interface{})) > 1 {
		return fmt.Errorf("multiple AWS accounts are not supported at the moment")
	}
	return nil
}

// configES decides wether to use Flags values or config file
func configES() error {
	// Flag overrides config
	// Selecting ES config passed by flags
	// If no Flag or Config is passed the test will not index the documents.
	if viper.GetString("elastic-server") != "" {
		if viper.GetString("elastic-index") == "" {
			return fmt.Errorf("ES configuration needs an index set `elastic-index` flag")
		}
		config := map[string]interface{}{
			"server":   viper.GetString("elastic-server"),
			"user":     viper.GetString("elastic-user"),
			"insecure-skip-verify":     viper.GetBool("elastic-insecure-skip-verify"),
			"password": viper.GetString("elastic-password"),
			"index":    viper.GetString("elastic-index"),
		}
		viper.Set("elastic", config)
	}
	return nil
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
	err = helpers.CreateFolder(cmd.Context(), viper.GetString("output-path"), logger)
	if err != nil {
		logger.Fatal(cmd.Context(), "creating api connection: %v", err)
	}
	logger.Info(cmd.Context(), "Using output directory: %s", viper.GetString("output-path"))

	connection, err := helpers.BuildConnection(cmd.Context(), viper.GetString("gateway-url"),
		viper.GetString("client.id"),
		viper.GetString("client.secret"),
		viper.GetString("ocm-token"),
		logger,
	)
	if err != nil {
		logger.Fatal(cmd.Context(), "creating api connection: %v", err)
	}
	defer helpers.Cleanup(cmd.Context(), connection)

	configTests()

	err = configAWS()
	if err != nil {
		logger.Fatal(cmd.Context(), "Configuring AWS: %s", err)
	}

	err = configES()
	if err != nil {
		logger.Fatal(cmd.Context(), "Configuring ES: %s", err)
	}

	runner := tests.NewRunner(
		viper.GetString("test-id"),
		viper.GetString("output-path"),
		logger,
		connection,
	)

	if err := runner.Run(cmd.Context()); err != nil {
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
