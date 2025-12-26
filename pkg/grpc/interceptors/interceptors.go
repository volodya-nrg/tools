package interceptors

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/volodya-nrg/tools/pkg/errors/custom"
)

func InterceptorErrorLogic(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var (
		code        = codes.Internal
		msg         = "internal server error"
		customErr   *custom.CustomError
		originalErr error
	)

	if s, ok := status.FromError(err); ok {
		originalErr = s.Err()

		if s.Code() != codes.Internal {
			code = s.Code()
			msg = s.Message()
		}
	} else if errors.As(err, &customErr) {
		code = customErr.GetCode()
		msg = customErr.Error()
		originalErr = customErr.Unwrap()
	} else {
		originalErr = err
	}

	// в любом случае надо показать
	if originalErr != nil {
		slog.ErrorContext(ctx, msg, slog.String("error", originalErr.Error()))
	}

	return status.New(code, msg).Err()
}
