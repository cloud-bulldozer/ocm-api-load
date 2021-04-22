package handlers

import (
	"fmt"
	"log"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/report"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/result"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func TestStaticEndpoint(options *helpers.TestOptions) error {

	testName := options.TestName
	// Vegeta Results File
	fileName := fmt.Sprintf("%s_%s.json", options.ID, testName)
	resultFile, err := helpers.CreateFile(fileName, options.OutputDirectory)
	if err != nil {
		return err
	}
	defer resultFile.Close()

	// Specify the HTTP request(s) that will be executed
	target := vegeta.Target{
		Method: options.Method,
		URL:    options.Path,
	}
	if len(options.Body) > 0 {
		target.Body = options.Body
	}
	targeter := vegeta.NewStaticTargeter(target)

	// Create a Metrics bucket for this test run
	options.Metrics[testName] = new(vegeta.Metrics)
	defer options.Metrics[testName].Close()

	// Experimental:

	// Create the File to Write the Results to
	fileName2 := fmt.Sprintf("%s_%s.gob", options.ID, testName)
	out, err := helpers.CreateFile(fileName2, options.OutputDirectory)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create an encoder
	enc := vegeta.NewEncoder(out)

	// Execute the HTTP Requests; repeating as needed to meet the specified duration
	for res := range options.Attacker.Attack(targeter, options.Rate, options.Duration, options.TestName) {
		result.Write(res, resultFile)
		options.Metrics[testName].Add(res)
		enc.Encode(res)
	}

	log.Printf("Results written to: %s", fileName)

	if options.WriteReport {
		err = report.Write(fmt.Sprintf("%s_%s-report", options.ID, options.TestName), options.ReportsDirectory, options.Metrics[testName])
		if err != nil {
			return err
		}
	}

	return nil

}
