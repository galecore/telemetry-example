package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func NewPushReader(ctx context.Context) (*sdkmetric.PeriodicReader, error) {
	/*
		There are tons of configuration options for OTLP exporter. They can all be set via environment variables.
		Mainly: OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_INSECURE, and many more.
	*/

	exporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	/*
		If we are not using prometheus exporter, we need to not forget to register go runtime stats manually.
		This is done in NewPushMeterProvider() function.
	*/

	return sdkmetric.NewPeriodicReader(
		exporter,
		// sdkmetric.WithInterval() is set to 1m by default and can be configured via OTEL_METRIC_EXPORT_INTERVAL
		// sdkmetric.WithTimeout() is set to 30s by default and can be configured via OTEL_METRIC_EXPORT_TIMEOUT
	), nil
}

func NewPullReader() (*prometheus.Exporter, error) {
	/*
		With prom exporter you usually get go runtime stats out of the box, as
		default prom registry already has them registered and scheduled for collection.
	*/
	return prometheus.New() // could be configured with options, mainly .WithNamespace("...") and .WithRegisterer
}

func NewMeterProvider(reader sdkmetric.Reader) *sdkmetric.MeterProvider {
	/*
		MeterProvider is a factory for Meters.

		In order to create a MeterProvider, you need to supply several entities:
		- mandatory in a real app:
			- a Reader+Exporter pair, which defines the pull/push model of metrics export
				- Push model uses PeriodicReader and OTLP exporters
				- Pull model uses ManualReader and some exposed http endpoint (e.g. promhttp.Handler)
		- other things:
			- Meter Views, that allow to rename, filter, aggregate and overall modify the exported metrics output
				- Views are a new idea in go metrics, go prometheus lib doesn't have them

		Meter is used to create specific instruments - Counters, Gauges, and Histograms.
		Meters are named to show the instrumentation scope inside the app.
		Exported metric does not have the name of the meter it was created by.
	*/
	r := resource.Default()
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(r),    // if not resource is given, resource.Default() would be called
		sdkmetric.WithReader(reader), // if no reader is given, no metrics are exported
	)
}

func NewPushMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	reader, err := NewPushReader(ctx)
	if err != nil {
		return nil, err
	}
	meterProvider := NewMeterProvider(reader)

	// as said before, without prometheus exporter, runtime metrics should be registered manually
	err = runtime.Start(
		runtime.WithMeterProvider(meterProvider), // if no meter provider is set, the globally set one is used
		//runtime.WithMinimumReadMemStatsInterval() could be set, default interval is 15s. The call is quite expensive
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %w", err)
	}

	return meterProvider, nil
}
