"""
Smoke test for the Python TodoService MCP server.

Verifies the full CRUD pipeline over an in-memory MCP transport:
  1. List tools — expect 5
  2. CreateTodo
  3. GetTodo
  4. ListTodos
  5. DeleteTodo
  6. Verify deletion

Usage:
    cd example/python && uv run python -m pytest smoke_test.py -v
"""

from __future__ import annotations

import asyncio
import json
import os
import sys

# Make generated code importable. Must run before the first `mcp` import so the
# vendored mcp.protobuf namespace (under proto/generated/python/mcp) is loaded
# and extends onto the installed `mcp` SDK package via pkgutil.extend_path.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "proto", "generated", "python"))

import pytest  # noqa: E402
from anyio import create_memory_object_stream  # noqa: E402
from mcp import types  # noqa: E402
from mcp.client.session import ClientSession  # noqa: E402
from mcp.server.lowlevel import NotificationOptions, Server  # noqa: E402
from mcp.server.models import InitializationOptions  # noqa: E402

from todo.v1.todo_service_pb2_mcp import register_todo_service_mcp_handler  # noqa: E402

from internal.impl import TodoServer  # noqa: E402


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def impl() -> TodoServer:
    """Fresh in-memory TodoService implementation."""
    return TodoServer()


# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------

@pytest.mark.asyncio
async def test_smoke_todo_service(impl: TodoServer) -> None:
    """Full CRUD smoke test over in-memory MCP transport."""

    # --- Server setup ---
    server = Server("smoke-test")
    register_todo_service_mcp_handler(server, impl)

    init_options = InitializationOptions(
        server_name="smoke-test",
        server_version="0.0.1",
        capabilities=server.get_capabilities(
            notification_options=NotificationOptions(),
            experimental_capabilities={},
        ),
    )

    # --- In-memory transport (paired streams) ---
    server_tx, server_rx = create_memory_object_stream[types.JSONRPCMessage](100)
    client_tx, client_rx = create_memory_object_stream[types.JSONRPCMessage](100)

    async with asyncio.TaskGroup() as tg:
        tg.create_task(server.run(client_rx, server_tx, init_options))

        async with ClientSession(server_rx, client_tx) as session:
            await session.initialize()

            # 1) List tools
            tools_result = await session.list_tools()
            assert len(tools_result.tools) == 5
            for tool in tools_result.tools:
                print(f"  - {tool.name}: {tool.description}")

            # 2) CreateTodo
            create = await session.call_tool(
                "todo_service-create_todo_v1",
                {
                    "parent": "users/alice",
                    "todo_id": "task-1",
                    "todo": {
                        "title": "Buy groceries",
                        "description": "Milk, eggs, bread",
                        "priority": "PRIORITY_HIGH",
                    },
                },
            )
            data = json.loads(create.content[0].text)
            assert data["name"] == "users/alice/todos/task-1"
            assert data["title"] == "Buy groceries"

            # 3) GetTodo
            get = await session.call_tool(
                "todo_service-get_todo_v1",
                {"name": "users/alice/todos/task-1"},
            )
            assert json.loads(get.content[0].text)["title"] == "Buy groceries"

            # 4) ListTodos
            lst = await session.call_tool(
                "todo_service-list_todos_v1",
                {"parent": "users/alice"},
            )
            assert "task-1" in lst.content[0].text

            # 5) DeleteTodo
            await session.call_tool(
                "todo_service-delete_todo_v1",
                {"name": "users/alice/todos/task-1"},
            )

            # 6) Verify deletion
            lst2 = await session.call_tool(
                "todo_service-list_todos_v1",
                {"parent": "users/alice"},
            )
            assert "task-1" not in lst2.content[0].text

            tg._abort()
