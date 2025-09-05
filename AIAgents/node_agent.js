// agent.ts
import OpenAI from "openai";

const client = new OpenAI();

// Define tool
const tools = [
  {
    type: "function",
    function: {
      name: "get_current_time",
      description: "Returns the current UTC time",
      parameters: {},
    },
  },
];

async function main() {
  // Create agent
  const assistant = await client.beta.assistants.create({
    name: "Time Agent",
    instructions: "You are an assistant that can tell the current time using the tool.",
    model: "gpt-4.1",
    tools,
  });

  // Start a thread
  const thread = await client.beta.threads.create();

  // Ask it something
  const run = await client.beta.threads.runs.create(thread.id, {
    assistant_id: assistant.id,
    input: "Please tell me the time.",
  });

  console.log(run);
}

main();
