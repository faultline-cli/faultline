# Release Boundary

Faultline ships a deliberately narrow default experience for the next release:

- Diagnose a failing CI log with `faultline analyze`
- Turn the winning diagnosis into a deterministic handoff with `faultline workflow`
- Inspect the bundled catalog with `faultline list` and `faultline explain`
- Use `faultline fix` when only the top remediation steps are needed

Everything else should either be a bounded companion surface with explicit verification or a hidden maintainer workflow.

## Scope Classification

### Ship-Ready Core

- `analyze` text, markdown, and JSON output
- `workflow` local and agent output
- `list` and `explain`
- bundled playbook catalog under `playbooks/bundled/`
- checked-in minimal and real fixture corpora
- deterministic release archives and Docker packaging
- release verification via `make release-check`

### Complete Now

- `fix` as a narrow remediation view over the top diagnosis
- `trace` as an advanced deterministic companion for rule-by-rule evaluation and rejection context
- `replay` as a deterministic companion for re-rendering saved analysis artifacts
- `compare` as a deterministic companion for diffing saved analysis artifacts
- `history`, `signatures`, and `verify-determinism` as deterministic forensic-memory companions
- `inspect` and `guard` as advanced local-prevention companions
- `packs install` and `packs list` for optional extra catalog composition
- hidden `fixtures` commands for corpus curation and maintainer workflows

These are supported, but they are not the first-run story. Docs and help text should keep the default emphasis on log diagnosis plus workflow handoff.

### Gate Behind Flag

- provider-backed GitHub Actions and GitLab CI delta via `--delta-provider github-actions|gitlab-ci`
- constrained playbook hooks via hidden `--hooks <mode>`

This path remains available only behind the hidden opt-in `FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1` (preferred; legacy `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1` is still accepted). It is intentionally excluded from the default help surface and release narrative until it has release-grade coverage equivalent to the core CLI flow.

Hooks are also intentionally hidden in the current release. They extend
playbooks through typed, policy-gated local checks, but they are not part of
the default onboarding path and should stay additive to `analyze` and `trace`
rather than becoming a generic automation surface.

### Defer Or Remove From Default Narrative

- broad "CI automation platform" framing
- implying that provider-backed delta is part of the standard product path
- treating repo inspection or pack management as the primary onboarding path

These capabilities may exist, but they should not define the release boundary.

## v0.4 Direction

The current roadmap for v0.4 should extend this boundary rather than replace it:

- keep the default narrative centered on `analyze`, `workflow`, `list`,
  `explain`, and `fix`
- treat managed inheritance as a `packs`-driven capability with explicit sync
  or update flows, not runtime network fetch during analysis
- keep the authoring assistant hidden and maintainer-only until it has
  deterministic validation equivalent to the existing fixture workflows
- add reliability metrics and quarantine recommendations first as additive JSON
  and workflow outputs, not as new first-run command surfaces
- preserve the no-runtime-network expectation for `analyze`, `workflow`, and
  `trace`
- keep history value explicit in output and companion commands rather than
  turning recurrence into hidden ranking behavior

## Release-Readiness Contract

The repository is release-ready only when all of these stay true:

- `make test` passes
- `make review` passes after playbook or pattern changes
- `make fixture-check` passes on the accepted real corpus baseline
- `make cli-smoke` passes against checked-in examples and workflow snapshots
- `make release-check VERSION=<tag>` passes before a release cut

## Bayes Promotion Gate

`--bayes` remains an explicit opt-in flag. Before it can graduate to a default or release-gated path, all of these must hold:

- `make bayes-check` shows zero regressions across the real fixture corpus
- `make bayes-check` shows no Top-1 or Top-3 rate regression vs the baseline scorer
- The comparison report is reviewed and checked in as part of the promotion commit
- The release notes document the promotion explicitly

Run the gate with `--fail-on-regression` to enforce it in CI:

```bash
./bin/faultline fixtures compare-modes --class real --fail-on-regression
```

The current known state is one Bayes regression (`gitlab-gitlab-org-gitlab-runner-6557-s3-64c99cfe7a2f9dfa`, rank 1 → 2). Bayes stays opt-in until that regression is resolved.

## Contribution Rule

New user-facing surfaces should start hidden, flagged, or non-default unless they ship with:

- deterministic command coverage
- fixture-backed regression proof when matching or ranking changes
- checked-in example or snapshot validation when the output is user-facing
- release-check integration when the feature is part of the shipped path
