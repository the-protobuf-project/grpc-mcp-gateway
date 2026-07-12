package runtime

import (
	"context"

	"google.golang.org/grpc"
)

// InvokeUnary dispatches a unary tool call the way a gRPC server dispatches an
// RPC: when interceptor is non-nil it runs with a grpc.UnaryServerInfo carrying
// the RPC's full method name (e.g. "/pkg.Service/Method") and srv as the
// server, so middleware written for gRPC — request validation, auth, tracing —
// sees an MCP call exactly as it would see the wire call. A nil interceptor
// invokes handler directly. Generated MCP handlers call this instead of the
// service method.
func InvokeUnary(ctx context.Context, interceptor grpc.UnaryServerInterceptor, fullMethod string, srv any, req any, handler grpc.UnaryHandler) (any, error) {
	if interceptor == nil {
		return handler(ctx, req)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: fullMethod,
	}
	return interceptor(ctx, req, info, handler)
}
