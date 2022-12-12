package config

import (
	"github.com/spf13/viper"
)

// Config represents app configuration.
type Config struct {
	Debug       bool
	LogLevelStr string
	StatsdURL   string

	MySQL MySQL
	Mongo Mongo
	Redis Redis
	HTTP  HTTP

	Otel Otel
}

func LoadConfig() Config {
	var c Config

	viper.SetDefault("debug", false)
	viper.SetDefault("log.level", "info")

	c.Debug = viper.GetBool("debug")
	c.LogLevelStr = viper.GetString("log.level")
	c.StatsdURL = viper.GetString("statsd.url")

	c.MySQL = loadMySQLConfig()
	c.Mongo = loadMongoConfig()
	c.Redis = loadRedisConfig()
	c.HTTP = loadHTTPConfig()

	c.Otel = loadOtelConfig()

	return c
}
