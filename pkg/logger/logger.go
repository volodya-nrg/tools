package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	fd *os.File
}

func (l *Logger) Close() error {
	if l.fd != nil {
		if err := l.fd.Close(); err != nil {
			return fmt.Errorf("failed to close log-file: %w", err)
		}
	}

	return nil
}

func NewLogger(service, version, level, filepath string) (*Logger, error) {
	var (
		logWriter io.Writer = os.Stdout
		result              = Logger{}
	)

	if filepath != "" {
		const perm = 0600 //nolint:gofumpt

		fd, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
		if err != nil {
			return nil, fmt.Errorf("failed to open log-file: %s", err)
		}

		logWriter, result.fd = fd, fd
	}

	programLevel := new(slog.LevelVar)

	switch strings.ToLower(level) {
	case "error":
		programLevel.Set(slog.LevelError) // error
	case "warn":
		programLevel.Set(slog.LevelWarn) // error, warn
	case "info":
		programLevel.Set(slog.LevelInfo) // error, warn, info
	case "debug":
		programLevel.Set(slog.LevelDebug) // error, warn, info, debug
	default:
		return nil, fmt.Errorf("unknown log level: %s", level)
	}

	jsonHandler := slog.
		NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level:     programLevel,
			AddSource: false,
		}).
		WithAttrs([]slog.Attr{
			slog.String("service", service),
			slog.String("version", version),
		})
	newSlog := slog.New(cxtHandler{jsonHandler})

	slog.SetDefault(newSlog)

	return &result, nil
}
