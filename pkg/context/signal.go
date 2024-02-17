package context

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// SignalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func SignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.InfoContext(ctx, "received signal:", "signal", sig)
		cancel()
	}()

	return ctx
}
