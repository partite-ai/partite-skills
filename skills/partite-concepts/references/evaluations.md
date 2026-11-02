# Enterprise evaluations

Evaluation is an Enterprise capability. A standalone Mesh engine can execute
and trace a request without providing the Enterprise evaluation service.
Availability still depends on the target deployment and account access.

| Concept | Purpose |
|---|---|
| Computed evaluator | Measures execution, such as latency or tool errors |
| Judge evaluator | Uses a model to assess quality or safety against instructions |
| Rubric | Combines metric definitions, normalization, and relative weights |
| Monitor | Samples live hosted execution for ongoing evaluation |
| Evaluation run | Scores completed trace Requests on demand |
| Version report | Compares measurements and scores across Mesh versions |

A judge Model Set selects evaluator models separately from the Mesh's execution
Model Set. Judge trials add model work and can change evaluation cost. A Rubric
combines normalized metrics; a raw latency value is not itself a quality score.
Zero-weight metrics can be measured without contributing to the combined score.

The published guide distinguishes hosted Monitors from on-demand runs. Monitors
score hosted cloud execution. On-demand runs can also score completed traces
sent from a self-hosted engine. Explain this distinction before suggesting a
Monitor for a self-hosted deployment.

A trace explains what happened. Evaluation assesses that execution against
chosen criteria. It does not replace tracing or make an incomplete Request a
successful outcome. A high score needs context: which version, Rubric, judge
configuration, cases, and fraction of traffic were measured. Missing or failed
measurements are not passing results.

The CLI uses terms such as `judge-metric`, `computed-metric`, and `trace-run`.
The dashboard documentation also uses evaluator and evaluation run. Match the
terms to the user's interface; do not invent a CLI creation command because a
run can be started in the dashboard or through an Enterprise REST endpoint.

## Published references

- [Evaluations guide](https://docs.partite.ai/guides/evaluations/)
- [Evaluation configuration and on-demand runs](https://docs.partite.ai/configuration-reference/observability/evaluations/)

Checked 2026-10-30. Verify current metric keys, API contracts, and deployment
support before turning an explanation into executable setup instructions.
