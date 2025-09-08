"""
Context-aware task tool for DeepAgents that injects supervisor context into sub-agents.

This module provides a custom task tool that extracts context from the supervisor agent's
RunnableConfig and injects it into sub-agent prompts, ensuring consistent context across
all agent levels.

Key features:
- Extracts context (repository, user_id, session_id) from RunnableConfig
- Injects context into sub-agent prompts at runtime
- Creates minimal state for sub-agents (only messages, not full AgentState)
- Maintains clean separation between supervisor and sub-agent state
"""

import logging
from typing import Annotated, Any

from deepagents.prompts import TASK_DESCRIPTION_PREFIX, TASK_DESCRIPTION_SUFFIX
from deepagents.state import DeepAgentState
from deepagents.sub_agent import SubAgent
from langchain_core.messages import HumanMessage, ToolMessage
from langchain_core.runnables import RunnableConfig
from langchain_core.tools import BaseTool, InjectedToolCallId, tool
from langgraph.prebuilt import create_react_agent
from langgraph.types import Command

from .context import construct_message_with_context
from .tool_trimming_hook import trim_tool_response_hook

logger = logging.getLogger(__name__)


def _create_context_aware_task_message(
    task_description: str, config: RunnableConfig
) -> HumanMessage:
    """Create a task message with context tags using existing construct_message_with_context function."""
    if not config or not hasattr(config, "get") or "configurable" not in config:
        return HumanMessage(
            content=f"<task_description>\n{task_description}\n</task_description>"
        )

    configurable = config["configurable"]

    # Use the existing construct_message_with_context function with task_description tag
    return construct_message_with_context(
        query=task_description,
        repository=configurable.get("repository"),
        branch=configurable.get("branch"),
        repository_key=configurable.get("repository_key"),
        file_path=configurable.get("file_path"),
        user_policies=configurable.get("user_policies"),
        query_wrapper_tag="task_description",
    )


def _create_context_aware_task_tool(
    main_agent_tools: list[BaseTool],
    all_tools: list[BaseTool],
    instructions: str,
    subagents: list[SubAgent],
    model: Any,
    state_schema: type,
) -> BaseTool:
    """Create a context-aware task tool that injects supervisor context into sub-agents."""

    agents = {
        "general-purpose": create_react_agent(
            model,
            prompt=instructions,
            tools=main_agent_tools,
            pre_model_hook=trim_tool_response_hook,
        )
    }

    tools_by_name = {}
    for tool_ in all_tools:
        if not isinstance(tool_, BaseTool):
            tool_ = tool(tool_)
        tools_by_name[tool_.name] = tool_

    # Create sub-agents with context injection capability
    for _agent in subagents:
        if "tools" in _agent:
            _tools = [tools_by_name[t] for t in _agent["tools"]]
        else:
            _tools = all_tools

        # Store the original prompt for context injection
        base_prompt = _agent["prompt"]

        def create_context_aware_agent(
            agent_name: str, agent_prompt: str, agent_tools: list[BaseTool]
        ):
            """Create an agent that can receive context dynamically."""

            def context_aware_agent_wrapper(
                state: DeepAgentState, config: RunnableConfig = None
            ):
                agent = create_react_agent(
                    model,
                    prompt=agent_prompt,
                    tools=agent_tools,
                    state_schema=state_schema,
                    pre_model_hook=trim_tool_response_hook,
                )

                # Invoke with the provided config to maintain context chain
                return agent.invoke(state, config=config)

            return context_aware_agent_wrapper

        agents[_agent["name"]] = create_context_aware_agent(
            _agent["name"], base_prompt, _tools
        )

    # Build agent descriptions for the task tool
    other_agents_string = [
        f"- {_agent['name']}: {_agent['description']}" for _agent in subagents
    ]

    @tool(
        description=TASK_DESCRIPTION_PREFIX.format(other_agents=other_agents_string)
        + TASK_DESCRIPTION_SUFFIX
    )
    def task(
        description: str,
        subagent_type: str,
        tool_call_id: Annotated[str, InjectedToolCallId],
        config: RunnableConfig,
    ):
        """Execute a task using the specified sub-agent with context injection."""
        if subagent_type not in agents:
            return f"Error: invoked agent of type {subagent_type}, the only allowed types are {[f'`{k}`' for k in agents]}"

        sub_agent = agents[subagent_type]

        # Create a minimal state with context-aware task message
        context_aware_message = _create_context_aware_task_message(description, config)
        sub_state = {"messages": [context_aware_message]}
        try:
            # Invoke the sub-agent with context
            if subagent_type == "general-purpose":
                # General purpose agent doesn't need special context handling
                result = sub_agent.invoke(sub_state, config=config)
            else:
                # Context-aware sub-agents get the config for context injection
                result = sub_agent(sub_state, config)

            logger.debug(f"Sub-agent {subagent_type} completed task successfully")

            return Command(
                update={
                    "files": result.get("files", {}),
                    "messages": [
                        ToolMessage(
                            result["messages"][-1].content, tool_call_id=tool_call_id
                        )
                    ],
                }
            )

        except Exception as e:
            logger.error(f"Sub-agent {subagent_type} failed: {e}")
            return Command(
                update={
                    "messages": [
                        ToolMessage(
                            f"Sub-agent execution failed: {str(e)}",
                            tool_call_id=tool_call_id,
                        )
                    ],
                }
            )

    return task
