package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/cmd"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/tests"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Cobra requires a variable to store each command-line argument's value. These
// variables serve that purpose.
var (
	configFile   string
	ocmTokenURL  string
	ocmToken     string
	testID       string
	outputPath   string
	duration     int
	rate         string
	gatewayUrl   string
	clientSecret string
	clientID     string
	testName     string
	verbose      bool
	ccsAccountID string
	ccsAccessKey string
	ccsSecretKey string
	ccsRegion    string
)

const (
	longHelp = `
	A set of load tests for OCM's clusters-service, based on vegeta.
	For example:

	ocm-load-test --ocm-token=$OCM_TOKEN --duration=20m --rate=5/s --output-path=./results/$TEST_ID_$TEST_NAME.json <test_name>
`
)

var rootCmd = &cobra.Command{
	Use:   "ocm-api-load",
	Short: "A set of load tests for OCM's clusters-service, based on vegeta.",
	Long:  longHelp,
	RunE:  run,
}

func init() {

	// OCM Authentication
	rootCmd.PersistentFlags().StringVar(&ocmTokenURL, "ocm-token-url", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token", "Token URL")
	rootCmd.PersistentFlags().StringVar(&ocmToken, "ocm-token", "", "OCM Authorization token")
	rootCmd.MarkFlagRequired("ocm-token")

	// API Endpoint
	rootCmd.PersistentFlags().StringVar(&gatewayUrl, "gateway-url", "https://api.integration.openshift.com", "Gateway url to perform the test against")
	rootCmd.PersistentFlags().StringVar(&clientSecret, "client-secret", "", "Client Secret")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "Client ID")

	// Global Test Options
	rootCmd.PersistentFlags().IntVar(&duration, "duration", 1, "Duration of each individual run in minutes.")
	rootCmd.PersistentFlags().StringVar(&rate, "rate", "1/s", "Rate of the attack. Format example 5/s. (Available units 'ns', 'us', 'ms', 's', 'm', 'h')")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set this flag to activate verbose logging.")
	rootCmd.PersistentFlags().StringVar(&testName, "test-name", "", "Name of the test to run.") // TODO: Enumerate test names in help text
	rootCmd.MarkFlagRequired("duration")
	rootCmd.MarkFlagRequired("rate")
	rootCmd.MarkFlagRequired("test-name")

	// Create-Cluster (CCS) Options
	rootCmd.PersistentFlags().StringVar(&ccsAccountID, "ccs-account-id", "", "AWS Account ID for CCS User")
	rootCmd.PersistentFlags().StringVar(&ccsAccessKey, "ccs-access-key", "", "AWS Access Key for CCS User")
	rootCmd.PersistentFlags().StringVar(&ccsSecretKey, "ccs-secret-key", "", "AWS Secret Key for CCS User")
	rootCmd.PersistentFlags().StringVar(&ccsRegion, "ccs-region", "", "AWS Region for the 'create-cluster' Test")

	// Sub-Commands
	rootCmd.AddCommand(cmd.NewVersionCommand())
}

// initLogging initializes and returns the log handler
func initLogging() *logging.GoLogger {
	logger, err := logging.NewGoLoggerBuilder().
		Debug(viper.GetBool("verbose")).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't build logger: %v\n", err)
		os.Exit(1)
	}
	return logger
}

func initConnection(cmd *cobra.Command, logger *logging.GoLogger) *sdk.Connection {

	// Validate the values used here because BuildConnection does not always
	// return the most useful error messages.
	// ex: can't parse token 0: token contains an invalid number of segments
	if len(gatewayUrl) == 0 {
		logger.Fatal(cmd.Context(), "gatewayURL is invalid: %s", gatewayUrl)
	}
	if len(ocmToken) == 0 {
		logger.Fatal(cmd.Context(), "ocm-token is invalid: '%s'", ocmToken)
	}
	logger.Info(cmd.Context(), "API Endpoint: %s", gatewayUrl)

	connection, err := helpers.BuildConnection(
		gatewayUrl,
		clientID,
		clientSecret,
		ocmToken,
		logger,
		cmd.Context(),
	)
	if err != nil {
		logger.Fatal(cmd.Context(), "Unable to initialize connection object: %v", err)
	}
	return connection
}

func run(cmd *cobra.Command, args []string) error {

	// Setup Logging
	logger := initLogging()

	// Construct an OCM Connection object which is used to handle all requests
	// made to OCM and manages the associated authentication.
	connection := initConnection(cmd, logger)
	defer helpers.Cleanup(connection)

	vegetaRate, err := helpers.ParseRate(rate)
	if err != nil {
		logger.Fatal(cmd.Context(), "Error parsing rate: %v", err)
	}

	// Flag overrides config
	// Selecting test passed in the Flag
	if len(viper.GetStringSlice("test-names")) > 0 {
		viper.Set("tests", map[string]interface{}{})
		tests := viper.GetStringSlice("test-names")
		for _, t := range tests {
			viper.Set(fmt.Sprintf("tests.%s", t), map[string]interface{}{})
		}
	}

	// Require explicit declaration of the test to be ran to avoid any
	// ambiguity on what is going to be executed. This also makes it easy to
	// understand any console output that might be shared and archived in the
	// future.
	// TODO: Enumerate available test names.
	if len(testName) == 0 {
		return errors.New("No test name was specified.")
	}

	// create-cluster requires additional configuration data to be provided.
	if testName == "create-cluster" {
		if len(ccsAccountID) == 0 {
			return errors.New("ccs-account-id is required to run the create-cluster test.")
		}
		if len(viper.GetString("ccs-access-key")) == 0 {
			return errors.New("ccs-access-key is required to run the create-cluster test.")
		}
		if len(viper.GetString("ccs-secret-key")) == 0 {
			return errors.New("ccs-secret-key is required to run the create-cluster test.")
		}
		if len(viper.GetString("ccs-region")) == 0 {
			return errors.New("ccs-region is required to run the create-cluster test.")
		}
	}

	if err := tests.Run(viper.GetString("test-id"),
		viper.GetString("output-path"),
		time.Duration(viper.GetInt("duration"))*time.Minute,
		vegetaRate,
		connection,
		viper.Sub("tests"),
		logger,
		cmd.Context()); err != nil {
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
