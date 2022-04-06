package config

import (
	"context"
	"testing"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/spf13/viper"
)

func TestConfigHelper_ResolveStringConfig(t *testing.T) {
	tests := []struct {
		name string
		def  string
		key  string
		want string
	}{
		{"no_key_use_def", "NoName", "testName", "NoName"},
		{"key_exists", "NoName", "test_name", "MyTest"},
		{"nested_key_exists", "5/s", "MyTest.rate", "2/s"},
	}
	conf := viper.New()
	conf.Set("test_name", "MyTest")
	conf.Set("MyTest", map[string]string{})
	conf.Set("MyTest.rate", "2/s")
	logBuilder := logging.NewGoLoggerBuilder()
	log, _ := logBuilder.Build()
	c := NewConfigHelper(log, conf)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.ResolveStringConfig(context.TODO(), tt.def, tt.key); got != tt.want {
				t.Errorf("ConfigHelper.ResolveStringConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigHelper_ResolveIntConfig(t *testing.T) {
	tests := []struct {
		name string
		def  int
		key  string
		want int
	}{
		{"no_key_use_def", 15, "testDuration", 15},
		{"key_exists", 15, "test_duration", 33},
		{"nested_key_exists", 4, "MyTest.duration", 99},
	}
	conf := viper.New()
	conf.Set("test_duration", 33)
	conf.Set("MyTest", map[string]interface{}{})
	conf.Set("MyTest.duration", 99)
	logBuilder := logging.NewGoLoggerBuilder()
	log, _ := logBuilder.Build()
	c := NewConfigHelper(log, conf)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.ResolveIntConfig(context.TODO(), tt.def, tt.key); got != tt.want {
				t.Errorf("ConfigHelper.ResolveIntConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigHelper_ValidateRampConfig(t *testing.T) {
	tests := []struct {
		name  string
		min   int
		max   int
		steps int
		want  bool
	}{
		{"correct_config", 1, 5, 2, true},
		{"steps_less_than_two", 1, 5, 1, false},
		{"min_lower_than_one", 0, 5, 2, false},
		{"max_equals_min", 5, 5, 2, false},
		{"max_lower_than_min", 5, 2, 2, false},
	}
	conf := viper.New()
	conf.Set("test_duration", 33)
	conf.Set("MyTest", map[string]interface{}{})
	conf.Set("MyTest.duration", 99)
	logBuilder := logging.NewGoLoggerBuilder()
	log, _ := logBuilder.Build()
	c := &ConfigHelper{
		logger: log,
		conf:   conf,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.ValidateRampConfig(context.TODO(), tt.min, tt.max, tt.steps); got != tt.want {
				t.Errorf("ConfigHelper.ValidateRampConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
