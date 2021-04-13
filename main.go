package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/tests"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/spf13/viper"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

var connAttacker func(*vegeta.Attacker)

var args struct {
	tokenURL   string
	configFile string
}

func init() {
	flag.StringVar(
		&args.tokenURL,
		"token-url",
		"https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token",
		"Token URL.",
	)
	flag.StringVar(
		&args.configFile,
		"config-file",
		"",
		"Test config file",
	)
}

func main() {
	flag.Parse()

	// set defaults
	duration := time.Duration(1) * time.Minute
	outputDirectory := "./results"

	// viper init
	viper.SetConfigName("ocm-api-load")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if args.configFile != "" {
		viper.SetConfigFile(args.configFile)
	}
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	// Validate config
	if !viper.InConfig("token") {
		log.Fatal("Token is a necessary configuration")
		os.Exit(1)
	}

	if !viper.InConfig("gateway-url") {
		log.Fatal("Gateway URL is a necessary configuration")
		os.Exit(1)
	}

	if viper.InConfig("duration-minutes") {
		duration = time.Duration(viper.GetInt("duration-minutes")) * time.Minute
	}

	if viper.InConfig("output-path") && viper.GetString("output-path") != "" {
		log.Printf("Using output directory: %s", viper.GetString("output-path"))
		outputDirectory = viper.GetString("output-path")
	}

	connection, err := sdk.NewConnectionBuilder().
		Insecure(true).
		URL(viper.GetString("gateway-url")).
		Client(viper.GetString("client.id"), viper.GetString("client.secret")).
		Tokens(viper.GetString("token")).
		TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
			return &helpers.CleanClustersTransport{Wrapped: wrapped}
		}).
		Build()
	if err != nil {
		fmt.Printf("Error creating api connection: %v", err)
		os.Exit(1)
	}
	defer helpers.Cleanup(connection)
	metrics := make(map[string]*vegeta.Metrics)

	connAttacker = vegeta.Client(&http.Client{Transport: connection})
	attacker := vegeta.NewAttacker(connAttacker)

	err = helpers.CreateFolder(outputDirectory)
	if err != nil {
		fmt.Printf("Error creating output directory: %s", err)
		os.Exit(1)
	}
	err = helpers.CreateFolder(path.Join(outputDirectory, helpers.ReportsDirectory))
	if err != nil {
		fmt.Printf("Error creating reports directory: %s", err)
		os.Exit(1)
	}

	if err := tests.Run(attacker, metrics, outputDirectory, duration, connection, viper.Sub("tests")); err != nil {
		fmt.Printf("Error running create cluster load test: %v", err)
		os.Exit(1)
	}
}
