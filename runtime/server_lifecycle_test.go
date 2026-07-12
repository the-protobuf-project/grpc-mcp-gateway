package runtime

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Cancelling the context must stop an HTTP-transport StartServer (it used to
// block in ListenAndServe forever) and must not mutate the caller's config.
func TestStartServer_HTTPShutdownOnContextCancel(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	addr := lis.Addr().String()
	_ = lis.Close()

	cfg := &MCPServerConfig{
		Name:      "test",
		Version:   "0.0.0",
		Transport: TransportStreamableHTTP,
		Addr:      addr,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- StartServer(ctx, cfg, func(s *mcp.Server) {}) }()

	// Give the listener a moment to bind, then cancel.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("StartServer returned %v, want nil or context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("StartServer did not return after context cancellation")
	}

	// The caller's config must be untouched by default resolution.
	if cfg.BasePath != "" {
		t.Fatalf("caller cfg.BasePath mutated to %q", cfg.BasePath)
	}
}
