package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/galecore/telemetry-example/internal/echohttp"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	if err := setupTelemetry(ctx, group); err != nil {
		panic(err)
	}

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	runServer(ctx, cfg, group)

	if err := group.Wait(); err != nil {
		panic(err)
	}
	slog.InfoContext(ctx, "shutdown success")
}

func runServer(ctx context.Context, cfg config, g *errgroup.Group) {
	echoServer := echohttp.NewServer()
	echoMux := echohttp.NewRouter(echoServer)

	httpServer := http.Server{
		Addr:        cfg.Addr,
		Handler:     echoMux,
		BaseContext: func(net.Listener) context.Context { return context.WithoutCancel(ctx) },
	}

	g.Go(func() error {
		<-ctx.Done()

		slog.InfoContext(ctx, "shutting down http server...")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*5)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		slog.InfoContext(ctx, "http server stopped gracefully")
		return nil
	})
}
