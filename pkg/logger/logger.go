package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func InitLogger(service, version, level string, writer io.Writer) error {
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
		return fmt.Errorf("unknown log level: %s", level)
	}

	jsonHandler := slog.
		NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     programLevel,
			AddSource: false,
		}).
		WithAttrs([]slog.Attr{
			slog.String("service", service),
			slog.String("version", version),
		})
	newSlog := slog.New(cxtHandler{jsonHandler})

	slog.SetDefault(newSlog)

	return nil
}

type cxtHandler struct {
	slog.Handler
}

// Handle извлекаем нужные данные из контекста для отображения в логе
func (h cxtHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	return h.Handler.Handle(ctx, r) //nolint:wrapcheck
}
