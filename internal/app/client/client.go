package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/itimofeev/word-of-wisdom/internal/entity"
)

type pow interface {
	SolveChallenge(data []byte, difficulty uint) []byte
}

type transport interface {
	WriteMessage(conn net.Conn, timeout time.Duration, msg any) error
	ReadMessage(conn net.Conn, timeout time.Duration, msg any) error
}

type Config struct {
	ServerAddress string        `validate:"required"`
	IOTimeout     time.Duration `validate:"required"`
	NIterations   int           `validate:"required"`
	Transport     transport     `validate:"required"`
	Pow           pow           `validate:"required"`
}

type Client struct {
	serverAddress string
	ioTimeout     time.Duration
	nIterations   int
	transport     transport
	pow           pow
}

func New(cfg Config) (*Client, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &Client{
		serverAddress: cfg.ServerAddress,
		ioTimeout:     cfg.IOTimeout,
		nIterations:   cfg.NIterations,
		transport:     cfg.Transport,
		pow:           cfg.Pow,
	}, nil
}

func (c *Client) Run(ctx context.Context) error {
	for i := 0; i < c.nIterations; i++ {
		if err := c.fetchQuote(ctx); err != nil {
			return fmt.Errorf("failed to fetch quote: %w", err)
		}
	}
	return nil
}

func (c *Client) fetchQuote(ctx context.Context) error {
	conn, err := net.Dial("tcp", c.serverAddress)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			slog.WarnContext(ctx, "connection close error", "error", err)
		}
	}()
	slog.InfoContext(ctx, "connected to server", "address", c.serverAddress)

	requestChallengeMessage := entity.RequestChallengeMessage{}
	if err := c.transport.WriteMessage(conn, c.ioTimeout, &requestChallengeMessage); err != nil {
		return fmt.Errorf("failed to send request challenge message: %w", err)
	}
	slog.InfoContext(ctx, "request challenge message sent")

	responseChallengeMessage := entity.ResponseChallengeMessage{}
	if err := c.transport.ReadMessage(conn, c.ioTimeout, &responseChallengeMessage); err != nil {
		return fmt.Errorf("failed to read response challenge message: %w", err)
	}

	start := time.Now()
	solution := c.pow.SolveChallenge(responseChallengeMessage.Data, responseChallengeMessage.Difficulty)
	slog.InfoContext(ctx, "solved challenge", "time", time.Since(start))

	if err := c.transport.WriteMessage(conn, c.ioTimeout, &entity.SolvedChallengeMessage{Nonce: solution}); err != nil {
		return fmt.Errorf("failed to send solved challenge message: %w", err)
	}
	slog.InfoContext(ctx, "solved challenge message sent")

	quoteMessage := entity.QuoteMessage{}
	if err := c.transport.ReadMessage(conn, c.ioTimeout, &quoteMessage); err != nil {
		return fmt.Errorf("failed to read quote message: %w", err)
	}

	slog.InfoContext(ctx, "quote message received", "quote", quoteMessage.Quote)

	return nil
}
