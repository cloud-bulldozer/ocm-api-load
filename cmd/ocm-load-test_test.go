package main

import (
	"testing"

	"github.com/spf13/viper"
)

func initConfigTests() {
	viper.Reset()
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigFile("test.yaml")
	viper.AutomaticEnv()
}
func Test_configAWS(t *testing.T) {

	t.Run("TestingFlags", func(t *testing.T) {
		initConfigTests()
		viper.Set("aws-region", "us-east-2")
		viper.Set("aws-access-key", "THISISAKEY")
		viper.Set("aws-access-secret", "THISISASECRET")
		viper.Set("aws-account-id", "THEID")
		configAWS()
		aws := viper.Get("aws").([]interface{})[0].(map[interface{}]interface{})
		if aws["region"] == "" {
			t.Fatalf("Flags are set and `aws.0.region` should be set")
		}
	})

	t.Run("TestingWithIncompleteFlags", func(t *testing.T) {
		initConfigTests()
		viper.Set("aws-region", "us-east-2")
		viper.Set("aws-access-secret", "THISISASECRET")
		viper.Set("aws-account-id", "THEID")
		err := configAWS()
		if err == nil {
			t.Fatalf("Shold fail because of missing AWS config")
		}
	})

	t.Run("TestingNoFlagsNoConfig", func(t *testing.T) {
		initConfigTests()
		err := configAWS()
		if err == nil {
			t.Fatalf("This test should fail because there is no configuraiton")
		}
	})

	t.Run("TestingNoFlagsWithConfig", func(t *testing.T) {
		initConfigTests()
		config := []interface{}{map[interface{}]interface{}{
			"region":            "aws-region",
			"access-key":        "aws-access-key",
			"secret-access-key": "aws-access-secret",
			"account-id":        "aws-account-id",
		}}
		viper.Set("aws", config)
		err := configAWS()
		if err != nil {
			t.Fatalf("This test should not fail because there is a YAML config")
		}
		aws := viper.Get("aws").([]interface{})[0].(map[interface{}]interface{})
		if aws["region"] == "" {
			t.Fatalf("Config is set and `aws.0.region` should be set")
		}
	})

	t.Run("TestingNoFlagsWithConfig", func(t *testing.T) {
		initConfigTests()
		config := []interface{}{map[interface{}]interface{}{
			"region":            "aws-region",
			"access-key":        "aws-access-key",
			"secret-access-key": "aws-access-secret",
			"account-id":        "aws-account-id",
		}, map[interface{}]interface{}{
			"region":            "aws-region-2",
			"access-key":        "aws-access-key-2",
			"secret-access-key": "aws-access-secret-2",
			"account-id":        "aws-account-id-2",
		}}
		viper.Set("aws", config)
		err := configAWS()
		if err == nil {
			t.Fatalf("This test should fail because there are multiple AWS configs")
		}
	})
}

func Test_configES(t *testing.T) {
	t.Run("TestWithFlags", func(t *testing.T) {
		initConfigTests()
		viper.Set("elastic-server", "http://localhost:9200")
		viper.Set("elastic-user", "user")
		viper.Set("elastic-insecure-skip-verify", false)
		viper.Set("elastic-password", "password")
		viper.Set("elastic-index", "es-index")
		if err := configES(); (err != nil) != false {
			if !viper.IsSet("elastic.server") {
				t.Fatalf("Flags are set and `elastic.server` should be set")
			}
		}
	})

	t.Run("TestingNoFlagsNoConfig", func(t *testing.T) {
		initConfigTests()
		err := configES()
		if err != nil {
			t.Fatalf("This test not should fail even with empty config")
		}
	})

	t.Run("TestingNoFlagsWithConfig", func(t *testing.T) {
		initConfigTests()
		config := map[string]interface{}{
			"server":   "https://localhost:9200",
			"user":     "user",
			"password": "password",
			"index":    "elastic-index",
		}
		viper.Set("elastic", config)
		err := configES()
		if err != nil {
			t.Fatalf("This test should not fail because there is a YAML config")
		}
		if viper.GetString("elastic.server") == "" {
			t.Fatalf("Config is set `elastic.server` should be set")
		}
	})

	t.Run("TestWithInclompleteFlags", func(t *testing.T) {
		initConfigTests()
		viper.Set("elastic-server", "http://localhost:9200")
		if err := configES(); (err != nil) != true {
			t.Fatalf("Should fail because of missing index.")
		}
	})
}
