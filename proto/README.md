# grpc-mcp-gateway

[![BSR](https://img.shields.io/badge/BSR-buf.build%2Fmachanirobotics%2Fgrpc--mcp--gateway-blue)](https://buf.build/the-protobuf-project/grpc-mcp-gateway)

Protobuf annotations for exposing gRPC services as [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) servers.

## Installation

Add this module as a dependency in your `buf.yaml`:

```yaml
version: v2
deps:
  - buf.build/the-protobuf-project/grpc-mcp-gateway
```

Then run:

```bash
buf dep update
```

## Usage

Import the annotations in your `.proto` files:

```protobuf
import "mcp/protobuf/annotations.proto";
```

### Service-level options

Configure MCP app metadata on your gRPC service:

```protobuf
service MyService {
  option (mcp.protobuf.service) = {
    app: {
      name: "My App"
      version: "1.0.0"
      description: "A brief description of your application"
    }
  };
}
```

### Tool options

Override the auto-generated MCP tool name or description on individual RPCs:

```protobuf
rpc CreateItem(CreateItemRequest) returns (Item) {
  option (mcp.protobuf.tool) = {
    description: "Creates a new item with the given fields."
  };
}
```

### Prompt options

Attach a prompt template to an RPC. The `schema` field references a proto message
whose fields define the prompt arguments:

```protobuf
rpc GetItem(GetItemRequest) returns (Item) {
  option (mcp.protobuf.prompt) = {
    name: "summarize_items"
    description: "Summarize all items for a user"
    schema: "mypackage.SummarizeItemsArgs"
  };
}
```

### Elicitation options

Attach a confirmation dialog to an RPC before it executes:

```protobuf
rpc DeleteItem(DeleteItemRequest) returns (google.protobuf.Empty) {
  option (mcp.protobuf.elicitation) = {
    message: "Are you sure you want to delete this item?"
    schema: "mypackage.DeleteConfirmation"
  };
}
```

### Field options

Add JSON Schema metadata to message fields for the MCP tool inputSchema:

```protobuf
string name = 1 [(mcp.protobuf.field) = {
  description: "Resource name of the item."
  examples: "items/123"
  format: "uri"
}];
```

### Enum options

Add descriptions to enum types and individual enum values:

```protobuf
enum Priority {
  option (mcp.protobuf.enum) = { description: "Priority level." };
  LOW = 0 [(mcp.protobuf.enum_value) = { description: "Low priority." }];
  HIGH = 1 [(mcp.protobuf.enum_value) = { description: "High priority." }];
}
```

### Progress (server streaming)

For long-running RPCs, use server streaming with `MCPProgress` to report progress:

```protobuf
import "mcp/protobuf/progress.proto";

message CreateItemStreamChunk {
  oneof payload {
    mcp.protobuf.MCPProgress progress = 1;
    Item result = 2;
  }
}

rpc CreateItem(CreateItemRequest) returns (stream CreateItemStreamChunk);
```

## Available Protos

| File                                 | Description                                       |
| ------------------------------------ | ------------------------------------------------- |
| `mcp/protobuf/annotations.proto`     | Service, tool, prompt, elicitation, field, enum extensions |
| `mcp/protobuf/app.proto`             | `MCPApp` message (name, version, description)     |
| `mcp/protobuf/prompt.proto`          | `MCPPrompt` and `MCPToolOptions` messages         |
| `mcp/protobuf/elicitation.proto`     | `MCPElicitation` message                          |
| `mcp/protobuf/service_options.proto` | `MCPServiceOptions` message                       |
| `mcp/protobuf/resource.proto`        | `MCPResource` message                             |
| `mcp/protobuf/field.proto`           | `MCPFieldOptions` (description, examples, format) |
| `mcp/protobuf/enum.proto`            | `MCPEnumOptions`, `MCPEnumValueOptions`           |
| `mcp/protobuf/progress.proto`        | `MCPProgress` for server-streaming progress       |
| `mcp/protobuf/field_type.proto`      | `MCPFieldType` enum                               |
| `mcp/protobuf/mime_type.proto`       | `MCPMimeType` enum                                |

## License

See the repository root for license information.
