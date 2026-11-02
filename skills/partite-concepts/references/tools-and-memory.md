# Tools, skills, and memory

## Different ways to add capability

| Capability | What it contributes |
|---|---|
| MCP Connection | External tools that an Agent can invoke |
| Particle | Sandboxed WebAssembly tools attached to an Agent |
| Client Task | A structured question or task sent to the client while execution waits |
| Runtime Skill | Instructions and supporting text files loaded by a Mesh Agent when needed |
| Memory Store | Stored documents, searchable chunks, facts, and observations |
| Interaction Profile | The Agent's rules for reading and writing memory and supplying history |

Tools perform operations or retrieve data. Skills supply instructions and
reference knowledge. Attaching a skill does not create a tool endpoint or grant
access to an external service. Tools and skills can complement one another.

Agent tool filters control which tools are exposed from an attached connection
or Particle. A tool being available does not prove it was called; inspect the
trace when the question concerns actual execution.

## Two meanings of skill in this repository

These `partite-*` packages run in a developer's coding assistant. Their CLI
instructions need the assistant's terminal tools when carrying out a workflow.
They do not automatically become Agent tools inside a running Mesh.

Mesh also has its own Skill resource. Its Agent sees attached skill names and
descriptions, then loads relevant instructions with `load_skill` and supporting
files with `read_skill_file`. The published runtime skill format permits UTF-8
text files and rejects a `scripts` folder. Do not assume that a developer skill
with executable helpers can be uploaded unchanged for runtime execution.

## Memory and conversation history

The Mesh selects at most one Memory Store. Agents attach Interaction Profiles
that define how they use it. A Profile without a store on the Mesh does not
supply that memory. An Embedding Profile supports semantic indexing when used.

The memory pipeline can ingest conversation turns, index chunks, extract facts,
and consolidate observations. Some stages run asynchronously. Request tags and
configured scopes select the memory slice, such as a particular user's data.
Tag conventions are part of the application's memory setup.

A Conversation identifies a session and contains Requests; it does not by itself
mean every Agent receives all prior turns in its prompt. Conversation history
is configured through an Interaction Profile, with conversation ingestion in
the Memory Store. Inspect injected history and memory in traces to see what the
Agent actually received. Store contents can persist across Conversations.

Example: `user=alice` can select Alice's memory slice when scopes and Profiles
are configured for that tag. Adding the tag alone does not configure memory or
prove that data isolation rules are correct.

## Published references

- [Agent tools and memory](https://docs.partite.ai/configuration-reference/agents/)
- [Runtime Skills](https://docs.partite.ai/configuration-reference/tools/skills/)
- [Memory guide](https://docs.partite.ai/guides/memory/)

Checked 2026-10-30. Consult the linked Agent page for MCP, Particle, Client Task,
and Interaction Profile references instead of guessing their configuration.
