package config

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Redis configuration
type Redis struct {
	Address             string
	Cluster             bool
	DB                  int
	DialTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	PoolSize            int
	PoolTimeout         time.Duration
	ConnMaxIdleTime     time.Duration
	ConnMaxLifetime     time.Duration
	ShortExpirationTime time.Duration
	LongExpirationTime  time.Duration
}

func loadRedisConfig() Redis {
	viper.SetDefault("redis.cluster", false)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.dial_timeout_ms", 5000)
	viper.SetDefault("redis.read_timeout_ms", 3000)
	viper.SetDefault("redis.write_timeout_ms", 3000)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.pool_timeout", 4)
	viper.SetDefault("redis.conn_max_idle_time", 300)

	addr := viper.GetString("redis.address")
	if addr == "" {
		logrus.Fatal("redis.address is not set")
	}

	conf := Redis{
		Address:      addr,
		Cluster:      viper.GetBool("redis.cluster"),
		DB:           viper.GetInt("redis.db"),
		DialTimeout:  time.Duration(viper.GetInt("redis.dial_timeout_ms")) * time.Millisecond,
		ReadTimeout:  time.Duration(viper.GetInt("redis.read_timeout_ms")) * time.Millisecond,
		WriteTimeout: time.Duration(viper.GetInt("redis.write_timeout_ms")) * time.Millisecond,
		PoolSize:     viper.GetInt("redis.pool_size"),
		PoolTimeout:  time.Duration(viper.GetInt("redis.pool_timeout")) * time.Millisecond,
	}

	connMaxIdleTime := viper.GetInt("redis.conn_max_idle_time")
	if connMaxIdleTime > 0 {
		conf.ConnMaxIdleTime = time.Duration(connMaxIdleTime) * time.Second
	}

	connMaxLifetime := viper.GetInt("redis.conn_max_lifetime")
	if connMaxLifetime > 0 {
		conf.ConnMaxLifetime = time.Duration(connMaxLifetime) * time.Second
	}

	shortExpirationTime := viper.GetInt("redis.short_expiration_time")
	if shortExpirationTime > 0 {
		conf.ShortExpirationTime = time.Duration(shortExpirationTime) * time.Second
	}
	longExpirationTime := viper.GetInt("redis.long_expiration_time")
	if longExpirationTime > 0 {
		conf.LongExpirationTime = time.Duration(longExpirationTime) * time.Second
	}

	return conf
}
