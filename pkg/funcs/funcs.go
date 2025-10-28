package funcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
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
	receiver any,
) error {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // default

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close response body", slog.String("error", err.Error()))
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body) // cut data (once)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf( //nolint:err113
			"response has statusCode: %d, status: %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	// задаём cookie если только они есть, иначе будет паника
	if cooks := resp.Cookies(); len(cooks) > 0 {
		client.Jar.SetCookies(req.URL, cooks)
	}

	if receiver != nil {
		if err = json.Unmarshal(bodyBytes, receiver); err != nil {
			return fmt.Errorf("failed to unmarshal response body to receiver: %w (%s)", err, string(bodyBytes))
		}
	}

	return nil
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
