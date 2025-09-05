# server.py
import asyncio
from mcp.server import Server
from mcp.types import Resource, TextResourceContents

# Create a server instance
server = Server("example-mcp-server")

# --- Handlers ---

@server.list_resources()
async def list_resources() -> list[Resource]:
    """List available resources (like files, APIs, etc.)."""
    return [
        Resource(
            uri="example://hello",
            name="Hello Resource",
            description="A simple resource that returns 'Hello World!'"
        )
    ]

@server.read_resource()
async def read_resource(uri: str):
    """Return resource contents for a given URI."""
    if uri == "example://hello":
        return [
            TextResourceContents(
                uri=uri,
                text="Hello World from MCP server!"
            )
        ]
    raise ValueError(f"Unknown resource: {uri}")

@server.ping()
async def ping():
    """Simple ping for connectivity checks."""
    return {"message": "pong"}

# --- Entry Point ---

async def main():
    await server.run_stdio()  # Communicate over stdio (default for MCP)

if __name__ == "__main__":
    asyncio.run(main())
