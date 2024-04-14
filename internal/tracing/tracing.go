package tracing

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	/*
		There are tons of configuration options for OTLP exporter. They can all be set via environment variables.
		Mainly: OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_INSECURE, and many more.
	*/
	return otlptracegrpc.New(ctx)
}

func NewTracerProvider(exporter sdktrace.SpanExporter) (*sdktrace.TracerProvider, error) {
	/*
		TracerProvider is a factory for Tracers.

		In order to create a TracerProvider, you need to state several other entities:
		- kind of mandatory things that devs want to set, even though you can create a TracerProvider without them:
			- Exporter, which defines where the traces should be sent
				- Both push and pull models for exporters could be used
				- In DI it would be most sensible to aggregate an exporter from outside and not compose it here
				- Trace exporters mainly use push model
				- OTEL heavily pushes OTLP exporter protocol to be used in combination with their local OTEL Collector
				- Collector is an agent (e.g. yandex push-client) running alongside that pushes telemetry further
			- Sampler, which defines which portion of traces should be sampled
			- Resource, which defines the base attributes for all spans
				- obvious stuff like opentelemetry sdk version that created the span
				- more app-oriented stuff like service name
		- other things:
			- Different SpanProcessors could be set here, which give you access to the created spans
				- Custom SpanProcessors usually add some filtration rules or additional attributes to spans
				- In reality, when an exporter is given to the TracerProvider, it is just a kind of SpanProcessor
				- That means that exporter should be placed at the end, as they are called in order of registration

		Tracer is used to start Spans.

		Spans allow to record events, attributes and other data.
	*/

	// An interesting thing about the resource.Default()
	// It actually also loads attributes from env variables
	// mainly ServiceName via OTEL_SERVICE_NAME, but also
	// any other attributes in kv pairs OTEL_RESOURCE_ATTRIBUTES
	// in format key1=val1,key2=val2
	r := resource.Default()

	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(r),                      // if no resource is given, resource.Default() would be called
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // if no sampler is given, AlwaysSample would be used
		// ... SpanProcessors could be added here ...

		// Exporter SpanProcessor is usually the last processor to be set
		// note: exporter is always wrapped with either BatchSpanProcessor or SimpleSpanProcessor
		// BatchSpanProcessor batches completed spans before sending them (should be used in 99.9% cases)
		// SimpleSpanProcessor sends them upon completion immediately
		//	- this is actually useful in FaaS and other one shot tasks, but not in general
		sdktrace.WithBatcher(exporter),
	), nil
}
