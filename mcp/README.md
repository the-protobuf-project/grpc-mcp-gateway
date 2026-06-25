# mcp/protobuf

> [!NOTE]
> This directory contains **pre-compiled proto libraries** generated from the MCP annotation definitions in [`proto/mcp/protobuf/`](../proto/mcp/protobuf/).
> Do not edit these files manually — regenerate them with `just generate-proto` or `cd proto && buf generate`.

Pre-compiled Go types for the `mcp` proto package. These register the MCP extension fields on `google.protobuf.ServiceOptions` and `google.protobuf.MethodOptions`. Types for other languages are generated client-side from the published Buf module ([`buf.build/the-protobuf-project/mcp`](https://buf.build/the-protobuf-project/mcp)).

| Directory | Language | Package | Registry | Documentation |
| --------- | -------- | ------- | -------- | ------------- |
| [`protobuf/`](protobuf/) | Go | `github.com/.../mcp/protobuf/mcppb` | [pkg.go.dev](https://pkg.go.dev/github.com/the-protobuf-project/grpc-mcp-gateway/mcp/protobuf/mcppb) | [README](protobuf/README.md) |

## Example protos

See [examples/proto/](../examples/proto/) for **TodoService** (CRUD, prompts, elicitation) and **CounterService** (progress streaming) implementations.
