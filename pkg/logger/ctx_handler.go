package logger

import (
	"context"
	"log/slog"
)

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
