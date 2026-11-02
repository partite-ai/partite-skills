---
name: partite-build-mesh
description: Create or extend a Partite Mesh with typed agents, intents, links, and tools through the CLI. Use for mesh design and draft configuration changes after connection setup.
---

# Build or extend a mesh

Use this skill in a developer's coding assistant with terminal access. It is
self-contained and does not require another Partite skill.

For concept questions, consult [Meshes](https://docs.partite.ai/configuration-reference/meshes/)
and [Agents](https://docs.partite.ai/configuration-reference/agents/). Use
[Message Schemas](https://docs.partite.ai/message-schemas/) for the supported
schema subset and grounding rules.

## Establish the target

Check `partite --version`, `context list -o json`, and `context show -o json`.
If the CLI is absent, install with
`go install github.com/partite-ai/partite-cli/cmd/partite@latest` using a compatible
Go toolchain. Check environment URL and credential overrides without printing
secrets. Use explicit `--context NAME` and `-o json` for server commands.

Inspect the target mesh, its versions, referenced agents, and Model Set before
editing. Ask for the intended behavior only when the request does not establish
it. Define entry intents, input and response schemas, tool needs, and the
conditions for each agent handoff. Add agents only when their separate
instructions or tool access help the task.

## Implement the draft

Read [configuration and version rules](references/configuration.md) before
changing a graph. Use the exported CLI spec as the starting point for existing
resources. Prefer targeted updates for small changes. Keep unrelated values.

Keep Agent input schemas and Intent/Link response schemas explicit. A model
profile belongs to a Model Provider; a Model Set chooses profiles; the mesh
selects the Model Set. Do not treat a model ID as a Model Set ID.

For tool connections, inspect command help and the existing resource first.
Agent tool flags can replace associations. Preserve existing tool filters and
associations unless the user asked to change them. Secrets stay in environment
variables or secret storage, outside configuration exports.

## Verify

Run representative inputs on `--draft`, including an input for each changed
intent or link. Check final responses and traces, not just resource creation.
Calls can incur model charges or tool effects; keep test actions within the
user's request. Do not replay a side-effecting operation just to obtain a trace.

Publish only when requested, with an unused label. Publishing and changing
production Routes are separate actions. Return the changed resources, version
selection, test results, and any remaining errors. Do not claim that existing
conversations moved to the new configuration.
