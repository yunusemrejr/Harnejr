# Engineered commands

Harnejr commands are runtime state transitions. They must not be implemented only as LLM prompt templates.

## `/goal`

Starts a judged autonomous goal loop. It stores the goal, locks out `/loop`, selects relevant skills, checks MCP availability, schedules subagents, executes work, and challenges completion with a judge. Completion is invalid without evidence.

## `/yolo`

Enables no-confirmation mode for normal safe work. It does not disable hard safety blocks. Unsafe actions are denied and the session continues with safer alternatives.

## `/loop`

Runs a fixed-iteration task loop. It cannot run at the same time as `/goal` because goal mode owns completion criteria.

## `/swarm`

Schedules five bounded subagents from available providers/models. The main agent keeps control, and every subagent must produce evidence that can be logged.

## `/export`

Writes JSONL session output into the active workspace with session ID and timestamp. It should include errors, harness actions, model usage, token usage, provider fallbacks, denied commands, subagent results, judge verdicts, and completion evidence.
