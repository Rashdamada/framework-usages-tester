# agent.py
import datetime
from openai import OpenAI
from openai.types.beta import AssistantTool, ToolFunction

client = OpenAI()

# Define a simple tool
def get_current_time():
    return datetime.datetime.utcnow().isoformat()

# Register tool with the agent
tools = [
    AssistantTool(
        type="function",
        function=ToolFunction(
            name="get_current_time",
            description="Returns the current UTC time",
            parameters={}
        )
    )
]

# Create an agent
assistant = client.beta.assistants.create(
    name="Time Agent",
    instructions="You are an assistant that can tell the current time using the tool.",
    model="gpt-4.1",
    tools=tools
)

# Run a test
thread = client.beta.threads.create()
run = client.beta.threads.runs.create(
    thread_id=thread.id,
    assistant_id=assistant.id,
    input="What time is it right now?"
)

print(run)
