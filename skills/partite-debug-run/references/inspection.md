# Inspect state and traces

Use actual IDs and the established context. These commands are read-only:

```sh
partite --context cloud conversation get CONVERSATION_ID -o json
partite --context cloud conversation requests CONVERSATION_ID -o json
partite --context cloud conversation events CONVERSATION_ID REQUEST_ID -o json
partite --context cloud conversation response CONVERSATION_ID REQUEST_ID -o json
partite --context cloud conversation task get CONVERSATION_ID REQUEST_ID -o json
partite --context cloud trace get CONVERSATION_ID -o json
```

`trace get` accepts a trace ID or conversation ID. A conversation can also be
addressed by its external ID through CLI conversation commands. When searching:

```sh
partite --context cloud trace list --mesh support --since 1d -o json
partite --context cloud mesh versions support -o json
partite --context cloud mesh version get support v1 -o json
```

Trace list status values are `active`, `cancelled`, and `complete`; these are
not the request state values. CLI request states include `running`, `waiting`,
`completed`, `failed`, and `cancelled`.

The trace contains `id`, `conversation_id`, mesh information, usage, and
`requests`. Trace requests have both `id` (trace request ID) and `request_id`
(runtime request ID). Use runtime IDs for conversation endpoints and trace
request IDs for evaluation filters. Inspect hop nesting rather than assuming
array order alone proves causality.

## Read the symptom

| Evidence | Next check |
|---|---|
| CLI exit 3 or 4 | Authentication, selected organization, scopes, and environment overrides |
| Invalid input or schema error | Intent schema from the conversation's selected version |
| Missing resource | Context URL, IDs, selected version, and retained resources |
| `waiting` or CLI exit 9 | Task kind and response schema; distinguish input from tool approval |
| CLI exit 10 | Existing request state; the client deadline does not cancel server work |
| Provider error | Selected Model Profile, provider credential presence, supported model parameters |
| Tool error | Tool arguments, connection state, filters, and access decision |
| Wrong answer | Actual input, selected Agent version, instructions, tool outputs, response schema |
| No trace | Trace access and whether conversation tracing was disabled |

CLI exit codes: 0 success; 1 general error; 2 usage or confirmation needed;
3 authentication; 4 permission; 5 not found; 6 conflict; 7 mesh request failed;
8 mesh request cancelled; 9 waiting for task; 10 wait timeout; 130 interrupted.
With `-o json`, CLI errors on stderr have an `error` object containing message,
code, hint, and optional param. Do not confuse this wrapper with REST problem
responses or a failed runtime request's `error` field.

Use `conversation response ... --wait --timeout 1m -o json` to wait within a
bounded interval when requested. Use `conversation events ... --follow` only
when live observation is needed. An interrupt can stop following without
cancelling the server request.

## Repair boundaries

Task completion, task cancellation, request cancellation, resource updates, and
replays are mutations. Do not perform them merely to inspect state. If a repair
was requested, use `conversation task complete --help` and the observed
`response_schema` before submitting an answer. Keep the user's approval decision
separate from technical JSON validation.

An Agent edit creates a version but does not automatically select it in the
mesh draft. Inspect pins and use `mesh set-agent-version` for an intended update.
Published mesh versions and existing conversations do not change when a draft
is edited. Report a draft-only fix as a draft-only fix.
