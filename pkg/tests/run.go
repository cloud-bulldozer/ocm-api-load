package tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/ocm"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/providers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func Run(tc types.TestConfiguration) error {
	logger := tc.Logger

	tests_conf := viper.Sub("tests")
	connections, err := ocm.BuildConnections(tc.Ctx, tc.Logger)
	if err != nil {
		return err
	}

	rrConnections := providers.NewRoundRobinProvider(connections...)

	for i, t := range tests {
		// Check if the test is set to run
		if !tests_conf.InConfig(t.TestName) && !tests_conf.InConfig("all") {
			continue
		}

		// Create an Attacker for each individual test. This is due to the
		// fact that vegeta (and compatible parsers, such as benchmark-wrapper)
		// expect the sequence to start at 0 for each result file. (Possibly a bug?)

		connAttacker := vegeta.Client(&http.Client{Transport: rrConnections.GetItem().(ocm.Connector).GetConnection()})
		attacker := vegeta.NewAttacker(connAttacker)

		// Open a file and create an encoder that will be used to store the
		// results for each test.
		fileName := fmt.Sprintf("%s_%s.json", tc.TestID, t.TestName)
		resultsFile, err := helpers.CreateFile(fileName, tc.OutputDirectory)
		if err != nil {
			return err
		}
		encoder := vegeta.NewJSONEncoder(resultsFile)

		// Bind "Test Harness"
		t.ID = tc.TestID
		t.Attacker = attacker
		t.Connection = rrConnections.GetItem().(ocm.Connector).GetConnection()
		t.Encoder = &encoder
		t.Logger = logger
		t.Context = tc.Ctx

		// Create the vegeta rate with the config values
		current_test_rate := tests_conf.GetString(fmt.Sprintf("%s.rate", t.TestName))
		if current_test_rate == "" {
			logger.Info(tc.Ctx, "no specific rate for test %s. Using default", t.TestName)
			t.Rate = tc.Rate
		} else {
			r, err := helpers.ParseRate(current_test_rate)
			if err != nil {
				logger.Warn(tc.Ctx, "error parsing rate for test %s: %s. Using default", t.TestName,
					current_test_rate)
				t.Rate = tc.Rate
			} else {
				t.Rate = r
			}
		}

		// Check for an override on the test duration
		dur := tests_conf.GetInt(fmt.Sprintf("%s.duration", t.TestName))
		if dur == 0 {
			// Using default
			t.Duration = tc.Duration
		} else {
			t.Duration = time.Duration(dur) * time.Minute
		}

		logger.Info(tc.Ctx, "Executing Test: %s", t.TestName)
		logger.Info(tc.Ctx, "Rate: %s", t.Rate.String())
		logger.Info(tc.Ctx, "Duration: %s", t.Duration.String())
		logger.Info(tc.Ctx, "Endpoint: %s", t.Path)
		err = t.Handler(&t)
		if err != nil {
			return err
		}

		// Cleanup (cannot defer as it must happen for each test in the loop)
		logger.Info(tc.Ctx, "Results written to: %s", fileName)
		err = resultsFile.Close()
		if err != nil {
			return err
		}

		if i+1 < len(tests_conf.AllSettings()) {
			logger.Info(tc.Ctx, "Cooling down for next test for: %v s", tc.Cooldown.Seconds())
			time.Sleep(tc.Cooldown)
		}
	}
	return nil
}
