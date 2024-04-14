package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/galecore/telemetry-example/internal/echohttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)

	if err := setupTelemetry(ctx, group); err != nil {
		panic(err)
	}

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	client := echohttp.New(cfg.Endpoint, time.Second*5)
	for i := 0; i < 5; i++ {
		slog.InfoContext(ctx, "sending echo request")
		response, err := client.Echo(ctx, fmt.Sprintf("sending %d message", i+1))
		if err != nil {
			slog.ErrorContext(ctx, "got bad echo response", slog.Any("error", err))
		} else {
			slog.InfoContext(ctx, "got echo response", slog.String("response", response))
		}
	}

	cancel()
	if err := group.Wait(); err != nil {
		panic(err)
	}
	slog.InfoContext(ctx, "graceful shutdown success")
}
