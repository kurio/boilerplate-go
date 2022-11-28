package main

import (
	"context"

	"github.com/kurio/boilerplate-go/internal/otel"
	"github.com/sirupsen/logrus"
	otelpkg "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func initOtel() {
	logrus.Debugf("Using otel agent address: %s", config.Otel.Exporter.OTLPEndpoint)
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
		),
	)
	if err != nil {
		logrus.Fatalf("Could not set otel resources: %+v", err)
	}

	/*****
	Tracer
	******/
	// TODO: change exporter
	// logrus.Debug("initializing new stdout span exporter")
	// spanExporter, err := otel.NewStdoutSpanExporter()
	logrus.Debug("initializing new OTLP span exporter...")
	spanExporter, err := otel.NewOTLPSpanExporter(config.Otel.Exporter.OTLPEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing span exporter: %+v", err)
	}

	logrus.Debug("initializing tracer provider...")
	tracerProvider = otel.InitTracerProvider(
		config.Otel.Tracer.SampleRate,
		spanExporter,
		resources)

	/*****
	Metric
	******/
	// TODO: change exporter
	// logrus.Debug("initializing new stdout metric exporter...")
	// metricExporter, err := otel.NewStdoutMetricExporter()
	logrus.Debug("initializing new OTLP metric exporter...")
	metricExporter, err := otel.NewOTLPMetricExporter(config.Otel.Exporter.OTLPEndpoint)
	if err != nil {
		logrus.Fatalf("error initializing metric exporter")
	}
	logrus.Debugf("initializing meter provider...")
	meterProvider = otel.InitMeterProvider(
		config.Otel.Metric.Interval,
		metricExporter,
		resources)
}