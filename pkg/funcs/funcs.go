package funcs

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	interceptorstimeout "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	randSource = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	mu         sync.Mutex
)

func Pointer[T comparable](value T) *T {
	return &value
}

func HTTPRequest(
	ctx context.Context,
	client *http.Client,
	method string,
	u url.URL,
	body []byte,
	headers map[string]string,
) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // default

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close response body", slog.String("error", err.Error()))
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body) // cut data (once)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// задаём cookie если только они есть, иначе будет паника
	if cooks := resp.Cookies(); len(cooks) > 0 {
		client.Jar.SetCookies(req.URL, cooks)
	}

	return resp.StatusCode, bodyBytes, nil // отдаем данные как есть, принимающая сторона распределится ими
}

func RandStrLimit(n int) string {
	mu.Lock()
	defer mu.Unlock()

	letters := []rune("abcdefghijklmnopqrstuvwxyz") // для универсальности пусть будут только буквы нижнего регистра
	lettersLen := len(letters)                      // count runes
	b := make([]rune, n)

	for i := range b {
		randomIdx := randSource.Intn(lettersLen) // 0 - (lettersLen-1)
		b[i] = letters[randomIdx]
	}

	return string(b)
}

func RandStr() string {
	return RandStrLimit(10)
}

func RandEmail() string {
	return RandStr() + "@example.com"
}

func RandIntByRange(minSrc, maxSrc int) int {
	mu.Lock()
	defer mu.Unlock()
	return randSource.Intn(maxSrc-minSrc) + minSrc
}

func RandIP() string {
	return fmt.Sprintf(
		"%d.%d.%d.%d",
		RandIntByRange(0, 255),
		RandIntByRange(0, 255),
		RandIntByRange(0, 255),
		RandIntByRange(0, 255),
	)
}

func RandPort() int {
	return RandIntByRange(1024, 65535)
}

func StrToTime(str string) time.Time {
	t, err := time.Parse(time.DateTime, str)
	if err != nil {
		return time.Time{}
	}
	return t
}

func GrpcClientConn(addr string, timeout time.Duration, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	transportCreds := insecure.NewCredentials()
	if tlsConfig != nil {
		transportCreds = credentials.NewTLS(tlsConfig)
	}

	unaryInterceptors := []grpc.UnaryClientInterceptor{
		interceptorstimeout.UnaryClientInterceptor(timeout),
	}
	var streamInterceptors []grpc.StreamClientInterceptor

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	}

	conn, err := grpc.NewClient(addr, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("could not create a new grpc-client: %w", err)
	}

	return conn, nil
}

func StopSignalNotify(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	reason := <-c
	slog.InfoContext(ctx, "program stopped", slog.String("reason", reason.String()))
	cancel()
}
