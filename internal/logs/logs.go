package logs

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/grafana/loki-client-go/loki"
	slogloki "github.com/samber/slog-loki/v3"
	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/sync/errgroup"
)

type lokiConfig struct {
	Endpoint string `env:"LOKI_PUSH_ENDPOINT"`
	TenantID string `env:"LOKI_TENANT_ID"`
}

func NewLokiClientFromEnv() (*loki.Client, error) {
	var cfg lokiConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load loki config: %w", err)
	}
	config, err := loki.NewDefaultConfig(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create loki config: %w", err)
	}
	config.TenantID = cfg.TenantID

	client, err := loki.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create loki client: %w", err)
	}
	return client, nil
}

func NewLogger(ctx context.Context, g *errgroup.Group, options *slog.HandlerOptions) (*slog.Logger, error) {
	lokiClient, err := NewLokiClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create loki client: %w", err)
	}

	g.Go(func() error {
		<-ctx.Done()
		lokiClient.Stop()
		return nil
	})

	return slog.New(slogmulti.Fanout(
		slog.NewTextHandler(os.Stdout, options),
		slogloki.Option{
			Level:       options.Level,
			Client:      lokiClient,
			Converter:   slogloki.DefaultConverter,
			AddSource:   options.AddSource,
			ReplaceAttr: options.ReplaceAttr,
		}.NewLokiHandler(),
	)), nil
}
