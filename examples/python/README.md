# Python Examples

TodoService MCP server examples in Python, demonstrating all supported transports.

## Prerequisites

- Python 3.13+
- [uv](https://docs.astral.sh/uv/) package manager
- Generated code already in `proto/generated/python/` (run `buf generate` from `examples/`)

The generated code includes both **TodoService** (`todo/v1/`) and **CounterService** (`counter/v1/`); the Python examples use TodoService.

## Structure

```
python/
├── impl.py            # Shared in-memory TodoServer implementation
├── grpc_servicer.py   # gRPC adapter wrapping TodoServer
├── pyproject.toml     # Dependencies (mcp, grpcio, protobuf, etc.)
├── smoke_test.py      # In-memory MCP transport smoke test
├── http/
│   └── main.py        # streamable-http + gRPC side-by-side
├── stdio/
│   └── main.py        # stdio transport (for Claude Desktop)
└── sse/
    └── main.py        # SSE transport (legacy)
```

The `impl.py` at the root is shared across all transports. Each transport folder contains a standalone `main.py` entrypoint.

## Setup

```bash
cd examples/python
uv sync
```

## Running

### Streamable HTTP (+ gRPC)

```bash
uv run python http/main.py
# gRPC  → [::]:50051 (reflection enabled)
# MCP   → 0.0.0.0:8082 (streamable-http)
```

Environment variables:
- `MCP_HOST` — bind address (default `0.0.0.0`)
- `MCP_PORT` — MCP port (default `8082`)
- `GRPC_PORT` — gRPC port (default `50051`)

### Stdio

```bash
uv run python stdio/main.py
# MCP communicates over stdin/stdout
```

For MCP Inspector:

```bash
npx @modelcontextprotocol/inspector -- uv run python stdio/main.py
```

### SSE

```bash
uv run python sse/main.py
# MCP → 0.0.0.0:8083 (SSE)
```

Environment variables:
- `MCP_HOST` — bind address (default `0.0.0.0`)
- `MCP_PORT` — MCP port (default `8083`)

## Testing

The smoke test exercises the full CRUD pipeline over an in-memory MCP transport (no network required):

```bash
uv run python -m pytest smoke_test.py -v
```

Test flow:
1. List tools (expects 5)
2. CreateTodo → verify name and title
3. GetTodo → verify title
4. ListTodos → verify item present
5. DeleteTodo
6. ListTodos → verify item removed

## Architecture

The generated code (`todo_service_pb2_mcp.py`) provides:
- `TodoServiceMCPServer` protocol — one async method per RPC
- `register_todo_service_mcp_handler(server, impl)` — registers tools, prompts, and resources on a low-level `mcp.server.lowlevel.Server`
- `serve_todo_service_mcp(impl, transport=..., host=..., port=...)` — convenience function that creates a `FastMCP` server and starts it

## Dependencies

The `mcp.protobuf` annotation types are generated client-side and vendored under
[`../proto/generated/python/mcp`](../proto/generated/python/mcp) — no external
package dependency.

| Package | Purpose |
| --------------------- | ----------------------------------- |
| `mcp` | MCP Python SDK (FastMCP + low-level) |
| `protobuf` | Protocol Buffers runtime |
| `grpcio` | gRPC runtime |
| `grpcio-reflection` | gRPC server reflection |
| `anyio` | Async transport for in-memory tests |
| `pytest` / `pytest-asyncio` | Test runner |
