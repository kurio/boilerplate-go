package otel

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var meterProvider *sdkmetric.MeterProvider

// NewStdoutMetricExporter initializes stdout metric exporter
func NewStdoutMetricExporter() (exporter sdkmetric.Exporter, err error) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	exporter, err = stdoutmetric.New(stdoutmetric.WithEncoder(enc))
	if err != nil {
		err = errors.Wrap(err, "error initializing stdout metric exporter")
	}
	return
}

// NewOTLPMetricExporter initializes OTLP metric exporter. If
// unset, localhost:4317 will be used as a default.
func NewOTLPMetricExporter(endpoint string) (exporter sdkmetric.Exporter, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exporter, err = otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		err = errors.Wrap(err, "error initializing otlp metric exporter")
	}
	return
}

// InitMeterProvider initializes meter provider
func InitMeterProvider(interval time.Duration, exporter sdkmetric.Exporter, res *resource.Resource) *sdkmetric.MeterProvider {
	if meterProvider != nil {
		return meterProvider
	}

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(interval),
	)

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	global.SetMeterProvider(meterProvider)
	return meterProvider
}
