# First mesh through the CLI

Examples use `--context cloud`. Replace it with the established context on every
command. Names beginning with `onboarding-` are samples: check that they are free
or choose unique names. Run each step separately and check its result.

## Reuse or configure models

```sh
partite --context cloud model-provider list -o json
partite --context cloud model-profile list -o json
partite --context cloud model-set list -o json
partite --context cloud agent list -o json
partite --context cloud mesh list -o json
```

A Model Provider connects to the inference service. A Model Profile selects a
model and parameters. A Model Set selects profiles for agent requests, with a
default profile and optional rules. Use a suitable existing Model Set if present.

If the user chose Anthropic and supplied `ANTHROPIC_API_KEY` in the environment:

```sh
partite --context cloud model-provider create onboarding-provider --type anthropic --api-key-env ANTHROPIC_API_KEY -o json
partite --context cloud model-provider models onboarding-provider -o json
```

Other providers have different credentials. Check `model-provider create --help`
and `model-provider capabilities TYPE -o json` for the selected provider. Do not
switch providers just to fit this example. Choose an accessible model ID from
the returned list. Set `PARTITE_MODEL_ID` to that ID, then create a profile:

```sh
partite --context cloud model-profile create onboarding-profile --provider onboarding-provider --model "$PARTITE_MODEL_ID" -o json
partite --context cloud model-set create onboarding-models --default onboarding-provider/onboarding-profile -o json
```

Inspect required model parameters and add them if needed. Do not assume a model
name or token limit from a different provider.

## Create the agent and mesh

Copy these files to the user's chosen project directory. Change names and the
Model Set reference when reusing existing resources.

`agent.yaml`:

```yaml
name: onboarding-helper
description: A small agent for a first successful request
task_instructions: >-
  Answer the question in the input text. Keep the answer brief.
  Return the answer through the intent's answer response with a text field.
model_complexity: MEDIUM
input_message_type:
  type: object
  properties:
    text:
      type: string
      description: The user's question
  required: [text]
  additionalProperties: false
```

`mesh.yaml`:

```yaml
name: onboarding-mesh
description: A first callable mesh
model_set_id: onboarding-models
inputs:
  - name: chat
    agent_id: onboarding-helper
    responses:
      - name: answer
        description: The answer to the user's question
        message_type:
          type: object
          properties:
            text:
              type: string
          required: [text]
          additionalProperties: false
links: []
```

```sh
partite --context cloud agent create -f agent.yaml -o json
partite --context cloud mesh create -f mesh.yaml -o json
partite --context cloud mesh get onboarding-mesh -o json
partite --context cloud run onboarding-mesh --draft --intent chat --input '{"text":"Say hello in one sentence."}' -o json
```

The CLI calls mesh entry points `input`; the runtime calls them `intent`. The
intent's input schema comes from the attached Agent version. Its `responses`
define the agent's permitted final outputs.

## Check execution

Save `conversation_id` and `request_id` from the run result. Use the returned
conversation ID with `partite --context cloud trace get CONVERSATION_ID -o json`.
Check the actual answer against the task. The result status must be completed.
If the command returns a task (exit 9), inspect it; do not answer an approval
request on the user's behalf without authorization. Exit 10 means the wait
expired, not that the server request failed. Inspect the existing request before
starting another run. Do not repeat runs indefinitely after a model error.

## Optional stable version

Check `mesh versions onboarding-mesh -o json` for an unused label, then:

```sh
partite --context cloud mesh publish onboarding-mesh --label onboarding-v1 -o json
partite --context cloud run onboarding-mesh --label onboarding-v1 --intent chat --input '{"text":"Say hello in one sentence."}' -o json
```

Publishing freezes the draft under the new label and opens a new draft. A label
is used once per mesh. It does not change a Route. New conversations need an
explicit selector: `--draft`, `--label`, `--version`, or `--route`.
