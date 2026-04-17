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
- `inspect` and `guard` as advanced local-prevention companions
- `packs install` and `packs list` for optional extra catalog composition
- hidden `fixtures` commands for corpus curation and maintainer workflows

These are supported, but they are not the first-run story. Docs and help text should keep the default emphasis on log diagnosis plus workflow handoff.

### Gate Behind Flag

- provider-backed GitHub Actions delta via `--delta-provider github-actions`

This path remains available only behind the hidden opt-in `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1`. It is intentionally excluded from the default help surface and release narrative until it has release-grade coverage equivalent to the core CLI flow.

### Defer Or Remove From Default Narrative

- broad "CI automation platform" framing
- implying that provider-backed delta is part of the standard product path
- treating repo inspection or pack management as the primary onboarding path

These capabilities may exist, but they should not define the release boundary.

## Release-Readiness Contract

The repository is release-ready only when all of these stay true:

- `make test` passes
- `make review` passes after playbook or pattern changes
- `make fixture-check` passes on the accepted real corpus baseline
- `make cli-smoke` passes against checked-in examples and workflow snapshots
- `make release-check VERSION=<tag>` passes before a release cut

## Contribution Rule

New user-facing surfaces should start hidden, flagged, or non-default unless they ship with:

- deterministic command coverage
- fixture-backed regression proof when matching or ranking changes
- checked-in example or snapshot validation when the output is user-facing
- release-check integration when the feature is part of the shipped path
