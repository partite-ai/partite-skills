---
name: partite-concepts
description: Explain Partite Mesh concepts and answer product questions about models, agents, intents, links, versions, conversations, tools, memory, and Enterprise evaluations. Use for definitions, comparisons, and how the parts fit together, without requiring CLI installation or server access.
---

# Explain Partite Mesh

Use this skill for questions such as "What is an Intent?", "Call or Transfer?",
"Why did my Agent edit not change the mesh?", or "Is a Conversation memory?".
It supplies product knowledge for a developer's coding assistant. It requires
no other skill, installed CLI, credentials, or running server.

## Select the relevant reference

- [Configuration and versions](references/configuration.md): Model Providers,
  Model Profiles, Model Sets, Agents, Meshes, Intents, Links, drafts, and Routes.
- [Conversations and traces](references/runtime.md): requests, tasks, typed
  messages, resources, grounding, and execution state.
- [Tools, skills, and memory](references/tools-and-memory.md): what each adds,
  where it is configured, and what persists across conversations.
- [Enterprise evaluations](references/evaluations.md): metrics, Rubrics,
  Monitors, evaluation runs, and version comparisons.

Read only the reference needed for the question. Explain the requested concept
first, then its relation to nearby concepts. Use a small example when it makes
the difference clearer. Retain product names such as Intent, Link, and Route.

## Documentation and accuracy

The public source is [Partite documentation](https://docs.partite.ai/). Reference
files contain short explanations and direct topic links, checked on 2026-10-30.
Use those links to verify changing capabilities or details beyond the reference.
Cite the relevant page when answering from published documentation. Do not
invent a documentation path, command, API field, or feature to fill a gap.

If published prose and an API example disagree, check the target version's
endpoint contract or installed CLI help before giving executable instructions.
For example, new conversations require an explicit version selector; do not
infer a default draft from older draft-testing prose. If live docs are
unavailable, explain from these references and state that current behavior was
not verified when that affects the answer.

## Keep explanation separate from execution

Answer knowledge questions without installing software, creating conversations,
or changing resources. A request to explain setup does not itself request setup.
If the user asks for implementation too, use available workflow skills when
helpful; these are optional, not prerequisites for answering the question.

Distinguish an Agent's model selection from a Route's mesh-version selection.
Distinguish conversation state from configured memory. Distinguish these
developer skill packages from skills loaded by agents running inside Mesh.
Describe Enterprise evaluation as an Enterprise capability, not a feature of
every standalone Mesh server. Do not infer customer access from a docs page.
