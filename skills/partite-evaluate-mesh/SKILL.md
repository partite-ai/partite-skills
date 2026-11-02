---
name: partite-evaluate-mesh
description: Set up or inspect Partite Enterprise mesh evaluations using metrics, rubrics, monitors, and version reports through the CLI. Use for measured mesh quality checks; requires an Enterprise server.
---

# Evaluate an Enterprise mesh

Use this skill in a developer's coding assistant with terminal access. It is
self-contained. Evaluation needs Mesh Enterprise; local Mesh configuration and
runtime support alone do not provide this feature.

The [Evaluations guide](https://docs.partite.ai/guides/evaluations/) explains
scoring, coverage, and the distinction between hosted Monitors and on-demand
runs on completed traces, including traces sent from self-hosted engines.

## Establish the evaluation

Check `partite --version` and the target context. If absent, install with
`go install github.com/partite-ai/partite-cli/cmd/partite@latest` using a compatible
Go toolchain. Resolve environment URL overrides without exposing credentials.
Use explicit `--context NAME` and `-o json` for server commands.

Read [evaluation workflow](references/evaluation.md). Check Enterprise access
with a read-only evaluation command. Do not mark a server Enterprise just to
bypass an unsupported-feature error. Distinguish unavailable features from
missing permission.

Identify the mesh/version, representative cases, rubric, judge Model Set, and
pass criteria. Inspect available metrics and existing monitors before creating
anything. Reuse suitable resources. Do not invent scoring thresholds or claim
quality based on latency alone.

## Execute and inspect

For report requests, inspect existing results without enabling a monitor. For
new evaluations, agree on the sample scope and model cost within the user's
request. Monitors score new conversations in the background; use an isolated
test mesh for a bounded trial when available.

Use a full sample only on an isolated test target when full coverage is needed.
Do not enable full sampling on production as an automatic default. Preserve
existing monitor settings. After an authorized temporary test, disable the
monitor created for that test and report its final state.

Inspect evaluation status, each metric result, failed measurements, sample
counts, and version IDs. Wait within a bounded deadline for asynchronous work.
Missing or failed measurements are not zero scores or successful checks.

## Completion

Return the rubric and metric choices, selected mesh versions, sample scope,
completed and failed counts, scores, agreed pass criteria, monitor state, and
limits of the comparison. Do not publish a version or change Routes as a side
effect of evaluation. If Enterprise is unavailable, explain the limit and offer
explicit response checks without calling them Enterprise evaluations.
