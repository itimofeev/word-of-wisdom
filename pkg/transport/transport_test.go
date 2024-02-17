package transport

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTransport(t *testing.T) {
	cConn, sConn := net.Pipe()

	tr := New()

	msg := "hello"
	msgWritten := make(chan struct{})
	go func() {
		defer close(msgWritten)
		require.NoError(t, tr.WriteMessage(sConn, 0, &msg))
	}()

	var readMsg string
	require.NoError(t, tr.ReadMessage(cConn, 0, &readMsg))

	<-msgWritten

	require.Equal(t, msg, readMsg)
}

func TestDeadline(t *testing.T) {
	cConn, _ := net.Pipe()

	tr := New()

	var readMsg string
	require.EqualError(t, tr.ReadMessage(cConn, time.Millisecond, &readMsg), "failed to read message length: read pipe: i/o timeout")
}
