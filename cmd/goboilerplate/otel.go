package main

import (
	"context"

	"github.com/sirupsen/logrus"
	otelpkg "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"

	"github.com/kurio/boilerplate-go/internal/otel"
)

var (
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
)

func initOtel() {
	if !config.Debug {
		// ignore error on production mode
		otelpkg.SetErrorHandler(otelpkg.ErrorHandlerFunc(func(err error) {}))
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(app),
			semconv.ServiceVersionKey.String(gitCommit),
		),
	)
	if err != nil {
		logrus.Fatalf("Could not set otel resources: %+v", err)
	}

	if config.Otel.Exporter.OTLPEndpoint == "" {
		return
	} else {
		logrus.Debugf("Using otel agent address: %s", config.Otel.Exporter.OTLPEndpoint)
	}

	/*****
	Tracer
	******/
	// TODO: change exporter
	// logrus.Debug("Initializing new stdout span exporter")
	// spanExporter, err := otel.NewStdoutSpanExporter()
	logrus.Debug("Initializing new OTLP span exporter...")
	spanExporter, err := otel.NewOTLPSpanExporter(config.Otel.Exporter.OTLPEndpoint)
	if err != nil {
		logrus.Errorf("Error initializing span exporter: %+v", err)
	} else {
		logrus.Debug("Initializing tracer provider...")
		tracerProvider = otel.InitTracerProvider(
			config.Otel.Tracer.SampleRate,
			spanExporter,
			resources)
	}

	/*****
	Metric
	******/
	// TODO: change exporter
	// logrus.Debug("Initializing new stdout metric exporter...")
	// metricExporter, err := otel.NewStdoutMetricExporter()
	logrus.Debug("Initializing new OTLP metric exporter...")
	metricExporter, err := otel.NewOTLPMetricExporter(config.Otel.Exporter.OTLPEndpoint)
	if err != nil {
		logrus.Errorf("Error initializing metric exporter: %+v", err)
	} else {
		logrus.Debugf("Initializing meter provider...")
		meterProvider = otel.InitMeterProvider(
			config.Otel.Metric.Interval,
			metricExporter,
			resources)
	}
}
