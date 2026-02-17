package rpcbus

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
)

const (
	defaultDelim = "⛔" // "⛔", "\n"
	readTimeout  = 10 * time.Second
	timeout      = 3 * time.Second
	keepAlive    = 30 * time.Second
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	delim  string
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil { // клиент закрывает connection
		return fmt.Errorf("failed to close: %w", err)
	}
	return nil
}

func (c *Client) Call(ctx context.Context, method string, params any) ([]byte, error) {
	req := request{
		JsonRpc: "2.0",
		Method:  method,
		// Params:  fmt.Sprintf(`{"client_id": "%s"}`, uuid.NewString()),
		ID: uuid.NewString(), // time.Now().Format(time.RFC3339Nano),
	}

	if params != nil {
		req.Params = params
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = c.writer.Write(reqBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to write msg: %w", err)
	}

	if c.delim != "" {
		_, err = c.writer.WriteString(c.delim)
		if err != nil {
			return nil, fmt.Errorf("failed to write delimiter: %w", err)
		}
	}

	if err = c.writer.Flush(); err != nil { // сброс с буффера, на чтение
		return nil, fmt.Errorf("failed to flush response: %w", err)
	}

	var resp []byte

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

marker:
	for { // считываем прямо, без разделителя
		select {
		case <-ctx.Done():
			if err = ctx.Err(); err != nil {
				slog.ErrorContext(ctx, "failed in call rpc (ctx.Done)", slog.String("error", err.Error()))
			}
			break marker
		default:
			b, err := c.reader.ReadByte()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break marker
				}
				return nil, fmt.Errorf("failed to read byte: %w", err)
			}

			resp = append(resp, b)
		}
	}

	return resp, nil
}

func (c *Client) SetDelim(delim string) {
	c.delim = delim
}

func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		delim:  defaultDelim,
	}, nil
}
