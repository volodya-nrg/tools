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
	ws     *websocket.Conn
	wsResp *http.Response
}

func (s *WSClient) Close() error {
	var (
		errs []error
		err  error
	)

	if s.wsResp != nil {
		if err = s.wsResp.Body.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close ws response body: %w", err))
		}
	}

	// на всякий случай подадим сигнал на закрытие
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	if err = s.ws.WriteMessage(websocket.CloseMessage, msg); err != nil {
		errs = append(errs, fmt.Errorf("failed to write close-message: %w", err))
	}

	if err = s.ws.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close websocket connection: %w", err))
	}

	return errors.Join(errs...)
}

func (s *WSClient) GetWS() *websocket.Conn {
	return s.ws
}

func NewWSClient(ctx context.Context, serviceName, address string, tlsConfig *tls.Config) (*WSClient, error) {
	wsHeaders := http.Header{}

	if serviceName != "" {
		wsHeaders.Set("Origin", serviceName)
	}

	d := websocket.Dialer{
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: handshakeTimeout,
	}

	wsConn, wsResp, err := d.DialContext(ctx, address, wsHeaders) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("failed to dial (%s): %w", address, err)
	}

	return &WSClient{
		ws:     wsConn,
		wsResp: wsResp,
	}, nil
}
