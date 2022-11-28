package config

import (
	"time"

	"github.com/spf13/viper"
)

// Otel configuration
type Otel struct {
	Exporter Exporter
	Tracer   Tracer
	Metric   Metric
}

type Exporter struct {
	OTLPEndpoint string
}

type Tracer struct {
	SampleRate float64
}

type Metric struct {
	Interval time.Duration
}

func loadOtelConfig() Otel {
	viper.SetDefault("otel.exporter.otlp_endpoint", "127.0.0.1:4317")
	viper.SetDefault("otel.tracer.sample_rate", 1.0)
	viper.SetDefault("otel.metric.interval", 5)

	return Otel{
		Exporter: Exporter{
			OTLPEndpoint: viper.GetString("otel.exporter.otlp_endpoint"),
		},
		Tracer: Tracer{
			SampleRate: viper.GetFloat64("otel.tracer.sample_rate"),
		},
		Metric: Metric{
			Interval: time.Duration(viper.GetInt("otel.metric.interval")) * time.Second,
		},
	}
}
