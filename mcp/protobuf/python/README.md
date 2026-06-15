# grpc-mcp-gateway-protos

Pre-compiled Protocol Buffer types for [grpc-mcp-gateway](https://github.com/the-protobuf-project/grpc-mcp-gateway) — the `mcp.protobuf` package containing MCP annotations for gRPC services.

## Install

Use the version matching the [grpc-mcp-gateway release](https://github.com/the-protobuf-project/grpc-mcp-gateway/releases) you use:

```bash
# Latest
pip install grpc-mcp-gateway-protos

# Or pin to a specific release (matches gateway / GitHub release tag without v)
pip install grpc-mcp-gateway-protos==1.5.63
```

## What's included

This package provides the Python bindings for:

- **`mcp.protobuf.annotations`** — Service, tool, prompt, elicitation, and field options for annotating `.proto` files
- **`mcp.protobuf.service_options`** — App metadata (`MCPServiceOptions`)
- **`mcp.protobuf.prompt`** — Prompt templates (`MCPPrompt`, `MCPToolOptions`)
- **`mcp.protobuf.elicitation`** — Confirmation dialogs (`MCPElicitation`)
- **`mcp.protobuf.resource`** — Resource definitions (`MCPResource`, `MCPMimeType`)
- **`mcp.protobuf.app`** — App info (`MCPApp`)
- **`mcp.protobuf.field`** — Field description option (`MCPFieldOptions`)
- **`mcp.protobuf.enum`** — Enum and enum-value descriptions (`MCPEnumOptions`, `MCPEnumValueOptions`)
- **`mcp.protobuf.progress`** — Progress notifications (`MCPProgress`) for server-streaming RPCs
- **`mcp.protobuf.field_type`** — Field type enums

## Usage

Import the annotations to register the proto extensions (required for generated code that uses MCP options):

```python
# In your generated _pb2.py or before using MCP-annotated protos
import mcp.protobuf.annotations_pb2  # noqa: F401 — registers extensions
```

When using [protoc-gen-mcp](https://github.com/the-protobuf-project/grpc-mcp-gateway) with `lang=python`, the generated code will depend on this package. Add it to your project:

```toml
# pyproject.toml
[project]
dependencies = [
    "grpc-mcp-gateway-protos",
    # ...
]
```

## Links

- **Source**: [github.com/the-protobuf-project/grpc-mcp-gateway](https://github.com/the-protobuf-project/grpc-mcp-gateway)
- **Proto definitions**: [buf.build/the-protobuf-project/grpc-mcp-gateway](https://buf.build/the-protobuf-project/grpc-mcp-gateway)
- **License**: Apache-2.0
