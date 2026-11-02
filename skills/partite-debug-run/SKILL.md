---
name: partite-debug-run
description: Diagnose failed, slow, or incorrect Partite mesh runs through CLI conversation state, events, and traces. Use for a specific run or runtime symptom, with evidence before configuration changes.
---

# Diagnose a mesh run

Use this skill from a developer's coding assistant with terminal access. It is
self-contained. Start with observation; a request to diagnose does not imply
permission to replay external tool actions or change production configuration.

For the trace model, consult [Traces](https://docs.partite.ai/configuration-reference/observability/traces/).
For request states, consult the [Conversation API guide](https://docs.partite.ai/guides/conversation-api/).

Check `partite --version`, `context list -o json`, and `context show -o json`.
If absent, install with
`go install github.com/partite-ai/partite-cli/cmd/partite@latest` using a compatible
Go toolchain. Resolve environment URL overrides without showing secrets. Use an
explicit context and `-o json` for server commands.

## Diagnosis

1. Collect the conversation ID, request ID if known, and observed symptom.
   Read [inspection commands](references/inspection.md). If no ID is supplied,
   find candidate traces for the named mesh and time range; do not choose an
   unrelated run when more than one matches.
2. Read the conversation and requests. Separate a failed request from a timeout,
   a request still running, a task waiting for input, or a cancelled request.
3. Read the trace. Follow the affected request through Agent hops, Link calls,
   model calls, and tool calls. Find the first relevant error or divergence.
   Keep trace IDs and runtime request IDs distinct.
4. Check the selected mesh version, Agent versions, input schema, and Model Set.
   Compare against that version, not just today's draft. Separate observed facts
   from suspected causes.
5. Recommend the smallest correction. When repair is in scope, change the draft
   and run a bounded test with safe input. New configuration needs a new
   conversation for a valid comparison. Do not replay side effects without the
   user's authorization.

## Completion

Return the affected conversation/request/version, observed state, evidence,
likely cause, correction, and verification result. If traces were disabled or
access was denied, state that limit. Never infer successful execution from
missing trace data. Keep credentials, private inputs, and full prompt text out
of the report unless the user needs and has requested that detail.
