# Faultline Roadmap

## Locked Commercial Boundary

Product packaging is locked to this model:

- Core (free): deterministic local diagnosis of "what failed"
- Team (paid): cross-run coordination for "what keeps failing, who owns it,
  and what to do about it"

Faultline should mirror the Git/GitHub split:

- CLI product = local deterministic substrate
- Team layer = persistence, aggregation, and organizational coordination

This means monetization should track persistent team value, not local parsing
or one-off diagnosis.

## Current Position

Faultline already ships the deterministic CLI foundations that the roadmap
should build on:

- stable `analyze`, `workflow`, `list`, `explain`, and `fix` flows
- deterministic playbook loading, matching, ranking, and rendering
- bundled-plus-extra pack composition through `internal/playbooks`
- checked-in fixture corpus, sanitizer flow, and regression gates
- stable JSON and workflow artifacts for automation and agent handoff

The next release should not restart the story from "basic log analyzer."
It should extend the shipped CLI into a fleet-safe deterministic operations
layer without weakening the local-first trust boundary.

## Team Plan (v1)

The first paid layer should ship as a lean, text-first extension of current
CLI workflows.

### Team Capabilities

1. Failure history and aggregation
2. Recurring failure detection
3. Org-level policy enforcement
4. Shared playbook layering (org + repo)
5. Basic failure insights (`faultline report`)
6. Team-level hooks automation
7. Versioned integration and schema contract

### MVP Build Order

Phase 1 (sellable baseline):

- local history store (SQLite)
- `faultline report` basic aggregation output
- policy enforcement evaluator
- playbook inheritance merge rules

Phase 2:

- hooks execution for Team automation paths
- recurring detection thresholds
- simple sync path (push-only first)

Phase 3:

- hosted aggregation backend
- optional dashboards and richer hosted surfaces

### Team Auth and Gating Model

- local diagnosis commands remain available without auth
- Team commands prompt upgrade flow when unauthenticated
- login should be optional and only introduced on Team command use
- token storage and refresh should be local and transparent after setup

### Team Unit of Sale

Sell to teams, not users:

- user, team, membership, and optional project entities
- per-team pricing first; seat complexity can follow later

## v0.4 Theme

**Fleet-safe deterministic operations**

v0.4 should make five things clear:

- the deterministic forensic engine remains the substrate for every new feature
- managed inheritance is the main enterprise headline
- authoring stays grounded in the existing fixture and review loop
- reliability metrics are explicit machine-readable outputs, not hidden scoring
- quarantine remains advisory policy, not CI orchestration

## Why This Order

The proposed v0.4 ordering is:

1. Deterministic Forensic Engine (Core)
2. Managed Inheritance (Golden Playbook Registry)
3. Authoring Assistant (Knowledge Codification)
4. Pipeline Reliability Metrics (PHI, TSS, FPC)
5. Deterministic Quarantine Policy

This v0.4 engineering order remains valid and should compose with the Team
plan above. Team work should be added as a coordination layer over the same
deterministic substrate, not as a replacement architecture.

This order matches the repository shape:

- the engine, output, workflow, fixture, and pack seams already exist
- managed inheritance extends the current pack boundary cleanly
- authoring should target the inheritance model rather than predate it
- metrics need explicit artifacts, pack provenance, and stable history inputs
- quarantine should consume metrics rather than invent a parallel subsystem

## v0.4 Delivery Order

### 1. Deterministic Forensic Engine (Core)

Treat the existing engine as the v0.4 substrate, not a solved problem that can
be ignored while new surfaces pile on.

v0.4 work:

- harden `analyze`, `workflow`, `fix`, `trace`, differential diagnosis, and
  stable JSON as the foundation every later feature depends on
- preserve deterministic tie-breaks, evidence provenance, and stable output
  schemas as product promises, not just implementation details
- keep `workflow` derived from analysis results, repo context, and checked-in
  playbook metadata only

Why first:

- every later feature needs a stable diagnosis object model
- the release boundary already treats these commands as the core story
- additive roadmap work is safer than parallel architecture creation

### 2. Managed Inheritance (Golden Playbook Registry)

This is the first true v0.4 feature pillar and the clearest enterprise-scale
capability extension.

v0.4 work:

