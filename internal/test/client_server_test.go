package test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/itimofeev/word-of-wisdom/internal/app/client"
	"github.com/itimofeev/word-of-wisdom/internal/app/server"
	"github.com/itimofeev/word-of-wisdom/internal/repository/quotes"
	"github.com/itimofeev/word-of-wisdom/pkg/pow"
	"github.com/itimofeev/word-of-wisdom/pkg/transport"
)

func TestIntegration(t *testing.T) {
	repo, err := quotes.New()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	port := getFreePort(t)

	s, err := server.New(server.Config{
		Addr:             ":" + strconv.FormatInt(int64(port), 10),
		Difficulty:       3,
		Pow:              pow.New(10),
		Transport:        transport.New(),
		QuotesRepository: repo,
		IOTimeout:        time.Second,
		SolveTimeout:     time.Second * 10,
	})
	require.NoError(t, err)

	c, err := client.New(client.Config{
		ServerAddress: "localhost:" + strconv.FormatInt(int64(port), 10),
		IOTimeout:     time.Second,
		NIterations:   1,
		Transport:     transport.New(),
		Pow:           pow.New(0),
	})
	require.NoError(t, err)

	serverStoppedCh := make(chan struct{})

	go func() {
		defer close(serverStoppedCh)

		require.NoError(t, s.Start(ctx))
	}()

	require.NoError(t, c.Run(ctx))

	cancel()

	require.Eventually(t, func() bool {
		_, open := <-serverStoppedCh
		return !open
	}, time.Second, 100*time.Millisecond, "server has to be stopped")

}

func getFreePort(t *testing.T) int {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port
}
