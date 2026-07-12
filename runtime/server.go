package runtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/grpc"
)

// MCPServerConfig holds configuration for starting an MCP server.
// Set Transports (or Transport) to choose one or more wire protocols.
// When multiple transports are specified they run concurrently in the
// same process -- e.g. ["stdio", "streamable-http"].
type MCPServerConfig struct {
	// Name is the MCP server name reported during initialization.
	Name string
	// Version is the MCP server version reported during initialization.
	Version string
	// Transport selects a single wire protocol (for backward compatibility).
	// Ignored when Transports is non-empty.
	Transport Transport
	// Transports selects one or more wire protocols to serve concurrently.
	// Takes precedence over Transport.
	Transports []Transport
	// Addr is the listen address for HTTP-based transports (default ":8080").
	Addr string
	// ServerOptions are passed to mcp.NewServer.
	ServerOptions *mcp.ServerOptions
	// StreamableHTTPOptions are passed to mcp.NewStreamableHTTPHandler.
	StreamableHTTPOptions *mcp.StreamableHTTPOptions
	// SSEOptions are passed to mcp.NewSSEHandler.
	SSEOptions *mcp.SSEOptions
	// BasePath is the HTTP path prefix for the MCP endpoint (default "/mcp").
	BasePath string
	// GeneratedBasePath is the proto-derived default BasePath. If set, it takes precedence over BasePath.
	GeneratedBasePath string
	// HeaderMappings configures HTTP header to gRPC metadata forwarding.
	// Each entry maps an HTTP header name to a gRPC metadata key.
	// Use DefaultHeaderMappings() for common headers.
	HeaderMappings []HeaderMapping
	// OnReady is called after BasePath is resolved, just before the server starts listening.
	// Use this to log or inspect the final endpoint.
	OnReady func(cfg *MCPServerConfig)
	// HealthCheckPath, when non-empty, registers an HTTP GET endpoint that performs
	// a gRPC health check via HealthCheckConn. Returns 200 if SERVING, 503 otherwise.
	// Use with HealthCheckConn for load balancer / k8s probes. Default: "/health".
	HealthCheckPath string
	// HealthCheckConn is the gRPC connection used for health checks when HealthCheckPath is set.
	// The backend gRPC server should register grpc_health_v1.HealthServer.
	HealthCheckConn *grpc.ClientConn
	// ReadTimeout is the maximum duration for reading the entire request. Zero means no limit.
	// For progress-enabled tools, keep at 0 so long-running requests are not interrupted.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response. Zero means no limit.
	// For progress-enabled tools, keep at 0 so streaming progress notifications do not time out.
	WriteTimeout time.Duration
	// UnaryInterceptor, when set, is wrapped around every unary tool call the
	// served handlers dispatch — the hosting server's chance to push its gRPC
	// unary interceptor chain (validation, auth, tracing) down to MCP so both
	// surfaces enforce identical middleware. Generated Serve*MCP functions
	// forward it to the handler registration; see Config.UnaryInterceptor.
	UnaryInterceptor grpc.UnaryServerInterceptor
	// Mux, when non-nil, mounts the HTTP transports on this shared mux at
	// BasePath instead of starting a listener of the server's own — the seam a
	// hosting process uses to serve many MCP services from ONE port, routed by
	// their distinct base paths. StartServer then blocks until ctx is
	// cancelled; the host owns the http.Server lifecycle. Ignored by the stdio
	// transport, which cannot share.
	Mux *http.ServeMux
}

// NewMCPServer creates an mcp.Server from a MCPServerConfig.
func NewMCPServer(cfg *MCPServerConfig) *mcp.Server {
	opts := cfg.ServerOptions
	if opts == nil {
		opts = &mcp.ServerOptions{}
	}
	return mcp.NewServer(&mcp.Implementation{Name: cfg.Name, Version: cfg.Version}, opts)
}

