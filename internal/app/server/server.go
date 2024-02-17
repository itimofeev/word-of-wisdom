package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/itimofeev/word-of-wisdom/internal/entity"
)

type quotesRepository interface {
	GetQuote() (string, error)
}

type pow interface {
	MakeChallenge() []byte
	ValidateSolution(data, nonce []byte, difficulty uint) bool
}

type transport interface {
	WriteMessage(conn net.Conn, timeout time.Duration, msg any) error
	ReadMessage(conn net.Conn, timeout time.Duration, msg any) error
}

type Config struct {
	Addr string `validate:"required"`

	Difficulty       uint             `validate:"required,gt=0"`
	Pow              pow              `validate:"required"`
	Transport        transport        `validate:"required"`
	QuotesRepository quotesRepository `validate:"required"`
	IOTimeout        time.Duration    `validate:"required,gt=0"`
	SolveTimeout     time.Duration    `validate:"required,gt=0"`
}

type Server struct {
	addr             string
	difficulty       uint
	pow              pow
	transport        transport
	quotesRepository quotesRepository
	ioTimeout        time.Duration
	solveTimeout     time.Duration
}

func New(cfg Config) (*Server, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &Server{
		addr:             cfg.Addr,
		difficulty:       cfg.Difficulty,
		pow:              cfg.Pow,
		transport:        cfg.Transport,
		quotesRepository: cfg.QuotesRepository,
		ioTimeout:        cfg.IOTimeout,
		solveTimeout:     cfg.SolveTimeout,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		if err := ln.Close(); err != nil {
			slog.WarnContext(ctx, "listener close error", "error", err)
		}
	}()

	slog.InfoContext(ctx, "server started", "addr", s.addr)

	wg := sync.WaitGroup{}

acceptCycle:
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				slog.InfoContext(ctx, "server stopped accepting connections as it closed")
				break acceptCycle
			}

			slog.WarnContext(ctx, "error on accepting connection", "err", err)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConnection(ctx, conn)
		}()
	}

	// waiting until all connections closed
	wg.Wait()

	return nil
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.WarnContext(ctx, "connection close error", "error", err)
		}
	}()

	// todo check that context not canceled

	reqChallenge := entity.RequestChallengeMessage{}
	if err := s.transport.ReadMessage(conn, s.ioTimeout, &reqChallenge); err != nil {
		slog.ErrorContext(ctx, "failed to read request challenge message", "error", err)
		return
	}
	slog.InfoContext(ctx, "request challenge message received")

	challenge := s.pow.MakeChallenge()
	respChallenge := entity.ResponseChallengeMessage{
		Data:       challenge,
		Difficulty: s.difficulty,
	}

	if err := s.transport.WriteMessage(conn, s.ioTimeout, respChallenge); err != nil {
		slog.ErrorContext(ctx, "failed to write response challenge message", "error", err)
		return
	}
	slog.InfoContext(ctx, "response challenge message sent")

	solved := entity.SolvedChallengeMessage{}
	if err := s.transport.ReadMessage(conn, s.solveTimeout, &solved); err != nil {
		slog.ErrorContext(ctx, "failed to receive solved challenge", "error", err)
		return
	}
	slog.InfoContext(ctx, "solved challenge received")

	solutionValid := s.pow.ValidateSolution(challenge, solved.Nonce, s.difficulty)

	if !solutionValid {
		slog.WarnContext(ctx, "solution is invalid, closing conn without a quote")
		return
	}

	slog.InfoContext(ctx, "solution is valid, getting quote")

	quote, err := s.quotesRepository.GetQuote()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get quote", "error", err)
		return
	}

	slog.InfoContext(ctx, "sending quote")
	quoteMessage := entity.QuoteMessage{
		Quote: quote,
	}
	if err := s.transport.WriteMessage(conn, s.ioTimeout, quoteMessage); err != nil {
		slog.ErrorContext(ctx, "failed to send quote", "error", err)
		return
	}

	slog.InfoContext(ctx, "quote sent")
}
