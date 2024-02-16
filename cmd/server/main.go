package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"

	signal "github.com/itimofeev/word-of-wisdom/internal/context"
	"github.com/itimofeev/word-of-wisdom/internal/quotes"
	"github.com/itimofeev/word-of-wisdom/internal/server"
	"github.com/itimofeev/word-of-wisdom/internal/transport"
	"github.com/itimofeev/word-of-wisdom/pkg/pow"
)

type config struct {
	Address string `envconfig:"SERVER_ADDRESS" default:"localhost:8080"`
}

func main() {
	rootCtx := signal.SignalContext()

	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		slog.ErrorContext(rootCtx, "failed to process env", "error", err)
		return
	}

	if err := run(rootCtx, cfg); err != nil {
		slog.ErrorContext(rootCtx, "error on running server", "error", err)
		return
	}
	slog.InfoContext(rootCtx, "server exit")
}

func run(ctx context.Context, cfg config) error {
	repo, err := quotes.New()
	if err != nil {
		return fmt.Errorf("failed to create quotes repository: %w", err)
	}

	s, err := server.New(server.Config{
		Addr:             cfg.Address,
		Difficulty:       5,
		Pow:              pow.New(15),
		Transport:        transport.New(),
		QuotesRepository: repo,
		IOTimeout:        time.Second * 5,
		SolveTimeout:     time.Second * 10,
	})
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return s.Start(egCtx)
	})

	return eg.Wait()
}
