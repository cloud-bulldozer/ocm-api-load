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
		{"1", "NoName", "testName", "NoName"},
		{"2", "NoName", "test_name", "MyTest"},
		{"3", "5/s", "rate", "5/s"},
		{"4", "5/s", "MyTest.rate", "2/s"},
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
		{"1", 15, "testDuration", 15},
		{"2", 15, "test_duration", 33},
		{"3", 4, "duration", 4},
		{"4", 4, "MyTest.duration", 99},
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
		{"1", 1, 5, 2, true},
		{"2", 1, 5, 1, false},
		{"3", 1, 5, 0, false},
		{"4", 0, 5, 2, false},
		{"4", 1, 0, 2, false},
		{"4", 5, 5, 2, false},
		{"4", 5, 2, 2, false},
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
			if got := c.ValidateRampConfig(context.TODO(), tt.max, tt.min, tt.steps); got != tt.want {
				t.Errorf("ConfigHelper.ValidateRampConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
