package config

import (
	"context"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
	"github.com/spf13/viper"
)

type ConfigHelper struct {
	logger logging.Logger
	conf   *viper.Viper
}

func NewConfigHelper(logger logging.Logger, conf *viper.Viper) *ConfigHelper {
	return &ConfigHelper{
		logger: logger,
		conf:   conf,
	}
}

func (c *ConfigHelper) ResolveStringConfig(ctx context.Context, def, key string) string {
	s := c.conf.GetString(key)
	if s == "" {
		c.logger.Info(ctx, "no value for %s. Using default value.", key)
		return def
	}
	return s
}

func (c *ConfigHelper) ResolveIntConfig(ctx context.Context, def int, key string) int {
	i := c.conf.GetInt(key)
	if i == 0 {
		c.logger.Info(ctx, "no value for %s. Using default value.", key)
		return def
	}
	return i
}

func (c *ConfigHelper) ValidateRampConfig(ctx context.Context, max, min, steps int) bool {
	if steps < 2 {
		c.logger.Warn(ctx,
			"steps must be always 2 or more. Ignoring ramping configuration.")
		return false
	}
	if min < 1 {
		c.logger.Warn(ctx,
			"min rate must be always 1 or more. Ignoring ramping configuration.")
		return false
	}
	if max < 1 {
		c.logger.Warn(ctx,
			"max rate must be always 1 or more. Ignoring ramping configuration.")
		return false
	}
	if max <= min {
		c.logger.Warn(ctx,
			"max rate must be bigger than min rate. Ignoring ramping configuration.")
		return false
	}
	return true
}
