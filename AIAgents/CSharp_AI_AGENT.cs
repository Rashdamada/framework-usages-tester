// Program.cs
using OpenAI;
using OpenAI.Assistants;
using OpenAI.Assistants.Tools;

class Program
{
    static async Task Main(string[] args)
    {
        var client = new OpenAIClient(Environment.GetEnvironmentVariable("OPENAI_API_KEY"));

        // Define a function tool
        var tools = new List<Tool>
        {
            Tool.CreateFunctionTool(
                name: "get_current_time",
                description: "Returns the current UTC time",
                parameters: new { }
            )
        };

        // Create agent (assistant)
        var assistant = await client.AssistantsEndpoint.CreateAssistantAsync(
            name: "Time Agent",
            instructions: "You are an assistant that can tell the current time using the tool.",
            model: "gpt-4.1",
            tools: tools
        );

        // Start a thread
        var thread = await client.ThreadsEndpoint.CreateThreadAsync();

        // Ask it something
        var run = await client.ThreadsEndpoint.CreateRunAsync(thread.Id, assistant.Id, "What's the time now?");
        Console.WriteLine(run.Id);
    }
}
