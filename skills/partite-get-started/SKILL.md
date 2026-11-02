---
name: partite-get-started
description: Install the Partite CLI, connect to local Mesh or Partite cloud, and build and run a first mesh. Use for new developer setup and first-response onboarding.
---

# Build and run a first mesh

This skill is for a developer's coding assistant with terminal access. It does
not install skills on agents inside Mesh. It works without any other skill.

For setup concepts, consult [Getting Started](https://docs.partite.ai/getting-started/)
and [Models](https://docs.partite.ai/configuration-reference/models/). The local
references below contain the steps needed for this workflow.

## Workflow

1. Read [installation and access](references/installation-and-access.md).
   Check for `partite`. If it is absent, install it and check that it runs.
   Establish the server and organization before creating resources.
2. Inspect Model Providers, Model Profiles, Model Sets, Agents, and Meshes with
   `list -o json`. Reuse suitable resources. An empty list means no resources;
   an access error does not. Ask only for missing choices that affect the setup,
   such as the user's provider or the purpose of the mesh.
3. Read [first mesh](references/first-mesh.md). Use its complete sample or adapt
   it to the user's task. Keep the first run small: one agent and one intent.
4. Run the draft with explicit JSON input and `-o json`. A completed response
   proves execution; creating configuration alone does not. Inspect the trace.
5. If a stable version is part of the request, publish with a new label and test
   a new conversation on that label. Do not change Routes as part of onboarding.

## Execution rules

- Use an explicit context on server commands. Check for environment overrides.
- Prefer CLI commands over direct GraphQL during setup. Read command `--help`
  if the installed CLI differs from the examples.
- Read provider credentials from environment variables or secret input. Never
  ask the user to paste credentials into chat or put them in tracked files.
- Apply the user's task authorization. Do not replace existing resources to
  make sample names available. Use unique names when sample names are taken.
- If a step partly succeeds, inspect server state before retrying. Keep the
  created IDs. Do not delete working resources as automatic cleanup.
- Model calls can incur charges. Use only the provider and scope the user chose.
  Do not enable background monitors during first-run setup.

## Completion

Return the context, resource names and IDs, selected version, conversation ID,
request ID, observed response, and a command to run again. State any missing
credential, server, or model access that prevented a real run. Never report a
successful run based only on command help or configuration creation.
