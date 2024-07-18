package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/galecore/telemetry-example/internal/logs"
	"github.com/galecore/telemetry-example/internal/metrics"
	"github.com/galecore/telemetry-example/internal/tracing"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"golang.org/x/sync/errgroup"
)

func setupTelemetry(ctx context.Context, g *errgroup.Group) error {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.ErrorContext(ctx, "otel error", slog.Any("error", err))
	}))

	if err := setupLogger(ctx, g); err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}
	if err := setupTraces(ctx, g); err != nil {
		return fmt.Errorf("failed to setup traces: %w", err)
	}
	if err := setupMetrics(ctx, g); err != nil {
		return fmt.Errorf("failed to setup metrics: %w", err)
	}
	return nil
}

func setupLogger(ctx context.Context, g *errgroup.Group) error {
	logExporter, err := logs.NewExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new logs exporter: %w", err)
	}
	logProvider, err := logs.NewLoggerProvider(logExporter)
	if err != nil {
		return fmt.Errorf("failed to create new logger provider: %w", err)
	}

	// this is experimental in v0.4.0 otel log sdk, and will be migrated to go.opentelemetry.io/otel when stable.
	global.SetLoggerProvider(logProvider)
	//otel.SetLoggerProvider(logProvider) // remove call to global and uncomment this when the otel log sdk is stable
	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*5)
		defer cancel()
		return logProvider.Shutdown(shutdownCtx)
	})

	logger := slog.New(logs.SlogFanout(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		otelslog.NewHandler(logs.ScopeName),
	))
	slog.SetDefault(logger)
	return nil
}

func setupTraces(ctx context.Context, g *errgroup.Group) error {
	traceExporter, err := tracing.NewExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new trace exporter: %w", err)
	}
	tracerProvider, err := tracing.NewTracerProvider(traceExporter)
	if err != nil {
		return fmt.Errorf("failed to create new tracer provider: %w", err)
	}
	otel.SetTracerProvider(tracerProvider)
	// propagators are used to extract and inject incoming and outgoing contexts with trace and span data
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*5)
		defer cancel()
		return tracerProvider.Shutdown(shutdownCtx)
	})
	return nil
}

func setupMetrics(ctx context.Context, g *errgroup.Group) error {
	meterProvider, err := metrics.NewPushMeterProvider(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new push meter provider: %w", err)
	}
	otel.SetMeterProvider(meterProvider)
	g.Go(func() error {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*5)
		defer cancel()
		return meterProvider.Shutdown(shutdownCtx)
	})
	return nil
}
