// Package runtime provides the Go runtime for grpc-mcp-gateway: server configuration,
// transport handling, and MCP integration helpers.
//
// # Install
//
//	go get github.com/the-protobuf-project/grpc-mcp-gateway/runtime
//
// # Quick Start
//
// Use with generated Serve*MCP functions from protoc-gen-mcp:
//
//	cfg := &runtime.MCPServerConfig{
//	    Name:             "my-service",
//	    Version:          "1.0.0",
//	    Transports:       []runtime.Transport{runtime.TransportStreamableHTTP},
//	    Addr:             ":8082",
//	    GeneratedBasePath: yourpb.YourServiceMCPDefaultBasePath,
//	}
//	err := yourpb.ServeYourServiceMCP(ctx, yourGRPCServerImpl, cfg)
//
// # Transports
//
//   - TransportStdio: stdin/stdout (for IDE integrations, local tools)
//   - TransportStreamableHTTP: modern HTTP transport (default for production)
//   - TransportSSE: legacy SSE transport
//
// Parse from env: runtime.ParseTransports(os.Getenv("MCP_TRANSPORT"))
//
// # Error Handling
//
// Convert gRPC errors to MCP tool results in your tool handlers:
//
//	resp, err := srv.MyRPC(ctx, req)
//	if err != nil {
//	    return runtime.HandleError(err)
//	}
//
// # Elicitation
//
// Run confirmation dialogs before tool execution:
//
//	result, err := runtime.RunElicitation(ctx, session, "Are you sure?", []runtime.ElicitField{
//	    {Name: "confirm", Required: true, Type: "string", EnumValues: []string{"yes", "no"}},
//	})
package runtime
