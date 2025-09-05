// Program.cs
using System.Text.Json;
using StreamJsonRpc;

public class McpServer
{
    [JsonRpcMethod("ping")]
    public object Ping()
    {
        return new { message = "pong" };
    }

    [JsonRpcMethod("resources/list")]
    public object ListResources()
    {
        return new[]
        {
            new {
                uri = "example://hello",
                name = "Hello Resource",
                description = "A simple resource that returns 'Hello World!'"
            }
        };
    }

    [JsonRpcMethod("resources/read")]
    public object ReadResource(string uri)
    {
        if (uri == "example://hello")
        {
            return new[]
            {
                new {
                    uri,
                    text = "Hello World from C# MCP server!"
                }
            };
        }

        throw new Exception($"Unknown resource: {uri}");
    }
}

class Program
{
    static async Task Main(string[] args)
    {
        var server = new McpServer();
        var jsonRpc = new JsonRpc(Console.OpenStandardInput(), Console.OpenStandardOutput(), server);
        jsonRpc.StartListening();

        await Task.Delay(-1); // keep process alive
    }
}
