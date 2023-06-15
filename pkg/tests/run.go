package tests

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/config"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/elastic"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	ramp "github.com/cloud-bulldozer/ocm-api-load/pkg/ramping"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Runner prepares config and runs tests
type Runner struct {
	connections     []*sdk.Connection
	logger          logging.Logger
	outputDirectory string
	testID          string
}

func NewRunner(testID, outputDirectory string, logger logging.Logger, connections []*sdk.Connection) *Runner {
	return &Runner{
		connections:     connections,
		logger:          logger,
		outputDirectory: outputDirectory,
		testID:          testID,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	r.logger.Info(ctx, "UUID: %s", r.testID)
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

	var wg sync.WaitGroup
	concurrentConnections := len(r.connections)
	for i, t := range tests {
		// Check if the test is set to run
		if !tests_conf.InConfig(t.TestName) && !tests_conf.InConfig("all") {
			continue
		} else {
			wg.Add(concurrentConnections)
		}

		for i, conn := range r.connections {
			go func(ctx context.Context, concurrentConnections int, index int, conn *sdk.Connection, testOptions types.TestOptions) error {
				// Create an Attacker for each individual test. This is due to the
				// fact that vegeta (and compatible parsers, such as benchmark-wrapper)
				// expect the sequence to start at 0 for each result file. (Possibly a bug?)
				connAttacker := vegeta.Client(&http.Client{Transport: conn})
				attacker := vegeta.NewAttacker(connAttacker)

				// Open a file and create an encoder that will be used to store the
				// results for each test.
				fileName := fmt.Sprintf("%s_%s_%d.json", r.testID, testOptions.TestName, index)

				var metrics vegeta.Metrics
				metrics.Histogram = &vegeta.Histogram{
					Buckets: vegeta.Buckets{time.Millisecond},
				}

				// Bind "Test Harness"
				testOptions.ID = r.testID
				testOptions.Attacker = attacker
				testOptions.Connection = conn
				testOptions.Metrics = &metrics
				testOptions.Logger = r.logger

				// Create the vegeta rate with the config values
				currentTestRate := confHelper.ResolveStringConfig(ctx, rate, fmt.Sprintf("%s.rate", testOptions.TestName))
				rate, err := helpers.ParseRate(currentTestRate, concurrentConnections)
				if err != nil {
					r.logger.Warn(ctx,
						"error parsing rate for test %s: %s. Using default",
						testOptions.TestName,
						currentTestRate)
				}
				testOptions.Rate = rate

				// Check for an override on the test duration
				currentTestDuration := confHelper.ResolveIntConfig(ctx, duration, fmt.Sprintf("%s.duration", testOptions.TestName))
				testOptions.Duration = time.Duration(currentTestDuration) * time.Minute

				remainingDuration := 0
				currentTestRamp := confHelper.ResolveStringConfig(ctx, rampType, fmt.Sprintf("%s.ramp-type", testOptions.TestName))
				currentRampDuration, ramper := buildRamper(ctx, currentTestRamp, confHelper, startRate, testOptions, endRate, rampSteps, rampDuration, r)

				if ramper == nil {
					r.logger.Info(ctx, "Executing Test: %s", testOptions.TestName)
					r.logger.Info(ctx, "Rate: %s", testOptions.Rate.String())
					r.logger.Info(ctx, "Duration: %s", testOptions.Duration.String())
					r.logger.Info(ctx, "Endpoint: %s", testOptions.Path)
					err = testOptions.Handler(ctx, &testOptions)
					if err != nil {
						return err
					}
				} else {
					r.logger.Info(ctx, "Executing Test: %s", testOptions.TestName)
					r.logger.Info(ctx, "Ramp type: %s", ramper.GetType())
					r.logger.Info(ctx, "Endpoint: %s", testOptions.Path)
					if currentRampDuration == 0 {
						duration := math.Round(testOptions.Duration.Minutes() / float64(ramper.GetSteps()))
						testOptions.Duration = time.Duration(duration) * time.Minute
					} else {
						remainingDuration = int(testOptions.Duration.Minutes()) - currentRampDuration
						duration := math.Round(float64(currentRampDuration) / float64(ramper.GetSteps()))
						testOptions.Duration = time.Duration(duration) * time.Minute
					}

					for i := 0; i < ramper.GetSteps(); i++ {
						r.logger.Info(ctx, "Ramping up... step %v", i+1)
						rateInt := ramper.NextRate()
						newRate, _ := helpers.ParseRate(fmt.Sprint(rateInt), concurrentConnections)
						testOptions.Rate = newRate
						if i+1 == ramper.GetSteps() && remainingDuration > 0 {
							testOptions.Duration = testOptions.Duration + (time.Duration(remainingDuration) * time.Minute)
						}
						r.logger.Info(ctx, "Rate: %s", testOptions.Rate.String())
						r.logger.Info(ctx, "Duration: %s", testOptions.Duration.String())
						err = testOptions.Handler(ctx, &testOptions)
						if err != nil {
							return err
						}
					}
				}

				// Index result file

				serverVersion := helpers.GetServerVersion(ctx, conn)
				r.logger.Info(ctx, "server version %s", serverVersion)
				err = elastic.IndexFile(ctx, r.testID, serverVersion, testOptions.TestName, fileName, *testOptions.Metrics, r.logger)
				if err != nil {
					r.logger.Error(ctx, "%s", err)
				}

				testOptions.Metrics.Close()
				wg.Done()
				return nil
			}(ctx, concurrentConnections, i, conn, t)
		}
		wg.Wait()

		if i < len(tests_conf.AllSettings()) {
			r.logger.Info(ctx, "Cooling down for next test for: %v s", cooldown)
			time.Sleep(time.Duration(cooldown) * time.Second)
		}
	}
	return nil
}

func buildRamper(ctx context.Context, currentTestRamp string, confHelper *config.ConfigHelper, startRate int, t types.TestOptions, endRate int, rampSteps int, rampDuration int, r *Runner) (int, ramp.Ramper) {
	var ramper ramp.Ramper
	var currentRampDuration int
	if currentTestRamp != "" {
		currentRampDuration := confHelper.ResolveIntConfig(ctx, rampDuration, fmt.Sprintf("%s.ramp-duration", t.TestName))
		currentStartRate := confHelper.ResolveIntConfig(ctx, startRate, fmt.Sprintf("%s.start-rate", t.TestName))
		currentEndRate := confHelper.ResolveIntConfig(ctx, endRate, fmt.Sprintf("%s.end-rate", t.TestName))
		currentSteps := confHelper.ResolveIntConfig(ctx, rampSteps, fmt.Sprintf("%s.ramp-steps", t.TestName))
		r.logger.Info(ctx, "Validating Ramp configuration for test %s", t.TestName)
		if !confHelper.ValidateRampConfig(ctx, currentStartRate, currentEndRate, currentSteps) {
			return currentRampDuration, ramper
		}
		switch currentTestRamp {
		case "linear":
			ramper = ramp.NewRampingService(ramp.LinearRamp, currentStartRate, currentEndRate, currentSteps)
		case "exponential":
			ramper = ramp.NewRampingService(ramp.ExponentialRamp, currentStartRate, currentEndRate, currentSteps)
		}
	}
	return currentRampDuration, ramper
}