- extend the existing pack model with **local sync / pinned reference**
  inheritance instead of runtime remote fetch during `analyze`
- keep any network activity inside explicit `packs` management flows
- add pack provenance so results can report the synced pack and version lineage
- support constrained insert/overlay bindings for service-local extension
  without free-form rule mutation

Why second:

- it builds directly on the current bundled-plus-extra pack architecture
- it solves the highest-value fleet maintenance problem without changing the
  analysis trust boundary
- it creates the right destination for later codified knowledge

### 3. Authoring Assistant (Knowledge Codification)

The authoring assistant should land after inheritance so new knowledge can be
captured into the right pack and overlay model.

v0.4 work:

- keep the assistant maintainer-only and hidden from the default CLI narrative
- reuse the existing deterministic fixture pipeline as the source of truth
- support sanitized-log intake, candidate playbook or overlay scaffolding, and
  fix/validation draft generation
- keep any LLM augmentation optional, non-authoritative, and outside core
  product logic

Why third:

- authoring without a target inheritance model creates rework
- the repository already has deterministic review gates that can police quality
- this stays aligned with the existing local skills and prompt workflows

### 4. Pipeline Reliability Metrics (PHI, TSS, FPC)

Reliability metrics should arrive as additive machine-readable outputs once pack
provenance and explicit artifact inputs are in place.

v0.4 work:

- add an additive `metrics` block to analysis and workflow JSON
- compute metrics from explicit artifact sets or supplied history only
- make TSS the first-class metric because it has the clearest deterministic path
- expose PHI and FPC only when sufficient input data exists
- surface drift-component reporting so external automation can identify what is
  degrading reliability

Why fourth:

- these metrics are most useful once pack provenance and authoring loops exist
- JSON and workflow artifacts already provide the right distribution boundary
- dashboards can stay external; Faultline only needs to emit stable data

### 5. Deterministic Quarantine Policy

Quarantine belongs last because it should be the policy layer built on top of
the metrics layer rather than a separate execution engine.

v0.4 work:

- emit advisory policy recommendations such as `blocking`, `observe`, or
  `quarantine`
- base policy on documented TSS and FPC thresholds
- keep retries, suite isolation, and CI routing outside Faultline itself
- expose the same policy through additive JSON and workflow hints

Why fifth:

- quarantine quality depends on the reliability metrics being explicit first
- keeping it advisory preserves Faultline's role as a diagnosis and policy CLI
- this avoids quietly turning the product into a flaky-test orchestrator

## Interface Direction

Planned additive interface changes for v0.4:

- `packs` grows synced-reference metadata and pinned update flows for managed
  inheritance
- analysis JSON grows additive sections for `pack_provenance`, `metrics`, and
  `policy`
- workflow JSON grows additive metrics and policy hints derived from the same
  deterministic analysis result
- authoring assistance remains hidden and should compose with the existing
  `fixtures` and playbook-authoring workflows rather than redefine the command
  maturity model

Planned additive Team-facing command surfaces:

- `faultline report` for aggregated text-first insights
- `faultline sync` for explicit Team state push/sync
- `faultline policy apply` for org-level deterministic policy evaluation
- `faultline analyze <log> --report` for mixed local + Team enrichment mode

Defaults to preserve:

- absent data means absent fields, not guessed values
- the same local playbook set plus the same input still yields the same output
- `analyze`, `workflow`, and `trace` must not require runtime network access

## v0.4 Release Boundary Rules

The current release boundary remains the guardrail for v0.4 planning:

- the default narrative stays centered on `analyze`, `workflow`, `list`,
  `explain`, and `fix`
- managed inheritance should land under `packs`, not as a default-networked
  analysis path
- the authoring assistant should stay hidden and maintainer-only until it has
  deterministic validation equivalent to other maintainer workflows
- metrics and quarantine should start as machine-readable companion outputs
  rather than new first-run commands
- any future promotion to the default narrative should require deterministic
  coverage, checked-in snapshots where relevant, and release-check integration

## Validation Standard

Core hardening and any v0.4 implementation work should satisfy these checks:

- snapshot-test JSON, workflow, and trace stability
- verify pack provenance is deterministic across repeated runs
- verify synced packs resolve offline after sync and preserve stable ordering
- require authoring output to pass `make review`, `make test`, and
  `make fixture-check` before promotion
