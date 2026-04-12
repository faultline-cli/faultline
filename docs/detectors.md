# Detector Modules

Faultline now supports multiple deterministic detector modules behind a shared
result model.

## Built-in detectors

- `log`: the existing CI log matcher based on `match.any`, `match.all`, and
  `match.none`, with IDF-weighted conflict resolution preserved.
- `source`: a source-aware detector that interprets trigger evidence,
  amplifiers, mitigations, suppressions, context, and change hints.

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

## Suppressions

Supported suppression styles are explicit and auditable:

- inline comments such as `faultline:ignore <playbook-id>`
- playbook-defined path suppressions
- playbook-defined pattern suppressions

Suppressions are reported in structured output and influence scoring rather
than silently disappearing findings.
