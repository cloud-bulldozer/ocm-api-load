package tests

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nimrodshn/cs-load-test/pkg/helpers"
	"github.com/nimrodshn/cs-load-test/pkg/result"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	defaultRate = 100
)

func TestSelfAccessToken(
	attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {

	testName := "self-access-token"
	fileName := fmt.Sprintf("%s_%s", testID, testName)

	log.Printf("Executing Test: %s", testName)

	target := vegeta.Target{
		Method: http.MethodPost,
		URL:    helpers.SelfAccessTokenEndpoint,
	}

	targeter := vegeta.NewStaticTargeter(target)
	resultFile, err := createFile(fileName, outputDirectory)
	defer resultFile.Close()
	if err != nil {
		return err
	}

	// Display some info about the test being ran to catch obvious issues
	// and include context
	log.Printf("Test: %s\n", testName)
	log.Printf("Output File: %s/%s\n", outputDirectory, fileName)

	metrics[testName] = new(vegeta.Metrics)
	defer metrics[testName].Close()

	for res := range attacker.Attack(targeter, rate, duration, testName) {
		metrics[testName].Add(res)
		result.Write(res, resultFile)
	}

	return nil
}

func TestListSubscriptions(attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {

	testName := "list-subscriptions"
	fileName := fmt.Sprintf("%s_%s", testID, testName)

	log.Printf("Executing Test: %s", testName)

	target := vegeta.Target{
		Method: http.MethodGet,
		URL:    helpers.ListSubscriptionsEndpoint,
	}

	targeter := vegeta.NewStaticTargeter(target)
	resultFile, err := createFile(fileName, outputDirectory)
	defer resultFile.Close()
	if err != nil {
		return err
	}

	// Display some info about the test being ran to catch obvious issues
	// and include context
	fmt.Printf("Test: %s\n", testName)
	fmt.Printf("Output File: %s/%s\n", outputDirectory, fileName)

	metrics[testName] = new(vegeta.Metrics)
	defer metrics[testName].Close()

	for res := range attacker.Attack(targeter, rate, duration, testName) {
		result.Write(res, resultFile)
	}

	return nil
}

func TestAccessReview(attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {

	testName := "access-review"
	fileName := fmt.Sprintf("%s_%s", testID, testName)

	log.Printf("Executing Test: %s", testName)

	target := vegeta.Target{
		Method: http.MethodPost,
		URL:    helpers.AccessReviewEndpoint,
		Body:   []byte("{\"account_username\": \"rhn-support-tiwillia\", \"action\": \"get\", \"resource_type\": \"Subscription\"}"),
	}

	targeter := vegeta.NewStaticTargeter(target)
	resultFile, err := createFile(fileName, outputDirectory)
	defer resultFile.Close()
	if err != nil {
		return err
	}

	// Display some info about the test being ran to catch obvious issues
	// and include contextq
	fmt.Printf("Test: %s\n", testName)
	fmt.Printf("Output File: %s/%s\n", outputDirectory, fileName)

	metrics[testName] = new(vegeta.Metrics)
	defer metrics[testName].Close()

	for res := range attacker.Attack(targeter, rate, duration, testName) {
		result.Write(res, resultFile)
	}

	return nil

}

func TestRegisterNewCluster(attacker *vegeta.Attacker,
	testID string,
	metrics map[string]*vegeta.Metrics,
	rate vegeta.Pacer,
	outputDirectory string,
	duration time.Duration) error {

	testName := "new-cluster-registration"
	fileName := fmt.Sprintf("%s_%s", testID, testName)

	log.Printf("Executing Test: %s", testName)

	// TODO: Generate a UUID for each Request
	// TODO: The authorization_token should be real. Not sure what to set it as, though.
	target := vegeta.Target{
		Method: http.MethodPost,
		URL:    helpers.ClusterRegistrationEndpoint,
		Body:   []byte("{\"authorization_token\": \"specify-me\", \"cluster_id\": \"c98550e5-1c9f-47bb-b46f-b2b6e7befeb3\"}"),
	}

	targeter := vegeta.NewStaticTargeter(target)
	resultFile, err := createFile(fileName, outputDirectory)
	defer resultFile.Close()
	if err != nil {
		return err
	}

	// Display some info about the test being ran to catch obvious issues
	// and include contextq
	fmt.Printf("Test: %s\n", testName)
	fmt.Printf("Output File: %s/%s\n", outputDirectory, fileName)

	metrics[testName] = new(vegeta.Metrics)
	defer metrics[testName].Close()

	for res := range attacker.Attack(targeter, rate, duration, testName) {
		result.Write(res, resultFile)
	}

	return nil

}
