package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

// ServerStream обертка для ServerStream (original) interface
type ServerStream struct {
	grpc.ServerStream

	ctx context.Context //nolint:containedctx
}

func (s *ServerStream) Context() context.Context {
	return s.ctx
}

// NewServerStream ctx может приходить другой
func NewServerStream(ctx context.Context, ss grpc.ServerStream) *ServerStream {
	return &ServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}
