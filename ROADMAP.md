# Faultline Roadmap

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

## Later, Not v0.4

The roadmap should stay disciplined about what it is not doing in this release:

- hosted pack registry
- runtime remote pack fetch during analysis
- dashboards or a hosted analytics surface
- CI or test execution orchestration inside Faultline
- AI-generated fixes in the product's authoritative decision path
- speculative governance layers such as signing or enterprise policy control
  before the pack and provenance model is stable
