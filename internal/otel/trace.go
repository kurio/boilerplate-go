package otel

import (
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tracerProvider *sdktrace.TracerProvider

// NewStdoutSpanExporter initializes stdout span exporter
func NewStdoutSpanExporter() (exporter sdktrace.SpanExporter, err error) {
	exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		err = errors.Wrap(err, "error initializing stdout span exporter")
	}
	return
}

// InitTracerProvider initializes tracer provider
func InitTracerProvider(sampleRate float64, exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	if tracerProvider != nil {
		return tracerProvider
	}

	sampler := sdktrace.TraceIDRatioBased(sampleRate)

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider
}
