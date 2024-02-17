package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/itimofeev/word-of-wisdom/internal/app/client"
	signal "github.com/itimofeev/word-of-wisdom/pkg/context"
	"github.com/itimofeev/word-of-wisdom/pkg/pow"
	"github.com/itimofeev/word-of-wisdom/pkg/transport"
)

type config struct {
	ServerAddress string        `envconfig:"SERVER_ADDRESS" default:"localhost:8080"`
	IOTimeout     time.Duration `envconfig:"IO_TIMEOUT" default:"10s"`
	NIterations   int           `envconfig:"N_ITERATIONS" default:"10"`
}

func main() {
	rootCtx := signal.SignalContext()

	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		slog.ErrorContext(rootCtx, "failed to process env", "error", err)
		return
	}

	if err := run(rootCtx, cfg); err != nil {
		slog.ErrorContext(rootCtx, "error on running client", "error", err)
		return
	}
	slog.InfoContext(rootCtx, "client exit")
}

func run(ctx context.Context, cfg config) error {
	c, err := client.New(client.Config{
		ServerAddress: cfg.ServerAddress,
		IOTimeout:     cfg.IOTimeout,
		NIterations:   cfg.NIterations,
		Transport:     transport.New(),
		Pow:           pow.New(0),
	})

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return c.Run(ctx)
}
