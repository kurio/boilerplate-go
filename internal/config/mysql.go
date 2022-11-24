package config

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MySQL configuration
type MySQL struct {
	DSN             string
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

func loadMySQLConfig() MySQL {
	viper.SetDefault("mysql.conn_max_lifetime", 3600)
	viper.SetDefault("mysql.max_idle_connections", 10)
	viper.SetDefault("mysql.max_open_connections", 100)

	dsn := viper.GetString("mysql.dsn")
	if dsn == "" {
		logrus.Fatal("mysql.dsn is not set")
	}

	return MySQL{
		DSN:             dsn,
		ConnMaxLifetime: time.Duration(viper.GetInt("mysql.conn_max_lifetime")) * time.Second,
		MaxIdleConns:    viper.GetInt("mysql.max_idle_connections"),
		MaxOpenConns:    viper.GetInt("mysql.max_open_connections"),
	}
}
