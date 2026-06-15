# mcp/protobuf

> [!NOTE]
> This directory contains **pre-compiled proto libraries** generated from the MCP annotation definitions in [`proto/mcp/protobuf/`](../proto/mcp/protobuf/).
> Do not edit these files manually — regenerate them with `just generate-proto` or `cd proto && buf generate`.

Pre-compiled types for the `mcp.protobuf` proto package, published as libraries for Go, Python, and Rust. These register the MCP extension fields on `google.protobuf.ServiceOptions` and `google.protobuf.MethodOptions`.

| Directory | Language | Package | Registry | Documentation |
| --------- | -------- | ------- | -------- | ------------- |
| [`protobuf/`](protobuf/) | Go | `github.com/.../mcp/protobuf/mcppb` | [pkg.go.dev](https://pkg.go.dev/github.com/the-protobuf-project/grpc-mcp-gateway/mcp/protobuf/mcppb) | [README](protobuf/README.md) |
| [`python/`](protobuf/python) | Python | `grpc-mcp-gateway-protos` | [PyPI](https://pypi.org/project/grpc-mcp-gateway-protos/) | [README](protobuf/python/README.md) |
| [`rust/`](protobuf/rust) | Rust | `mcp-protobuf` | [crates.io](https://crates.io/crates/mcp-protobuf) | [README](protobuf/rust/README.md) |

## Example protos

See [examples/proto/](../examples/proto/) for **TodoService** (CRUD, prompts, elicitation) and **CounterService** (progress streaming) implementations.
