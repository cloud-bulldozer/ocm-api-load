package tests

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/config"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/elastic"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	ramp "github.com/cloud-bulldozer/ocm-api-load/pkg/ramping"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Runner prepares config and runs tests
type Runner struct {
	connection      *sdk.Connection
	logger          logging.Logger
	outputDirectory string
	testID          string
}

func NewRunner(testID, outputDirectory string, logger logging.Logger, conn *sdk.Connection) *Runner {
	return &Runner{
		connection:      conn,
		logger:          logger,
		outputDirectory: outputDirectory,
		testID:          testID,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	duration := viper.GetInt("duration")
	cooldown := viper.GetInt("cooldown")
	rate := viper.GetString("rate")
	rampType := viper.GetString("ramp-type")
	startRate := viper.GetInt("start-rate")
	endRate := viper.GetInt("end-rate")
	rampSteps := viper.GetInt("ramp-steps")
	rampDuration := viper.GetInt("ramp-duration")

	tests_conf := viper.Sub("tests")
	confHelper := config.NewConfigHelper(r.logger, tests_conf)
	for i, t := range tests {
		// Check if the test is set to run
		if !tests_conf.InConfig(t.TestName) && !tests_conf.InConfig("all") {
			continue
		}

		// Create an Attacker for each individual test. This is due to the
		// fact that vegeta (and compatible parsers, such as benchmark-wrapper)
		// expect the sequence to start at 0 for each result file. (Possibly a bug?)
		connAttacker := vegeta.Client(&http.Client{Transport: r.connection})
		attacker := vegeta.NewAttacker(connAttacker)

		// Open a file and create an encoder that will be used to store the
		// results for each test.
		fileName := fmt.Sprintf("%s_%s.json", r.testID, t.TestName)
		resultsFile, err := helpers.CreateFile(fileName, r.outputDirectory)
		if err != nil {
			return err
		}
		encoder := vegeta.NewJSONEncoder(resultsFile)

		// Bind "Test Harness"
		t.ID = r.testID
		t.Attacker = attacker
		t.Connection = r.connection
		t.Encoder = &encoder
		t.Logger = r.logger

		// Create the vegeta rate with the config values
		currentTestRate := confHelper.ResolveStringConfig(ctx, rate, fmt.Sprintf("%s.rate", t.TestName))
		rate, err := helpers.ParseRate(currentTestRate)
		if err != nil {
			r.logger.Warn(ctx,
				"error parsing rate for test %s: %s. Using default",
				t.TestName,
				currentTestRate)
		}
		t.Rate = rate

		// Check for an override on the test duration
		currentTestDuration := confHelper.ResolveIntConfig(ctx, duration, fmt.Sprintf("%s.duration", t.TestName))
		t.Duration = time.Duration(currentTestDuration) * time.Minute

		var ramper ramp.Ramper
		currentRampDuration := 0
		remainingDuration := 0
		currentTestRamp := confHelper.ResolveStringConfig(ctx, rampType, fmt.Sprintf("%s.ramp-type", t.TestName))
		if currentTestRamp != "" {
			currentStartRate := confHelper.ResolveIntConfig(ctx, startRate, fmt.Sprintf("%s.start-rate", t.TestName))
			currentEndRate := confHelper.ResolveIntConfig(ctx, endRate, fmt.Sprintf("%s.end-rate", t.TestName))
			currentSteps := confHelper.ResolveIntConfig(ctx, rampSteps, fmt.Sprintf("%s.ramp-steps", t.TestName))
			currentRampDuration = confHelper.ResolveIntConfig(ctx, rampDuration, fmt.Sprintf("%s.ramp-duration", t.TestName))
			r.logger.Info(ctx, "Validating Ramp configuration for test %s", t.TestName)
			if !confHelper.ValidateRampConfig(ctx, currentStartRate, currentEndRate, currentSteps) {
				currentTestRamp = ""
			}
			switch currentTestRamp {
			case "linear":
				ramper = ramp.NewRampingService(ramp.LinearRamp, currentStartRate, currentEndRate, currentSteps)
			case "exponential":
				ramper = ramp.NewRampingService(ramp.ExponentialRamp, currentStartRate, currentEndRate, currentSteps)
			}
		}

		if ramper == nil {
			r.logger.Info(ctx, "Executing Test: %s", t.TestName)
			r.logger.Info(ctx, "Rate: %s", t.Rate.String())
			r.logger.Info(ctx, "Duration: %s", t.Duration.String())
			r.logger.Info(ctx, "Endpoint: %s", t.Path)
			err = t.Handler(ctx, &t)
			if err != nil {
				return err
			}
		} else {
			r.logger.Info(ctx, "Executing Test: %s", t.TestName)
			r.logger.Info(ctx, "Ramp type: %s", ramper.GetType())
			r.logger.Info(ctx, "Endpoint: %s", t.Path)
			if currentRampDuration == 0 {
				duration := math.Round(t.Duration.Minutes() / float64(ramper.GetSteps()))
				t.Duration = time.Duration(duration) * time.Minute
			} else {
				remainingDuration = int(t.Duration.Minutes()) - currentRampDuration
				duration := math.Round(float64(currentRampDuration) / float64(ramper.GetSteps()))
				t.Duration = time.Duration(duration) * time.Minute
			}

			for i := 0; i < ramper.GetSteps(); i++ {
				r.logger.Info(ctx, "Ramping up... step %v", i+1)
				rateInt := ramper.NextRate()
				newRate, _ := helpers.ParseRate(fmt.Sprint(rateInt))
				t.Rate = newRate
				if i+1 == ramper.GetSteps() && remainingDuration > 0 {
					t.Duration = t.Duration + (time.Duration(remainingDuration) * time.Minute)
				}
				r.logger.Info(ctx, "Rate: %s", t.Rate.String())
				r.logger.Info(ctx, "Duration: %s", t.Duration.String())
				err = t.Handler(ctx, &t)
				if err != nil {
					return err
				}
			}
		}

		// Cleanup (cannot defer as it must happen for each test in the loop)
		r.logger.Info(ctx, "Results written to: %s", fileName)
		err = resultsFile.Close()
		if err != nil {
			return err
		}

		// Index result file
		if viper.GetString("elastic.server") != "" {
			indexer, err := elastic.NewESIndexer(ctx, r.logger)
			if err != nil {
				r.logger.Error(ctx, "obtaining indexer: %s", err)
			}
			err = indexer.IndexFile(ctx, r.testID, resultsFile.Name(), r.logger)
			if err != nil {
				r.logger.Error(ctx, "Error during ES indexing: %s", err)
			}
		}

		if i < len(tests_conf.AllSettings()) {
			r.logger.Info(ctx, "Cooling down for next test for: %v s", cooldown)
			time.Sleep(time.Duration(cooldown) * time.Second)
		}
	}
	return nil
}
