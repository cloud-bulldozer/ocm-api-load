package tests

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	sdk "github.com/openshift-online/ocm-sdk-go"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func Run(
	attacker *vegeta.Attacker,
	metrics map[string]*vegeta.Metrics,
	outputDirectory string,
	duration time.Duration,
	connection *sdk.Connection,
	viper *viper.Viper) error {

	// testId provides a common value to associate all output data from running
	// the full test suite with a single test run.
	testID := uuid.NewV4().String()

	for _, t := range tests {

		// TODO: Create an Attacker for each individual test. This is due to the
		//       fact that vegeta (and compatible parsers, such as benchmark-wrapper)
		//       except the sequence to start at 0 for each result file. (Possibly a bug?)
		connAttacker := vegeta.Client(&http.Client{Transport: connection})
		attacker = vegeta.NewAttacker(connAttacker)

		// Bind "Test Harness"
		t.ID = testID
		t.OutputDirectory = outputDirectory
		t.ReportsDirectory = path.Join(outputDirectory, helpers.ReportsDirectory)
		t.Attacker = attacker
		t.Metrics = metrics
		t.Connection = connection
		t.Duration = duration
		t.Rate = helpers.DefaultRate

		// Go over the config file to find specifics for each test
		if viper.InConfig(t.TestName) {
			// Obtain frequency and per values
			freq := viper.GetInt(fmt.Sprintf("%s.freq", t.TestName))
			per := viper.GetString(fmt.Sprintf("%s.per", t.TestName))
			// To parse the duration we need to add the 1 to set a unit of i.
			// Eg.: 1s, 1m ,1h
			perDuration, err := time.ParseDuration("1" + per)
			if err != nil {
				log.Fatalf("parsing rate for test %s: %v%s", t.TestName, freq, per)
				return err
			}
			// Create the vegeta rate with the config values
			t.Rate = vegeta.Rate{Freq: freq, Per: perDuration}
			// Check for an override on the test duration
			dur := viper.GetString(fmt.Sprintf("%s.duration", t.TestName))
			if dur != "" {
				t.Duration, err = time.ParseDuration(dur)
				if err != nil {
					log.Fatalf("parsing duration override for test %s: %s", t.TestName, dur)
					return err
				}
			}
			key := fmt.Sprintf("%s.write-report", t.TestName)
			if viper.IsSet(key) {
				writeReport := viper.GetBool(key)
				t.WriteReport = writeReport
			}
		}

		// Execute the Test
		log.Printf("-------------------------------------------------------------------------------")
		log.Printf("Executing Test: %s", t.TestName)
		log.Printf("Rate: %s", t.Rate.String())
		log.Printf("Duration: %s", t.Duration.String())
		log.Printf("Path: %s", t.Path)
		log.Printf("Report: %v", t.WriteReport)
		err := t.Handler(&t)
		if err != nil {
			return err
		}
	}
	return nil
}
