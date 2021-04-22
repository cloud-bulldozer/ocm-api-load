package tests

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	sdk "github.com/openshift-online/ocm-sdk-go"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func Run(
	outputDirectory string,
	duration time.Duration,
	connection *sdk.Connection,
	viper *viper.Viper) error {

	// testId provides a common value to associate all output data from running
	// the full test suite with a single test run.
	testID := uuid.NewV4().String()

	for _, t := range tests {

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
		t.Duration = duration
		t.Rate = helpers.DefaultRate
		t.Encoder = &encoder

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
		}

		log.Printf("Executing Test: %s", t.TestName)
		log.Printf("Rate: %s", t.Rate.String())
		log.Printf("Duration: %s", t.Duration.String())
		log.Printf("Endpoint: %s", t.Path)
		err = t.Handler(&t)
		if err != nil {
			return err
		}

		// Cleanup (cannot defer as it must happen for each test in the loop)
		log.Printf("Results written to: %s", fileName)
		err = resultsFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
