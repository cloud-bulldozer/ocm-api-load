package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nimrodshn/cs-load-test/pkg/helpers"
	"github.com/nimrodshn/cs-load-test/pkg/tests"
	sdk "github.com/openshift-online/ocm-sdk-go"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var connection *sdk.Connection
var connAttacker func(*vegeta.Attacker)
var rate vegeta.Rate
var duration time.Duration

var args struct {
	load            bool
	tokenURL        string
	gatewayURL      string
	clientID        string
	clientSecret    string
	token           string
	durationInMin   int
	rate            int
	outputDirectory string
}

func init() {
	flag.BoolVar(
		&args.load,
		"load",
		false,
		"Run load tests.",
	)
	flag.StringVar(
		&args.tokenURL,
		"token-url",
		"https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token",
		"Token URL.",
	)
	flag.StringVar(
		&args.gatewayURL,
		"gateway-url",
		"http://localhost:8000",
		"Gateway URL.",
	)
	flag.StringVar(
		&args.clientID,
		"client-id",
		"cloud-services",
		"OpenID client identifier.",
	)
	flag.StringVar(
		&args.clientSecret,
		"client-secret",
		"",
		"OpenID client secret.",
	)
	flag.StringVar(
		&args.token,
		"token",
		"",
		"Offline token for authentication.",
	)
	flag.IntVar(
		&args.durationInMin,
		"duration-in-min",
		1,
		"How long should the load test take.",
	)
	flag.IntVar(
		&args.rate,
		"rate",
		5,
		"How many times (per second) should the endpoint be hit.",
	)
	flag.StringVar(
		&args.outputDirectory,
		"output-path",
		"",
		"path to output results.",
	)
}

func main() {
	flag.Parse()

	// Consider passing a yaml map from test-name to rate as config file to allow more flexability.

	connection, err := sdk.NewConnectionBuilder().
		Insecure(true).
		URL(args.gatewayURL).
		Client(args.clientID, args.clientSecret).
		Tokens(args.token).
		TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
			return &helpers.CleanClustersTransport{Wrapped: wrapped}
		}).
		Build()
	if err != nil {
		fmt.Printf("Error creating api connection: %v", err)
		os.Exit(1)
	}
	defer helpers.Cleanup(connection)
	attacker := new(vegeta.Attacker)
	metrics := make(map[string]*vegeta.Metrics)

	rate = vegeta.Rate{Freq: args.rate, Per: time.Second}
	duration = time.Duration(args.durationInMin) * time.Minute
	connAttacker = vegeta.Client(&http.Client{Transport: connection})
	attacker = vegeta.NewAttacker(connAttacker)

	if err := tests.Run(attacker, metrics, rate, args.outputDirectory, duration); err != nil {
		fmt.Printf("Error running create cluster load test: %v", err)
		os.Exit(1)
	}
}
