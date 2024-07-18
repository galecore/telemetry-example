package logs

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// ScopeName is not given by the otelslog bridge, for some reason
// todo: resolve this in https://github.com/open-telemetry/opentelemetry-go-contrib/issues/5927
const ScopeName = "go.opentelemetry.io/contrib/bridges/otelslog"

func NewExporter(ctx context.Context) (*otlploggrpc.Exporter, error) {
	/*
		There are tons of configuration options for OTLP exporter. They can all be set via environment variables.
		Mainly: OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_INSECURE, and many more.
	*/
	return otlploggrpc.New(ctx)
}

func NewLoggerProvider(exporter log.Exporter) (*log.LoggerProvider, error) {
	/*
			LoggerProvider is a factory for Loggers.

			In order to create a LoggerProvider, you need to state several other entities:
			- kind of mandatory things that devs want to set, even though you can create a LoggerProvider without them:
				- Exporter, which defines where the logs should be sent
			    - BatchProcessor, which is used to batch logs before sending them
				- Resource, which defines the base attributes for all logs
					- obvious stuff like opentelemetry sdk version that created the log
					- more app-oriented stuff like service name
			- other things:
				- Different LogProcessors could be set here, which give you access to the created records
					- Custom LogProcessors usually add some filtration rules or additional attributes to logs
					- In reality, when an exporter is given to the LoggerProvider, it is just a kind of LogProcessor
					- That means that exporter should be placed at the end, as they are called in order of registration

			Logger is used to log records.

			This is a "direct-to-exporter" setup, which resembles the logic of other OTel components.
		    At the same time, in some cases (for example apps where performance is most crucial),
			it might be more beneficial to export logs directly to files or stdout,
			and process them in the collector separately.
			In such cases, otel would not be used in the app logging part.
	*/

	r := resource.Default()
	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithResource(r),
		log.WithProcessor(processor),
	)
	return provider, nil
}
