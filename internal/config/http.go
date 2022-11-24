package config

import (
	"time"

	"github.com/spf13/viper"
)

type httpServer struct {
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type httpClient struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

type HTTP struct {
	Server httpServer
	Client httpClient
}

func loadHTTPConfig() HTTP {
	viper.SetDefault("http.server.timeout_ms", 2000)
	viper.SetDefault("http.server.read_timeout_ms", 2000)
	viper.SetDefault("http.server.write_timeout_ms", 5000)

	viper.SetDefault("http.client.timeout_ms", 3000)
	viper.SetDefault("http.client.max_idle_conns", 100)
	viper.SetDefault("http.client.max_idle_conns_per_host", 2)
	viper.SetDefault("http.client.idle_conn_timeout", 90)

	return HTTP{
		Server: httpServer{
			Timeout:      time.Duration(viper.GetInt("http.server.timeout_ms")) * time.Millisecond,
			ReadTimeout:  time.Duration(viper.GetInt("http.server.read_timeout_ms")) * time.Millisecond,
			WriteTimeout: time.Duration(viper.GetInt("http.server.write_timeout_ms")) * time.Millisecond,
		},
		Client: httpClient{
			Timeout:             time.Duration(viper.GetInt("http.client.timeout_ms")) * time.Millisecond,
			MaxIdleConns:        viper.GetInt("http.client.max_idle_conns"),
			MaxIdleConnsPerHost: viper.GetInt("http.client.max_idle_conns_per_host"),
			IdleConnTimeout:     time.Duration(viper.GetInt("http.client.idle_conn_timeout")) * time.Second,
		},
	}
}
