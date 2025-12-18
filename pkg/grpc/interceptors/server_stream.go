package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

// ServerStream обертка для ServerStream interface
type ServerStream struct {
	grpc.ServerStream

	ctx context.Context //nolint:containedctx
}

func (s *ServerStream) Context() context.Context {
	return s.ctx
}
