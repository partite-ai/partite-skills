# Conversations, requests, and traces

## The execution model

A Conversation is the runtime session on a selected Mesh version. A Request is
one input sent to an Intent within that Conversation. An Agent can make several
model calls or invoke tools and other Agents while handling one Request. These
are different units; one user message is not necessarily one model call.

```text
Conversation on a selected Mesh version
  Request -> Agent and tool execution -> response
  Request -> execution -> waiting task -> execution -> response
  Trace   -> entries that record the Requests and their execution
```

The Conversation API accepts a Mesh ID with one explicit selector: version
ordinal, version label, or `draft: true`. Alternatively, a Route ID selects the
version. A draft is not selected implicitly. Further messages use the existing
Conversation rather than selecting a new version each turn.

| Request state | Meaning |
|---|---|
| `running` | Execution is in progress |
| `waiting` | Execution needs a task answer or decision |
| `completed` | A final response is available |
| `failed` | Execution ended with an error |
| `cancelled` | The Request was cancelled |

A task can request client input or tool approval. Waiting is not success or
failure. The task's response schema defines the answer. Declining a task and
cancelling the whole Request are different operations. A client wait timeout
does not mean server execution stopped. Reconnecting to a stream follows the
existing Request; repeating its message POST starts new work.

## Typed messages and resources

Message schemas describe Agent inputs and Intent or Call Link responses. Mesh
uses a limited JSON Schema subset: root objects, described properties, and
simple structures. Do not assume support for every JSON Schema combinator.

A resource is content held through a reference, such as an attached file or a
tool-provided document. Passing its reference differs from putting the full
content into every prompt. The schema format `x-partite-ref` marks resource
fields; it does not make the resource persistent cross-conversation memory.

Grounding can require citations for selected response fields, and can validate
cited text. It connects an output to available evidence; it is not a promise
that any arbitrary answer is true. Schema validation, grounding, and evaluation
answer different questions: shape, evidence, and measured quality.

## Traces

A traced Conversation has a trace with entries for its Requests. A trace shows
what ran: selected versions, Agent hops, model and tool use, supplied context,
and outputs. It is execution evidence, not the editable mesh definition.

Tracing is on by default. A Conversation created with tracing disabled does not
have the same recorded evidence. Absence of a trace does not establish absence
of execution. Conversation IDs, runtime Request IDs, trace IDs, and trace
Request IDs identify different records and are not interchangeable.

## Published references

- [Conversation API guide](https://docs.partite.ai/guides/conversation-api/)
- [Message Schemas](https://docs.partite.ai/message-schemas/)
- [Traces](https://docs.partite.ai/configuration-reference/observability/traces/)

Checked 2026-10-30. Use the deployed API contract for full fields, task schemas,
stream event encodings, and endpoint-specific error handling.
