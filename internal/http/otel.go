package http

import (
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer
var tracerProvider *sdktrace.TracerProvider

func initTracer(name string) {
	tracer = otel.Tracer(name)
}

func initTracerProvider(name string) (*sdktrace.TracerProvider, error) {
	exporter, err := NewStdoutExporter()
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func NewSampler(fraction float64) sdktrace.Sampler {
	return sdktrace.TraceIDRatioBased(fraction)
}

func NewStdoutExporter() (sdktrace.SpanExporter, error) {
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	return exporter, err
}
