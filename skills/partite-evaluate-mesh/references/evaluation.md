# Enterprise evaluation workflow

## Inspect available resources

```sh
partite --context cloud eval judge-metric list -o json
partite --context cloud eval computed-metric list -o json
partite --context cloud eval rubric list -o json
partite --context cloud eval monitor list --mesh support -o json
partite --context cloud eval evaluation list --mesh support -o json
```

Judge metrics ask a model to score behavior. Computed metrics measure values
such as latency or cost. A Rubric combines weighted metrics. A Monitor samples
conversations for a mesh and scores requests against its Rubric. Its judge
Model Set is independent of the mesh's execution Model Set.

Use names and keys returned by discovery, not assumed built-in metric names.
Inspect definitions with `eval judge-metric get NAME -o json` and
`eval computed-metric get KEY -o json`. Use `eval rubric get NAME -o yaml` to
inspect weights and configuration. Rubric changes affect interpretation; retain
the rubric definition used for any comparison.

## Configure a test

Examples assume discovery confirmed a judge metric named `Helpfulness`, a
computed metric keyed `latency_ms`, a judge Model Set named `judges`, and an
isolated mesh named `support-test`. Substitute actual selected resources.

```sh
partite --context cloud eval rubric create support-quality --judge-metric Helpfulness=2 --computed-metric latency_ms=1 -o json
partite --context cloud eval monitor create --mesh support-test --rubric support-quality --model-set judges --sample-rate 1 --disabled -o json
partite --context cloud eval monitor enable support-test/support-quality -o json
```

A monitor is identified by its ID or `MESH/RUBRIC`; there is at most one per
mesh/rubric pair. `create` normally enables it; `--disabled` stages the setup.
Do not create a duplicate or overwrite an existing monitor to fit this example.
Sampling is in the range 0–1. Partial sampling does not guarantee every test
case receives an evaluation.

Run a small, named set of representative inputs on a chosen published label:

```sh
partite --context cloud run support-test --label candidate-v1 --intent chat --input '{"text":"Explain how to update my billing address."}' -o json
partite --context cloud eval evaluation list --mesh support-test --rubric support-quality -o json
```

Keep each run's conversation/request IDs, then use traces to map runtime request
IDs to trace request IDs. The evaluation filter `--request` expects the trace
request ID, not the runtime request ID. Inspect matching results rather than
assuming the newest evaluation belongs to the last submitted case.

After the temporary test, stop only the monitor created for it:

```sh
partite --context cloud eval monitor disable support-test/support-quality -o json
```

Do not change an existing monitor's enabled state as cleanup. If a failure
prevents cleanup, report the monitor ID and state.

## Results and version comparison

```sh
partite --context cloud eval evaluation get EVALUATION_ID -o json
partite --context cloud eval report support-test support-quality --since 7d -o json
partite --context cloud eval trace-run list --mesh support-test --rubric support-quality -o json
partite --context cloud eval trace-run get TRACE_RUN_ID -o json
partite --context cloud eval evaluation list --trace-run TRACE_RUN_ID -o json
```

The CLI can list and inspect on-demand trace runs, but it has no trace-run
creation command. Do not invent `eval run`, `eval evaluation create`, or
`eval trace-run create`. If the user needs on-demand scoring of existing traces,
use the available dashboard or inspect the deployed Enterprise GraphQL schema
before proposing a separate API implementation.

Evaluation results include overall status and optional `rubric_score`. Individual
results include `metric_name`, `metric_type`, `category`, nullable `score`,
optional `raw_value` and `unit`, `trial_count`, and optional measurement status,
reason, feedback, and error. Preserve null/missing values in summaries.

Reports group results by mesh version over a time range. Compare the same case
set, rubric definition, judge configuration, and sampling conditions. Separate
raw measurements from normalized scores. Small or unequal samples limit a
comparison; a score alone does not prove a release is better.
