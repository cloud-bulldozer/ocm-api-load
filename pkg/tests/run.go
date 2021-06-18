package tests

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func Run(
	testID string,
	outputDirectory string,
	duration time.Duration,
	rate vegeta.Rate,
	connection *sdk.Connection,
	viper *viper.Viper,
	logger logging.Logger,
	ctx context.Context) error {

	for _, t := range tests {
		// Check if the test is set to run
		if !viper.InConfig(t.TestName) && !viper.InConfig("all") {
			continue
		}

		// Create an Attacker for each individual test. This is due to the
		// fact that vegeta (and compatible parsers, such as benchmark-wrapper)
		// expect the sequence to start at 0 for each result file. (Possibly a bug?)
		connAttacker := vegeta.Client(&http.Client{Transport: connection})
		attacker := vegeta.NewAttacker(connAttacker)

		// Open a file and create an encoder that will be used to store the
		// results for each test.
		fileName := fmt.Sprintf("%s_%s.json", testID, t.TestName)
		resultsFile, err := helpers.CreateFile(fileName, outputDirectory)
		if err != nil {
			return err
		}
		encoder := vegeta.NewJSONEncoder(resultsFile)

		// Bind "Test Harness"
		t.ID = testID
		t.Attacker = attacker
		t.Connection = connection
		t.Encoder = &encoder
		t.Logger = logger
		t.Context = ctx

		// Bind additional test options
		// TODO: These are only required for the `create-cluster` test. Find
		//       a better way to pass these to the test(s) which requires them.
		//       Don't forget to validate they are set on startup as opposed to
		//       waiting for the dependent test to start running.
		t.CCSAccessKey = viper.GetString("ccs-account-key")
		t.CCSSecretKey = viper.GetString("ccs-secret-key")
		t.CCSAccountID = viper.GetString("ccs-account-id")
		t.CCSRegion = viper.GetString("ccs-region")

		fmt.Printf("Test Options: %+v\n", t)
		fmt.Printf("Test Option Access Key: %s", t.CCSAccessKey)

		// Create the vegeta rate with the config values
		if viper.GetString(fmt.Sprintf("%s.rate", t.TestName)) == "" {
			logger.Info(ctx, "no specific rate for test %s. Using default", t.TestName)
			t.Rate = rate
		} else {
			r, err := helpers.ParseRate(viper.GetString(fmt.Sprintf("%s.rate", t.TestName)))
			if err != nil {
				logger.Warn(ctx, "error parsing rate for test %s: %s. Using default", t.TestName, fmt.Sprintf("%s.rate", t.TestName))
				t.Rate = rate
			} else {
				t.Rate = r
			}
		}

		// Check for an override on the test duration
		dur := viper.GetInt(fmt.Sprintf("%s.duration", t.TestName))
		if dur == 0 {
			// Using default
			t.Duration = duration
		} else {
			t.Duration = time.Duration(dur) * time.Minute
		}

		logger.Info(ctx, "Executing Test: %s", t.TestName)
		logger.Info(ctx, "Rate: %s", t.Rate.String())
		logger.Info(ctx, "Duration: %s", t.Duration.String())
		logger.Info(ctx, "Endpoint: %s", t.Path)
		err = t.Handler(&t)
		if err != nil {
			return err
		}

		// Cleanup (cannot defer as it must happen for each test in the loop)
		logger.Info(ctx, "Results written to: %s", fileName)
		err = resultsFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
