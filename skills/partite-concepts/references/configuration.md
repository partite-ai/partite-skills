# Configuration and versions

## How the parts fit

```text
Model Provider -> Model Profile <- Model Set rules
                                       |
                                       v
Client -> Intent -> Agent version in a Mesh -> response
                         |
                         +-> Call Link -> another Agent -> result back
                         +-> Transfer Link -> successor Agent continues

Route -> selects a Mesh version when a Conversation starts
```

The mesh owns its graph and Model Set selection. Agents own their behavior and
input schemas. The Model Set chooses a Model Profile for an agent invocation.

| Concept | Meaning |
|---|---|
| Model Provider | Connection and credentials for an inference service |
| Model Profile | A model and its inference parameters under a Provider |
| Model Set | Ordered rules that select a Profile from Agent complexity and tags, plus a default |
| Agent | A focused role with instructions, input schema, tools, and optional memory behavior |
| Mesh | A graph of selected Agent versions and their typed connections |
| Intent | A named external entry point into a selected Agent, with defined response types |
| Call Link | A source Agent invokes another Agent and receives its result |
| Transfer Link | A source Agent hands control to a successor instead of awaiting a return |
| Route | A traffic split that selects the Mesh version for a new Conversation |

Model Set rules use the first match, so their order matters. Their default
Profile applies when none matches. Agent complexity is a routing signal, not
a direct model ID. Model Sets can use Profiles from different Providers.

An Intent's input schema comes from its target Agent version. Intent response
schemas describe final outputs. A Call Link uses the target Agent's input
schema and declares the results that can return to the caller. A transferred
Agent continues with the response tools available to the source.

Example: a support Agent calls a billing Agent for a calculation, then writes
the reply itself. Use a Transfer when the billing Agent should take over the
conversation instead. A Link exposes a tool the source Agent can choose to use;
its presence does not prove it ran on every request.

## Version boundaries

Behavior edits create Agent versions. Name and description edits do not. Meshes
select Agent versions; a newer Agent version does not automatically replace
every existing selection. The dashboard can offer to update mesh drafts when
saving. The CLI has an explicit Agent-version selection operation.

Mesh edits go to a mutable draft. Publishing labels a snapshot and opens a new
draft. An edited draft does not rewrite a published graph. New conversations
select a version explicitly or through a Route; existing conversations keep
their selection. A Route chooses by weight per conversation, not per model call.
Selecting a draft is for mutable development configuration, not a frozen release.

Published graph versions do not freeze every external dependency. For example,
the Mesh Memory Store association is not versioned, and mesh versions can share
that store. Do not promise that publishing freezes tools, credentials, or data.

In CLI specs, mesh entry points are named `inputs`; the runtime calls them
Intents. CLI names can resolve to IDs, but that convenience does not imply that
the HTTP API accepts names wherever it expects IDs.

## Published references

- [Models](https://docs.partite.ai/configuration-reference/models/)
- [Model Sets](https://docs.partite.ai/configuration-reference/models/sets/)
- [Agents](https://docs.partite.ai/configuration-reference/agents/)
- [Meshes, Links, versions, and Routes](https://docs.partite.ai/configuration-reference/meshes/)

Checked 2026-10-30. Verify exact label editing rules and Route weight formats in
the interface or API being used; conceptual traffic shares are not a wire schema.
