# Release Boundary

## Locked Product Boundary

Faultline uses an open-core product boundary:

- Core (free): "What failed?"
- Team (paid): "What keeps failing, who owns it, and what do we do about it?"

This boundary is intentional and should remain stable unless explicitly revised.

### Core (free)

Core remains local-first, deterministic, and useful on a single log without
team state.

Core includes:

- CLI analysis surfaces (`analyze`, `batch`, `workflow`, `list`, `explain`, `fix`)
- deterministic playbook matching and ranking
- local playbook system and pack composition
- opt-in single-repo local history and forensic recall
- thin diff and deterministic artifact generation

Core may include persistence when it remains local and scoped to a single
repository. If a feature depends on cross-repo state, organizational
coordination, hosted context, or correlation across multiple repositories, it
should not be presented as part of the free default story.

### Team (paid)

Team is the coordination and memory layer across repeated runs, repositories,
and engineers.

Team includes:

- failure history and aggregation across runs
- recurring failure detection over time
- cross-repo correlation and ownership-aware recurrence
- org-level policy enforcement
- shared playbook layering (org base + repo override)
- managed execution context
- aggregated reporting and trend insights
- stable integration contracts, schema versioning, and sync surfaces

Team is the thing sold and billed, not the individual command user. The CLI may
stay local-first, but the paid value comes from persistent coordination state
tied to a team context.

Rule of thumb:

- If a feature works on one log or within one repository without cross-repo
  correlation, it belongs in Core.
- If a feature aggregates, correlates across repositories, coordinates,
  enforces, or learns across team-wide runs, it belongs in Team.

Faultline ships a deliberately narrow default experience for the next release:

- Diagnose a failing CI log with `faultline analyze`
- Diagnose several local CI logs with `faultline batch`
- Inspect a local source tree with `faultline inspect`
- Turn the winning diagnosis into a deterministic handoff with `faultline workflow`
- Inspect the bundled catalog with `faultline list` and `faultline explain`
- Use `faultline fix` when only the top remediation steps are needed

Everything else should either be a bounded companion surface with explicit verification or a hidden maintainer workflow.
`report` remains a bounded companion surface: useful after stored local runs
exist, but not part of the first-run diagnosis narrative.

Team commands and Team-enriched modes are intentionally outside this default
onboarding narrative.

## Scope Classification

### Ship-Ready Core

- `analyze` text, markdown, and JSON output
- `batch` text, markdown, and JSON output for local multi-log diagnosis
- `workflow` local and agent output
- `inspect` text and JSON output for deterministic source scanning
- `list` and `explain`
- `fix` as a narrow remediation view over the top diagnosis
- bundled playbook catalog under `playbooks/bundled/`
- checked-in minimal and real fixture corpora
- deterministic release archives and Docker packaging
- release verification via `make release-check` or the script-backed
  `make release-verify`

### Complete Now

- `report` as a local-only recurrence summary over stored `analyze` runs
- explicit `--playbook-pack` flags for deterministic local catalog composition
- hidden `fixtures` commands for corpus curation and maintainer workflows

These are supported, but they are not the first-run story. Docs and help text should keep the default emphasis on log diagnosis, source inspection, and workflow handoff.

### Gate Behind Flag

- provider-backed GitHub Actions and GitLab CI delta

This path is removed from the shipped `analyze` surface. It may return later as
a Team enrichment or sync surface, but it should not live inside the local
diagnosis path.

### Team-Gated Commands

These commands or modes require Team auth and backend state:

- `faultline sync` (explicit push of local failure artifacts to Team state)
- `faultline policy apply`
- `faultline analyze <log> --report` (enriched mode)

Expected behavior pattern:

- without Team auth: deterministic local output only
- with Team auth: deterministic local output plus Team enrichment

Auth and upgrade model:

- login remains optional until a Team surface is used
- `faultline auth login` should feel like an upgrade path, not a prerequisite
- stored credentials should make later Team usage invisible after setup
- team context, not user identity alone, is the unit of state and billing

Core flags should not be license-gated when they only operate on local input.

### Defer Or Remove From Default Narrative

- broad "CI automation platform" framing
- implying that provider-backed delta is part of the standard product path
- treating repo inspection or pack management as the primary onboarding path
- reviving hidden command surfaces without deterministic tests, shipped docs,
  and a clear release-boundary entry

These capabilities may exist, but they should not define the release boundary.

## v0.4 Direction

The current roadmap for v0.4 should extend this boundary rather than replace it:

- keep the default narrative centered on `analyze`, `batch`, `inspect`,
  `workflow`, `list`, `explain`, and `fix`
- treat managed inheritance as explicit local pack composition or a future Team
  sync flow, not runtime network fetch during analysis
- keep the authoring assistant hidden and maintainer-only until it has
  deterministic validation equivalent to the existing fixture workflows
- add reliability metrics and quarantine recommendations first as additive JSON
  and workflow outputs, not as new first-run command surfaces
- preserve the no-runtime-network expectation for `analyze` and `workflow`
- keep history value explicit in output and companion commands
  rather than turning recurrence into hidden ranking behavior

## Team Layer Delivery Contract

When Team features are implemented, preserve these constraints:

- CLI remains local-first and usable without login
- auth is upgrade-driven and only required for Team surfaces
- Team state is tied to team context, not individual user state
- paid value comes from persistent coordination and aggregation, not from
  crippling local diagnosis
- default analysis path keeps deterministic no-runtime-network behavior unless
  explicit Team sync/report mode is requested
- initial pricing and packaging should optimize for per-team value before
  introducing seat-based complexity

## Release-Readiness Contract

The repository is release-ready only when all of these stay true:

- `make test` passes
- `make review` passes after playbook or pattern changes
- `make fixture-check` passes on the accepted real corpus baseline
- `make bayes-check` passes because Bayes is default ranking behavior
- `make docs-check` passes so generated failure docs stay aligned with bundled playbooks
- `make cli-smoke` passes against checked-in examples and workflow snapshots
- `make release-check VERSION=<tag>` passes before a release cut

## Bayes Promotion

`--bayes` was promoted to the default in v0.4.5. The Bayesian-inspired reranking layer is now active by default on the shipped analysis commands (`analyze`, `workflow`). Pass `--bayes=false` to disable it and revert to the deterministic baseline scorer.

Promotion criteria satisfied before v0.4.5:

- `make bayes-check` showed zero regressions across the real fixture corpus (delta 0.000)
- `make bayes-check` showed no Top-1 or Top-3 rate regression vs the baseline scorer
- The comparison report was reviewed as part of the promotion commit
- The release notes document the promotion explicitly

The gate can still be run with `--fail-on-regression` to verify the corpus state:

```bash
./bin/faultline fixtures compare-modes --class real --fail-on-regression
```

Zero Bayes regressions across 215 accepted real fixtures. `make bayes-check` exits zero with delta 0.000, top-1 1.000, top-3 1.000 in both modes.

## Contribution Rule

New user-facing surfaces should start hidden, flagged, or non-default unless they ship with:

- deterministic command coverage
- fixture-backed regression proof when matching or ranking changes
- behavioral fixture coverage that distinguishes positive matches, near-miss or disallowed-playbook assertions, and strict top-1 expectations rather than relying on raw coverage percentages or fixture filename counts
- checked-in example or snapshot validation when the output is user-facing
- release-check integration when the feature is part of the shipped path
