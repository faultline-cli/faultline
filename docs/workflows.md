# Workflows

Faultline workflows are the typed remediation layer that sits between a matched
failure artifact and an explicit, auditable fix path.

They are intentionally constrained:

- static authored YAML only
- strict `workflow.v1` schema
- typed steps with registry-backed validation
- no arbitrary expression language
- no shell interpolation
- no freeform generation in the execution path

This keeps remediation reviewable, replayable, and safe for both engineers and
automation agents.

## Core Model

- `Workflow Definition`: a static `workflow.v1` YAML document
- `Bound Workflow`: a definition plus resolved inputs from a failure artifact
- `Execution Plan`: the ordered concrete steps Faultline will run
- `Execution Record`: the persisted result of an applied workflow

## CLI

- `faultline workflow <logfile>`
  Compatibility surface that preserves the legacy follow-through checklist.
- `faultline workflow explain <log-or-artifact>`
  Resolve the recommended typed remediation workflow and explain it.
- `faultline workflow apply <log-or-artifact> --dry-run`
  Resolve bindings and render the concrete execution plan without mutation.
- `faultline workflow apply <log-or-artifact> --allow-environment-mutation`
  Apply the workflow under explicit policy flags.
- hidden `faultline workflow show <execution-id>`
  Show a persisted execution record.
- hidden `faultline workflow history`
  List persisted workflow executions.

`faultline workflow explain` and `apply` accept either a raw log or a saved
analysis artifact JSON document. When a log is supplied, Faultline analyzes it
first and then selects the recommended remediation workflow from the artifact.

## Definition Schema

Example:

```yaml
schema_version: workflow.v1
workflow_id: missing-executable.install
title: Install missing executable
description: Install the missing executable detected by the failure artifact and verify it is available in PATH.
inputs:
  missing_executable:
    type: string
    required: true
steps:
  - id: detect-manager
    type: detect_package_manager
    args: {}
  - id: install-tool
    type: install_package
    args:
      package: ${inputs.missing_executable}
      manager: ${steps.detect-manager.manager}
verification:
  - id: verify-tool
    type: which_command
    args:
      command: ${inputs.missing_executable}
    expect:
      found: true
```

Rules:

- `schema_version`, `workflow_id`, `title`, `description`, and `steps` are required
- inputs are typed and currently support `string`
- verification is mandatory, either through explicit `verification` steps or
  step-level expectations
- unknown top-level fields are rejected
- step args and expectations are validated against the step registry

## Variable Binding

`workflow.v1` supports only explicit interpolation:

- `${inputs.<name>}`
- `${steps.<step-id>.<output>}`
- `${artifact.facts.<name>}`
- `${artifact.fingerprint}`
- `${artifact.matched_playbook.id}`
- `${runtime.workdir}`
- `${runtime.repo_root}`

There is no expression language, branching DSL, or loop construct in v1.

## Supported Step Types

Current shipped step set:

- `which_command`
- `file_exists`
- `detect_package_manager`
- `install_package`
- `noop`
- `fail`

The registry is extensible, but new step types should remain typed,
deterministic, and narrow.

## Safety Model

Each step resolves to one safety class:

- `read`
- `local_mutation`
- `environment_mutation`
- `external_side_effect`

Policy is explicit:

- `explain` never executes steps
- `dry-run` may use safe read probes to concretize later step bindings
- `apply` requires explicit allow flags for non-read safety classes

For example, `install_package` is `environment_mutation`, so
`faultline workflow apply` will block it unless
`--allow-environment-mutation` is passed.

## Playbook References

Playbooks can recommend one or more typed workflows:

```yaml
remediation:
  workflows:
    - ref: missing-executable.install
      inputs:
        missing_executable:
          from: artifact.facts.missing_executable
```

Faultline resolves those bindings into the failure artifact so saved analysis
artifacts remain sufficient for later `workflow explain` or `workflow apply`
runs.

## Store Records

Applied workflows are persisted in the local store as execution records. Each
record includes:

- execution id
- workflow id and title
- source failure fingerprint and failure id
- mode
- resolved inputs
- per-step results
- verification status
- final status

Maintainers and compatibility paths can use the hidden `faultline workflow
show` and `faultline workflow history` commands to inspect them.

## Non-Goals

- arbitrary scripting
- general shell wrappers
- hidden LLM planning
- rollback orchestration
- remote control plane features

The workflow layer should stay deterministic, local-first, explicit, and easy
to audit.
