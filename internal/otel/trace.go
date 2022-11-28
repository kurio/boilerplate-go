package otel

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
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

// NewOTLPSpanExporter initializes OTLP span exporter. If
// unset, localhost:4317 will be used as a default.
func NewOTLPSpanExporter(endpoint string) (exporter sdktrace.SpanExporter, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exporter, err = otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		),
	)
	return
}

// InitTracerProvider initializes tracer provider
func InitTracerProvider(sampleRate float64, exporter sdktrace.SpanExporter, res *resource.Resource) *sdktrace.TracerProvider {
	if tracerProvider != nil {
		return tracerProvider
	}

	sampler := sdktrace.TraceIDRatioBased(sampleRate)

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider
}
