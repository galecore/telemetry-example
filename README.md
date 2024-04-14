# Telemetry example via OpenTelemetry and Grafana Cloud

This is a basic telemetry example for exporting golang telemetry to some grafana cloud instance. 

This example features:
- A simple echo golang http server
- A one-shot client program, that fires 5 requests and then exits
- An OTEL Collector, that accumulates all telemetry alongside the apps before pushing it into Grafana Cloud

Golang OTEL library currently provides support for Traces and Metrics. Logging support is not provided. 

Therefore, the example provides an internal/metrics and internal/tracing packages with some commentary 
regarding the usage of Traces and Metrics in OTEL. Metrics and Traces are exported to the Collector via OTLP.

Logging is done via log/slog, a unified structured logging interface recently added to the standard library.
In order to export everything through OTEL collector, the collector exposes a loki receiver, 
and a loki slog handler is used to push data into it. When logs support becomes stable in OTEL, 
slog would hopefully get a handler that would export logs into the Collector in OTLP format also.

OTEL Collector is used to not bother with telemetry export security directly in the applications.
In production such a Collector would probably be privately available somewhere near the application,
and would manage additional telemetry augmentation and processing, as well as security - auth, certs, etc.

In this example, OTEL also adds host metrics to the exported metrics data, allowing for infra resource tracking. 

To not lose any unexported telemetry before finishing, both apps have basic graceful shutdown logic implemented. 

The config for OpenTelemetry Collector is mostly generated by the Grafana Cloud.
Consult [this link in grafana docs](https://grafana.com/docs/grafana-cloud/monitor-applications/application-observability/setup/collector/opentelemetry-collector/#application-observability-with-opentelemetry-collector) to get your version.
