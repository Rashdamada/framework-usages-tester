// server.ts
import { Server } from "@modelcontextprotocol/sdk/server";
import { Resource, TextResourceContents } from "@modelcontextprotocol/sdk/types";

// Create server instance
const server = new Server("example-mcp-node-server");

// --- Handlers ---

server.resources.list(async () => {
  return [
    {
      uri: "example://hello",
      name: "Hello Resource",
      description: "A simple resource that returns 'Hello World!'",
    } as Resource,
  ];
});

server.resources.read(async (uri: string) => {
  if (uri === "example://hello") {
    return [
      {
        uri,
        text: "Hello World from Node MCP server!",
      } as TextResourceContents,
    ];
  }
  throw new Error(`Unknown resource: ${uri}`);
});

server.ping(async () => {
  return { message: "pong" };
});

// --- Entry Point ---
server.start();