- snapshot-test TSS, PHI, and FPC calculations, missing-data behavior, and
  rounding
- verify quarantine recommendations never trigger retries or CI mutations inside
  Faultline itself

## v0.4.3 Release

### Theme

**Corpus accuracy, command promotion, and operator ergonomics**

v0.4.3 ships the highest-leverage improvements that are unblocked today:
narrow the corpus gap with targeted playbooks, promote tested commands out of
hidden status, and expose the operator flags most requested for `fix`.

### Delivery Order

#### 1. Top-5 corpus gap playbook sprint

Address the five largest unmatched clusters from the current eval corpus
(2f3942e9, 249d8ae2, 8b75bb23, 9851e304, df423c00). Each cluster represents
10–11 unmatched fixtures. Closing all five raises large-corpus coverage by
approximately 0.11–0.12 percentage points and validates the playbook sprint
workflow as a repeatable quality lever.

Why first: highest coverage ROI, fully unblocked, and each new playbook ships
with a checked-in fixture so regressions are immediately visible.

#### 2. Jest test failure playbook

A dedicated `jest-test-failure` playbook for the Jest test runner failure
pattern accounts for roughly 0.9% of unmatched corpus cases. Ships with a
checked-in fixture and follows the standard review gate.

Why second: small, self-contained, and immediately measurable against the
eval corpus.

#### 3. Promote `faultline coverage` to stable

Remove `Hidden: true` from `newCoverageCommand()`. Add unit tests and a
smoke snapshot. Coverage becomes a documented command in `faultline --help`
and the ship-ready command list.

Why third: two-day task, zero risk, immediate utility for playbook authors
and CI pipelines verifying catalog completeness.

#### 4. Ontology Phase 2–4 (tag existing playbooks)

Tag all 181 bundled playbooks with ontology labels from the Phase 1 design.
Extend the coverage command to report by ontology tag. This completes the
Phase 1 design intent and makes the catalog machine-navigable.

Why fourth: safe additive change to YAML metadata with no analysis path
impact; best done after coverage is promoted so the tag report is immediately
visible.

#### 5. `faultline fix` operator flags

Add `--commands-only`, `--with-preconditions`, and `--with-risks` flags to
`faultline fix`. No new fix logic; these flags filter the existing structured
fix output rendered by each playbook.

Why fifth: operator ergonomics with zero analysis risk; depends only on the
existing structured fix model.

#### 6. Promote `faultline compare` to stable

Remove the experimental flag from `faultline compare`. Add smoke snapshot.
Promote to the default command narrative.

Why sixth: compare is fully implemented and useful; the only blocker is a
missing smoke snapshot and the experimental gate.

### Scope Summary

| Item | Effort | Coverage delta | Risk |
|------|--------|---------------|------|
| Top-5 corpus gap sprint | 8–10 days | +0.11–0.12% (large corpus) | Low |
| Jest playbook | 2–3 days | +0.9% (large corpus) | Low |
| Promote coverage | 2–3 days | — | Very low |
| Ontology Phase 2–4 | 14–17 days | — | Low (metadata only) |
| fix operator flags | 3–5 days | — | Very low |
| Promote compare | 2–3 days | — | Very low |

### Success Criteria

- `make test` and `make cli-smoke` pass on all changes
- eval corpus large-corpus coverage moves from 74.0% toward 74.1%+
- `faultline coverage` and `faultline compare` appear in `faultline --help`
- all new playbooks pass `make review`
- no regressions in existing `analyze`, `workflow`, `fix`, or `explain` flows

### Team Layer Deferral

The Team layer (`faultline report`, `faultline login`, `faultline sync`) is
deferred to v0.5. The local forensic store is shipped and the patterns in
`internal/app/history.go` provide the implementation template. v0.4.3 should
not touch the Team command surface.

## Later, Not v0.4.3

The roadmap should stay disciplined about what it is not doing in this release:

- Team layer commands (`faultline report`, `faultline login`, `faultline sync`)
- hosted pack registry
- runtime remote pack fetch during analysis
- dashboards or a hosted analytics surface
- CI or test execution orchestration inside Faultline
- AI-generated fixes in the product's authoritative decision path
- speculative governance layers such as signing or enterprise policy control
  before the pack and provenance model is stable
