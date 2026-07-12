package runtime

import (
	"context"

	"google.golang.org/grpc"
)

// Transport represents the transport protocol for the MCP server.
type Transport string

const (
	// TransportStreamableHTTP is the modern Streamable HTTP transport (default).
	TransportStreamableHTTP Transport = "streamable-http"
	// TransportSSE is the legacy SSE transport (2024-11-05 spec).
	TransportSSE Transport = "sse"
	// TransportStdio runs the MCP server over stdin/stdout.
	TransportStdio Transport = "stdio"
)

// Option is a functional option for configuring MCP handler registration.
type Option func(*Config)

// Config holds runtime configuration for MCP handlers.
// It is typically built via ApplyOptions and passed to Register*MCPHandler.
type Config struct {
	// ExtraProperties are injected into tool schemas and extracted from
	// request arguments into context. Use WithExtraProperties to add them.
	ExtraProperties []ExtraProperty
	// HeaderMappings configure HTTP header to gRPC metadata forwarding.
	// Use WithHeaderMappings or DefaultHeaderMappings().
	HeaderMappings []HeaderMapping
	// Transport selects the wire protocol (stdio, streamable-http, sse).
	Transport Transport
	// Addr is the listen address for HTTP transports (default ":8080").
	Addr string
	// ElicitHook is called before every RunElicitation, allowing callers to
	// modify elicitation fields at runtime (e.g. inject dynamic enum values).
	// toolName is the MCP tool name. Returning an error aborts the tool call.
	ElicitHook func(ctx context.Context, toolName string, fields []ElicitField) ([]ElicitField, error)
	// UnaryInterceptor, when set, wraps every unary tool call exactly like a
	// gRPC server's unary interceptor chain wraps an RPC: the generated
	// handlers invoke it (via InvokeUnary) with a grpc.UnaryServerInfo carrying
	// the RPC's full method name, so cross-cutting middleware — request
	// validation, auth, tracing — behaves identically whether a request
	// arrives over gRPC or MCP. Streaming (progress) tools are not covered:
	// they would need a grpc.StreamServerInterceptor, which has a different
	// shape. Compose multiple interceptors with ChainUnaryInterceptors.
	UnaryInterceptor grpc.UnaryServerInterceptor
}

// ExtraProperty defines an additional property to inject into tool schemas
// and extract from request arguments into context.
//
// Example: add an "api_key" property that gets extracted into context:
//
//	runtime.WithExtraProperties(runtime.ExtraProperty{
//	    Name: "api_key", Description: "API key for auth", Required: true,
//	    ContextKey: contextKeyForAPIKey,
//	})
type ExtraProperty struct {
	Name        string // JSON property name in tool arguments
	Description string // Shown in tool schema
	Required    bool   // If true, adds to schema.required
	ContextKey  any    // Key for context.WithValue(ctx, ContextKey, value)
}

// WithExtraProperties returns an Option that adds extra properties to tool
// schemas. These properties are extracted from incoming requests and placed
// into the context using the specified ContextKey.
func WithExtraProperties(properties ...ExtraProperty) Option {
	return func(c *Config) {
		c.ExtraProperties = append(c.ExtraProperties, properties...)
	}
}

// WithHeaderMappings returns an Option that configures HTTP header to gRPC
// metadata forwarding. Each mapping specifies an HTTP header name and the
// corresponding gRPC metadata key. Use DefaultHeaderMappings() for common
// headers (Authorization, X-Request-Id, X-Trace-Id).
func WithHeaderMappings(mappings ...HeaderMapping) Option {
	return func(c *Config) {
		c.HeaderMappings = append(c.HeaderMappings, mappings...)
	}
}

// WithTransport sets the transport protocol for the MCP server.
func WithTransport(t Transport) Option {
	return func(c *Config) {
		c.Transport = t
	}
}

// WithAddr sets the listen address (for HTTP-based transports).
// Defaults to ":8080" if not set.
func WithAddr(addr string) Option {
	return func(c *Config) {
		c.Addr = addr
	}
}

// WithElicitHook returns an Option that sets a hook called before every
// RunElicitation. The hook receives the tool name and current fields, and
// returns the (possibly modified) fields. Use it to inject dynamic enum values.
func WithElicitHook(hook func(ctx context.Context, toolName string, fields []ElicitField) ([]ElicitField, error)) Option {
	return func(c *Config) {
		c.ElicitHook = hook
	}
}

// WithUnaryInterceptor returns an Option that sets the unary interceptor
// wrapped around every unary tool call. Pass ChainUnaryInterceptors(...) to
// compose several.
func WithUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) Option {
	return func(c *Config) {
		c.UnaryInterceptor = interceptor
	}
}

// ChainUnaryInterceptors composes interceptors into one, invoked in the given
// order (the first interceptor is outermost), matching
// grpc.ChainUnaryInterceptor semantics. Nil entries are skipped; with no
// non-nil entries it returns nil.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	chain := make([]grpc.UnaryServerInterceptor, 0, len(interceptors))
	for _, i := range interceptors {
		if i != nil {
			chain = append(chain, i)
		}
	}
	switch len(chain) {
	case 0:
		return nil
	case 1:
		return chain[0]
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		wrapped := handler
		for i := len(chain) - 1; i >= 0; i-- {
			next, interceptor := wrapped, chain[i]
			wrapped = func(ctx context.Context, req any) (any, error) {
				return interceptor(ctx, req, info, next)
			}
		}
		return wrapped(ctx, req)
	}
}

// ApplyOptions creates a Config and applies all provided options.
func ApplyOptions(opts ...Option) *Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