// ParseTransports splits a comma-separated transport string into a []Transport slice.
// Use with MCP_TRANSPORT env var:
//
//	transports := runtime.ParseTransports(os.Getenv("MCP_TRANSPORT"))
//	if len(transports) == 0 {
//	    transports = []runtime.Transport{runtime.TransportStreamableHTTP}
//	}
func ParseTransports(s string) []Transport {
	parts := strings.Split(s, ",")
	out := make([]Transport, 0, len(parts))
	for _, p := range parts {
		if t := Transport(strings.TrimSpace(p)); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// shutdownTimeout bounds the graceful drain of an HTTP transport once the
// caller's context is cancelled.
const shutdownTimeout = 5 * time.Second

// StartServer starts the MCP server using the configured transport(s).
// Multiple transports run concurrently -- HTTP-based transports share a
// single net/http server while stdio gets its own mcp.Server instance.
// This call blocks until the context is cancelled (HTTP transports are then
// drained gracefully and nil is returned) or a serve error occurs.
func StartServer(ctx context.Context, cfg *MCPServerConfig, register func(s *mcp.Server)) error {
	// Defaults are resolved on a copy so a caller sharing one config across
	// several servers never observes another server's resolved state.
	resolved := *cfg
	cfg = &resolved

	transports := cfg.Transports
	if len(transports) == 0 {
		t := cfg.Transport
		if t == "" {
			t = TransportStreamableHTTP
		}
		transports = []Transport{t}
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.GeneratedBasePath != "" {
		cfg.BasePath = cfg.GeneratedBasePath
	} else if cfg.BasePath == "" {
		cfg.BasePath = "/mcp"
	}

	var httpTransports []Transport
	hasStdio := false
	for _, t := range transports {
		switch t {
		case TransportStdio:
			hasStdio = true
		case TransportSSE, TransportStreamableHTTP:
			httpTransports = append(httpTransports, t)
		default:
			return fmt.Errorf("runtime: unsupported transport %q", t)
		}
	}

	// Notify caller that BasePath is resolved.
	if cfg.OnReady != nil {
		cfg.OnReady(cfg)
	}

	// Shared-mux mode: mount on the host's mux and let it own the listener.
	if cfg.Mux != nil && len(httpTransports) > 0 {
		httpServer := NewMCPServer(cfg)
		register(httpServer)
		mountHTTPTransports(cfg.Mux, httpServer, cfg, httpTransports)
		if hasStdio {
			stdioServer := NewMCPServer(cfg)
			register(stdioServer)
			return serveStdio(ctx, stdioServer)
		}
		<-ctx.Done()
		return nil
	}

	// Start HTTP transport(s) if requested. Header-mapping middleware is
	// applied per mounted handler inside buildHTTPMux.
	if len(httpTransports) > 0 {
		httpServer := NewMCPServer(cfg)
		register(httpServer)

		srv := &http.Server{
			Addr:         cfg.Addr,
			Handler:      buildHTTPMux(httpServer, cfg, httpTransports),
			ReadTimeout:  cfg.ReadTimeout,  // 0 = no limit; progress requests must not time out
			WriteTimeout: cfg.WriteTimeout, // 0 = no limit; streaming progress must not time out
		}
		if hasStdio {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("runtime: HTTP server error: %v", err)
				}
			}()
			// The stdio path below blocks on ctx; drain HTTP when it ends.
			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
			}()
		} else {
			errCh := make(chan error, 1)
			go func() { errCh <- srv.ListenAndServe() }()
			select {
			case err := <-errCh:
				if err != nil && err != http.ErrServerClosed {
					return err
				}
				return nil
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
				return ctx.Err()
			}
		}
	}

	if hasStdio {
		stdioServer := NewMCPServer(cfg)
		register(stdioServer)
		return serveStdio(ctx, stdioServer)
	}
	return fmt.Errorf("runtime: no transports configured")
}

// buildHTTPMux registers HTTP-based transports on a fresh ServeMux.
func buildHTTPMux(server *mcp.Server, cfg *MCPServerConfig, transports []Transport) *http.ServeMux {
	mux := http.NewServeMux()
	mountHTTPTransports(mux, server, cfg, transports)
	return mux
}

// mountHTTPTransports registers the HTTP-based transport handlers for one MCP
// server on mux at cfg.BasePath. Distinct base paths let many MCP services
// share one mux (and therefore one port); header-mapping middleware is applied
// per handler so shared-mux mounts keep their own mappings.
func mountHTTPTransports(mux *http.ServeMux, server *mcp.Server, cfg *MCPServerConfig, transports []Transport) {
	wrap := func(h http.Handler) http.Handler { return HeadersMiddleware(cfg.HeaderMappings, h) }
	for _, t := range transports {
		switch t {
		case TransportStreamableHTTP:
			h := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return server }, cfg.StreamableHTTPOptions)
			mux.Handle(cfg.BasePath, wrap(h))
		case TransportSSE:
			h := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server { return server }, cfg.SSEOptions)
			mux.Handle(cfg.BasePath+"/", wrap(h))
		}
	}
	if cfg.HealthCheckPath != "" && cfg.HealthCheckConn != nil {
		path := cfg.HealthCheckPath
		if path[0] != '/' {
			path = "/" + path
		}
		mux.Handle(path, HealthCheckHandler(cfg.HealthCheckConn))
	}
}

func serveStdio(ctx context.Context, server *mcp.Server) error {
	log.SetOutput(os.Stderr)
	return server.Run(ctx, &mcp.StdioTransport{})
}
