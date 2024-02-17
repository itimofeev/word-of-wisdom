package transport

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"
)

// Transport provides methods for writing and reading arbitrary messages via net.Conn.
type Transport struct {
}

func New() *Transport {
	return &Transport{}
}

func (t *Transport) WriteMessage(conn net.Conn, timeout time.Duration, msg any) error {
	defer setWriteDeadline(conn, timeout)()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err = binary.Write(conn, binary.BigEndian, uint64(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (t *Transport) ReadMessage(conn net.Conn, timeout time.Duration, msg any) error {
	defer setReadDeadline(conn, timeout)()

	var length uint64
	err := binary.Read(conn, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("failed to read message length: %w", err)
	}

	r := io.LimitReader(conn, int64(length))

	err = json.NewDecoder(r).Decode(msg)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	return nil
}

func setWriteDeadline(conn net.Conn, timeout time.Duration) func() {
	if timeout == 0 {
		return func() {}
	}

	err := conn.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		slog.Warn("failed to set deadline", "error", err)
		return func() {}
	}

	return func() {
		noDeadline := time.Time{}
		if err := conn.SetDeadline(noDeadline); err != nil {
			slog.Warn("failed to reset deadline", "error", err)
		}
	}
}

func setReadDeadline(conn net.Conn, timeout time.Duration) func() {
	if timeout == 0 {
		return func() {}
	}

	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		slog.Warn("failed to set deadline", "error", err)
		return func() {}
	}

	return func() {
		noDeadline := time.Time{}
		if err := conn.SetReadDeadline(noDeadline); err != nil {
			slog.Warn("failed to reset deadline", "error", err)
		}
	}
}
