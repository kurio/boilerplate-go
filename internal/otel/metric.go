package otel

import (
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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

// InitMeterProvider initializes meter provider
func InitMeterProvider(interval time.Duration, exporter sdkmetric.Exporter) *sdkmetric.MeterProvider {
	if meterProvider != nil {
		return meterProvider
	}

	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(interval),
	)

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
	)

	global.SetMeterProvider(meterProvider)
	return meterProvider
}
