package wsclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const handshakeTimeout = 10 * time.Second

type WSClient struct {
	conn *websocket.Conn
	resp *http.Response
}

func (s *WSClient) Close() error {
	var errs []error

	// подадим сигнал на мягкое закрытие
	err := s.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseGoingAway, ""),
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to write close-message: %w", err))
	}

	// закроем response
	if s.resp != nil {
		if err = s.resp.Body.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close ws response body: %w", err))
		}
	}

	// закроем базовое соединение (жестко)
	if err = s.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close websocket connection: %w", err))
	}

	return errors.Join(errs...)
}

func (s *WSClient) GetConn() *websocket.Conn {
	return s.conn
}

func NewWSClient(ctx context.Context, serviceName, addr string, tlsConfig *tls.Config) (*WSClient, error) {
	wsHeaders := http.Header{}

	if serviceName != "" {
		wsHeaders.Set("Origin", serviceName)
	}

	d := websocket.Dialer{
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: handshakeTimeout,
	}

	conn, resp, err := d.DialContext(ctx, addr, wsHeaders) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("failed to dial (%s): %w", addr, err)
	}

	return &WSClient{
		conn: conn,
		resp: resp,
	}, nil
}

/*
more example:

conn, respWS, err := websocket.DefaultDialer.DialContext(
	ctx,
	fmt.Sprintf("ws://%s/ws", addr),
	http.Header{},
)
*/
