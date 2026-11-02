# Conversation REST API

These examples follow Mesh Conversation API v2. Use the target server's base
URL. Cloud uses `https://api.partite.ai`; a local server commonly uses
`http://localhost:8080`. Paths below have no `/v1` prefix. The source contract is
`pkg/rest/conversation_api.yml` in the Mesh repository. Use the deployed version's
contract for capabilities beyond this reference.

## Authentication and input

Authenticated calls use `Authorization: Bearer TOKEN`. JSON calls send
`Content-Type: application/json` and `Accept: application/json`. A local server
configured without auth does not require a cloud token.

Start a conversation with `POST /conversations`. Select exactly one target:

```json
{"mesh":{"id":"MESH_ID","version_label":"v1"}}
```

Inside `mesh`, choose exactly one of `version_label` (string), `version`
(integer ordinal), or `draft: true`. Alternatively use
`{"route_id":"ROUTE_ID"}` instead of `mesh`. Use real IDs, not CLI names.

To start the first request at the same time, add `intent` (string) and `input`
(object matching the discovered schema). Optional conversation fields are
`external_id` (unique string, 1–255 bytes), `tags` (array of objects with string
`name` and `value`), `notification_channel` (string), and `tracing` (boolean,
default true). An external ID is immutable; a duplicate returns HTTP 409.
Do not treat external IDs as automatic POST idempotency keys.

For a mesh with a `chat` intent accepting `{text: string}`, an example request:

```sh
curl --fail-with-body --silent --show-error \
  "$PARTITE_URL/conversations?wait=50" \
  -H "Authorization: Bearer $PARTITE_API_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  --data @request.json
```

`request.json` contains actual selected IDs and this shape:

```json
{
  "mesh": {"id": "MESH_ID", "version_label": "v1"},
  "intent": "chat",
  "input": {"text": "Say hello in one sentence."}
}
```

The example's IDs must be replaced before execution. Application code must use
its JSON encoder, not construct JSON by concatenating user input.

## Responses and state

Creation returns HTTP 201 with a conversation object. `id` is the conversation
ID. When an initial input was sent, `request` is the started request object.
It contains `id`, `status`, `input`, `tags`, and `created_at`, plus fields for
its current state. Save both IDs.

| Request status | Client action |
|---|---|
| `running` | Wait or follow the existing request |
| `waiting` | Inspect `tasks`; collect an answer or approval |
| `completed` | Read `response` using the intent response schema |
| `failed` | Report `error`; do not blindly retry the original POST |
| `cancelled` | Report cancellation |

`GET /conversations/{conversation_id}/requests/{request_id}?wait=50` returns the
request object. `wait` is at most 50 seconds per HTTP call; it is not an overall
client deadline. Poll within the application's bounded deadline. On expiry,
return a pending state with saved IDs, or cancel explicitly if that is the
application's defined behavior.

Send another message with
`POST /conversations/{conversation_id}/requests`, body
`{"intent":"chat","input":{"text":"A follow-up question"}}`. It returns a
request object. `intent` is required when more than one is available.

Discover schemas with `GET /conversations/{conversation_id}/intents`; the body
is `{intents: [...]}`. Each intent has a `name`, `schema`, and `responses` array.
Each response has a `name` and `schema`, with optional descriptions.

Errors use `application/problem+json`. Check HTTP status and retain useful
problem details, including `param` when present. Do not decode a problem as a
successful conversation. Separate authentication errors, denied operations,
invalid inputs, and conflicts in the app's error handling.

## Tasks and cancellation

A waiting request lists `tasks`. Each task has `id`, `kind`, `response_schema`,
and kind-specific `client_task` or `tool_call_approval` details.

- Complete with `POST /conversations/{conversation_id}/requests/{request_id}/tasks/{task_id}/complete`.
  The body is the answer itself, matching `response_schema`, without a wrapper.
- Decline with `POST /conversations/{conversation_id}/requests/{request_id}/tasks/{task_id}/cancel`.
  Declining a task lets the requesting agent handle the refusal; it does not
  cancel the whole request.
- Cancel a request with `POST /conversations/{conversation_id}/requests/{request_id}/cancel`.

## Streaming and reconnects

If required, request `Accept: text/event-stream` or `application/jsonl` on the
message POST. Resume with
`GET /conversations/{conversation_id}/requests/{request_id}/events` and
`Last-Event-ID`. Do not repeat the POST to reconnect. Handle message, thinking,
task, task_response, response, error, and cancelled events. The final events are
response, error, and cancelled; a task pauses execution and is not completion.
Do not log thinking events by default.

Use JSON polling first when the app has no streaming requirement. For file
attachments, resource references, and notification channels, consult the server
contract rather than inventing encodings or endpoints.
