# Detector Modules

Faultline now supports multiple deterministic detector modules behind a shared
result model.

The bundled catalog currently splits into 98 log playbooks and 3 source
playbooks.

## Built-in detectors

- `log`: the existing CI log matcher based on `match.any`, `match.all`, and
  `match.none`, with IDF-weighted conflict resolution preserved.
- `source`: a source-aware detector that interprets trigger evidence,
  amplifiers, mitigations, suppressions, context, and change hints.

Use `log` playbooks when the failure signature is visible in build output and
should be diagnosable from a CI log alone.

Use `source` playbooks when the risk is visible from repository structure or
code paths before a CI log exists. In the shipped CLI, `inspect` and `guard`
are the main source-detector surfaces.

## Source playbook schema

Source playbooks declare reusable primitives instead of bespoke code:

```yaml
detector: source
source:
  triggers: []
  amplifiers: []
  mitigations: []
  suppressions: []
  context: []
  compound_signals: []
  local_consistency: []
  path_classes: []
  change_sensitivity: {}
  safe_context: []
scoring: {}
```

## Scoring flow

The source detector computes a final score from:

```text
base signal
+ compound bonus
+ blast radius bonus
+ hot path bonus
+ change bonus
- mitigation discounts
- suppression discounts
- safe-context discounts
```

The output keeps the full evidence split for explainability:

- triggers
- amplifiers
- mitigations
- suppressions
- context

Source playbooks currently live under `playbooks/bundled/source/`. Use
`faultline inspect .` to exercise the full source detector against a repository
tree and `faultline guard .` when you only want quiet, high-confidence local
prevention findings. When the inspected path lives inside a git worktree,
`inspect` and `guard` also use the local diff when available so changed files
and line-level edits can be scored as introduced or modified rather than only
as legacy repository risk. Positive and mitigated repository fixtures for the
shipped source rules live under `internal/engine/testdata/source/` and are
validated in `internal/engine/source_playbook_fixtures_test.go`.
The repository scan skips dependency trees such as `.git/`, `node_modules/`,
`vendor/`, `.venv/`, and `venv/` so `inspect` and `guard` stay focused on the
repository's own source rather than bundled tooling copies.

Shipped source playbooks should also carry deterministic `workflow.likely_files`,
`workflow.local_repro`, and `workflow.verify` hints so source findings hand off
cleanly to the next maintainer or agent.

## Suppressions

Supported suppression styles are explicit and auditable:

- inline comments such as `faultline:ignore <playbook-id>`
- playbook-defined path suppressions
- playbook-defined pattern suppressions

Suppressions are reported in structured output and influence scoring rather
than silently disappearing findings.
