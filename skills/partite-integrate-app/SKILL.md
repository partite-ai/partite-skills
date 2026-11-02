---
name: partite-integrate-app
description: Connect application code to an existing Partite mesh. Use the CLI to discover versions and intent schemas, then implement Conversation REST API calls, request states, and task handling.
---

# Connect an application to a mesh

Use this skill for application integration in a developer's coding assistant.
It is self-contained. It does not provision a new mesh or require the app to
invoke a CLI subprocess in production.

The [Conversation API guide](https://docs.partite.ai/guides/conversation-api/)
explains the runtime concepts and links to the full endpoint reference.

## Discover with the CLI

Check `partite --version` and the selected context. If absent, install with
`go install github.com/partite-ai/partite-cli/cmd/partite@latest` using a compatible
Go toolchain. Resolve environment URL overrides without exposing credentials.
Use explicit `--context NAME` and `-o json` for discovery.

Inspect `mesh get MESH`, `mesh versions MESH`, and any selected Route. Resolve
names to IDs for HTTP calls. Select the intended published version or Route.
Use a draft only for development when the user asks for it.

Create a conversation on that selector with `conversation create MESH --label
LABEL -o json`, then call `conversation intents CONVERSATION_ID -o json`.
Creating a conversation is a server mutation; reuse a suitable existing one for
read-only discovery when available. Read the actual intent and response schemas.
Do not infer the input shape from example text.

## Implement

Read [Conversation API](references/conversation-api.md). Use the application's
existing language and HTTP client. Use the CLI to verify a sample call first
when the requested task includes live execution. Calls can incur model charges
and external tool effects.

Keep API credentials on the server. Do not ship a configuration or conversation
API key in frontend code. Use the least privilege needed; `mesh:converse` is the
conversation scope. Trace access is separate. Credential provisioning must
follow the user's existing secret management process.

Handle `running`, `waiting`, `completed`, `failed`, and `cancelled` explicitly.
Use a bounded deadline. Preserve conversation and request IDs before waiting.
Resume a request after a network loss instead of repeating its POST. Do not
assume that a connection timeout cancelled server work.

For tasks, validate against `response_schema`. Obtain the required user input
or authorization before completing a task. Do not approve all tool calls as an
integration shortcut.

## Verify and deliver

Use focused tests for response decoding, HTTP errors, pending requests, tasks,
deadline expiry, and duplicate-POST prevention. Add streaming tests only if the
integration needs streaming. Run a live smoke test only with configured access
and authorization for the input and tool effects. Return the code changes,
required environment variables, selected mesh/version, and verification results.
