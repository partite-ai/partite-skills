# Configuration and version rules

## Inspect before editing

Examples use a mesh named `support`. Use the actual context and resource names.

```sh
partite --context cloud mesh get support -o yaml
partite --context cloud mesh versions support -o json
partite --context cloud mesh input list support -o json
partite --context cloud mesh link list support -o json
partite --context cloud agent get triage -o yaml
partite --context cloud agent versions triage -o json
```

Save exports only to the requested working directory. `get -o yaml` produces
the CLI's spec format. `apply -f` reconciles a full spec; it is not a general
PATCH. For meshes, omitted `inputs`, `links`, or `boundary_ids` preserve those
collections; an empty list removes all entries. Other omitted fields can be
cleared by full apply, including memory settings. Preserve the full export when
applying an existing mesh. For a targeted change, prefer `update` or a subcommand.
Graph reconciliation is not an atomic transaction: inspect state after errors.

## Graph example

Create the referenced Agents and Model Set before this mesh. Agent names in
references resolve through the CLI; HTTP APIs require actual IDs.

```yaml
name: support
description: Answer questions and delegate billing analysis
model_set_id: support-models
inputs:
  - name: chat
    agent_id: triage
    responses:
      - name: answer
        description: Final reply to the user
        message_type:
          type: object
          properties:
            text: {type: string}
          required: [text]
          additionalProperties: false
links:
  - name: ask_billing
    description: Use for billing questions that need specialist analysis
    from_agent_id: triage
    to_agent_id: billing
    transfers_control: false
    responses:
      - name: finding
        description: Billing analysis returned to triage
        message_type:
          type: object
          properties:
            summary: {type: string}
          required: [summary]
          additionalProperties: false
```

The `billing` Agent's input schema defines what `triage` sends on the link.
Link responses define what comes back. `transfers_control: true` transfers
conversation control instead of a normal call and return. Choose it only when
the intended conversation flow needs that behavior.

For an Agent spec, use `name`, `task_instructions`, `model_complexity`, and
`input_message_type`. Example input schema:

```json
{"type":"object","properties":{"text":{"type":"string"}},"required":["text"],"additionalProperties":false}
```

Create with `partite --context cloud agent create -f agent.yaml -o json` and
`partite --context cloud mesh create -f mesh.yaml -o json`. Use `apply` only
after checking which existing resource it resolves and the full desired spec.

## Agent versions and tools

Agent changes create versions. A mesh can still point at an older Agent version.
Inspect the mesh draft's `agent_version_id`, `from_agent_version_id`, and
`to_agent_version_id`. After an intended Agent change, select its version:

```sh
partite --context cloud agent update triage --task-instructions @triage.md -o json
partite --context cloud mesh set-agent-version support triage -o json
```

`set-agent-version` selects the latest Agent version by default; pass
`--version AGENT_VERSION_ID` to pin a specific one. Updating a published mesh
requires a new mesh version; it does not rewrite the published graph.

For MCP tools, inspect `partite mcp --help` and `partite agent update --help`.
`--mcp-connection`, `--particle`, `--skill`, and similar association flags replace
their corresponding lists. Use the full retained list or a complete spec with
filters. A `--skill` association refers to a skill executed inside Mesh, not to
this developer skill package.

## Drafts and release checks

```sh
partite --context cloud run support --draft --intent chat --input '{"text":"Explain this invoice."}' -o json
partite --context cloud mesh publish support --label v2 -o json
partite --context cloud mesh version get support v2 -o json
```

Use a free label. Publishing freezes the current draft and opens a new one.
`mesh draft --from-version` copies version content and can replace current draft
work; do not use it as an automatic repair. Existing conversations retain their
selected mesh version. Test a new conversation to verify a new version.
