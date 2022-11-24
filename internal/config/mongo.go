package config

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Mongo configuration
type Mongo struct {
	URI                    string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
}

func loadMongoConfig() Mongo {
	viper.SetDefault("mongo.connect_timeout_ms", 2000)
	viper.SetDefault("mongo.server_selection_timeout_ms", 3000)

	mongoUri := viper.GetString("mongo.uri")
	if mongoUri == "" {
		logrus.Fatal("mongo.uri is not set")
	}

	return Mongo{
		URI:                    mongoUri,
		ConnectTimeout:         time.Duration(viper.GetInt("mongo.connect_timeout_ms")) * time.Millisecond,
		ServerSelectionTimeout: time.Duration(viper.GetInt("mongo.server_selection_timeout_ms")) * time.Millisecond,
	}
}
