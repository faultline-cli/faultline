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

### Technical Debt Backlog

The following items were identified in a full technical debt audit of the
codebase at the v0.4.3 planning milestone. Each item includes the affected
location, observed risk, and the recommended fix. Priority ordering and quick
wins are listed at the end of this section.

#### TD-1: `AnalyzeOptions` mega-struct

`internal/app/commands.go` defines `AnalyzeOptions` with 30+ fields spanning
at least eight unrelated concerns: provider selection, store configuration,
output format, trace options, replay flags, Delta settings, scoring, and
deprecated history fields. Every command receives the full struct even when
it uses two or three fields. Callers in `internal/cli/root.go` set the struct
in 20+ locations, making future changes high-blast-radius.

Fix: introduce focused sub-structs — `ProviderOptions`, `DeltaOptions`,
`TraceOptions`, `OutputOptions` — and have `AnalyzeOptions` compose them.
Commands that need only a subset get only their sub-struct. This shrinks
call-site diffs from file-wide to function-local.

Priority: High. Every future feature addition currently touches this struct.

#### ~~TD-2: `cli/root.go` monolith~~ ✅ DONE

`internal/cli/root.go` is 1 161 lines containing all 20 command factory
functions (`newAnalyzeCommand` through `newVerifyDeterminismCommand`).
Every new command and every flag addition increases merge risk for the whole
file. There are no unit tests for individual command factories; breakage
surfaces only at integration time.

Fix: move each command factory to its own file
(`internal/cli/cmd_analyze.go`, `internal/cli/cmd_workflow.go`, etc.).
`root.go` becomes a thin registration file. Command-level unit tests become
straightforward.

**Resolved**: Split into 12 per-command files; `root.go` is now 54 lines.
All tests pass.

#### ~~TD-3: Dual workflow system (`legacy.go` co-exists with `workflow.go`)~~ ✅ DONE

`internal/workflow/legacy.go` and `workflow/workflow.go` are two diverging
implementations of the same concept. `legacy.BuildWithOptions` still drives
the default output for `faultline workflow`. The richer post-v0.4 system is
used only by subcommands. Schema fields and serialization rules are diverging
silently; no marker distinguishes a legacy workflow record from a current one
at query time.

Fix: write an ADR capturing the migration intent; add a `[LEGACY]` label
comment at the top of `legacy.go`; block any new feature work from landing in
the legacy path. Gate removal on the eval corpus confirming zero regression.

**Resolved**: ADR 0010 written (`docs/adr/0010-legacy-workflow-migration-path.md`);
`[LEGACY]` header block added to `legacy.go` with removal criteria. The legacy
path remains until `make eval` confirms zero regression.

#### ~~TD-4: `renderer.go` monolith~~ ✅ DONE

`internal/renderer/renderer.go` was 1 276 lines with 60+ methods. Split into:
`renderer_analysis.go`, `renderer_coverage.go`, `renderer_fix.go`,
`renderer_common.go`, and `renderer_workflow.go` (stub). `renderer.go` now
contains only the package-level `leadingHeadingPattern` var, the `Renderer`
type, and `New`. No logic changes; pure file decomposition. All tests pass.

#### ~~TD-5: SQLite double-UPDATE for workflow execution ID~~ ✅ DONE

`internal/store/sqlite.go` — `RecordWorkflowExecution` executes:
INSERT → `LastInsertId` → UPDATE `execution_id` → re-marshal → UPDATE
`record_json`. A crash between the INSERT and the second UPDATE leaves a
permanently incomplete row. The pattern exists because the execution ID is not
known before the INSERT.

Fix: generate the execution ID on the client side (UUID or deterministic hash)
before the INSERT so both columns are set in a single statement. Eliminates
the partial-row failure window.

**Resolved**: `newExecutionID()` helper generates a random `wf-<16 hex>` ID
via `crypto/rand` before the INSERT. Single INSERT now sets both
`execution_id` and `record_json`; both post-INSERT UPDATEs removed.

#### TD-6: `internal/app` coverage gap (65%)

`internal/app` contains the main analysis entry points — `writeAnalysis`,
`analyzeLog`, and all store interactions — but is covered at only 65%, the
lowest non-trivial package coverage in the codebase. Edge cases in store
writes, history branching, and structured output are exercised only indirectly
through end-to-end tests.

Fix: add golden-output unit tests for `writeAnalysis` and `analyzeLog` using
captured fixtures; add table-driven tests for the store interaction branches
in `app/service.go`. Target 80%+ line coverage.

Priority: High. This is the most critical coverage gap by impact surface.

#### ~~TD-7: Deprecated `NoHistory` and `MetricsHistoryFile` still active~~ ✅ DONE

`NoHistory` removed from `engine.Options` and `app.AnalyzeOptions`.
`MetricsHistoryFile` removed from `engine.Options` (retained in
`app.AnalyzeOptions` where it is still consumed by `store_support.go`). All
11 call sites updated to use `Store: "off"`. No behaviour change. All tests
pass.

#### TD-8: `AnalyzeOptions` passed to commands that use two or three fields

Several commands receive the full `AnalyzeOptions` struct but access only
`OutputFormat`, `Verbose`, and `StoreDir` (or similar). This is a symptom of
TD-1 and is resolved by introducing the sub-struct decomposition described
there.

Priority: Low (addressed by TD-1).

---

**Priority order for scheduling:**

1. TD-6 — `internal/app` coverage gap (highest impact, directly improves
   confidence in the analysis path)
2. TD-1 — `AnalyzeOptions` decomposition (architectural force-multiplier;
   every future feature benefits)
3. TD-2 — `cli/root.go` split (contributor DX; one-time clean-up PR)
4. TD-3 — Dual workflow ADR + legacy marker (prevents silent schema drift)
5. TD-5 — SQLite execution-ID fix (correctness, low probability but
   non-recoverable failure mode)
6. TD-7 — Remove deprecated `NoHistory`/`MetricsHistoryFile` (clean-up sprint)
7. TD-4 — `renderer.go` split (safe file decomposition, low urgency)
8. TD-8 — addressed by TD-1

**Quick wins (≤ 1 day each):** TD-7 (mechanical field removal), TD-4
(file-split only, no logic changes), TD-3 (ADR writeup + comment marker).

**Strategic assessment:** The single highest-leverage improvement for
long-term velocity is TD-1 (`AnalyzeOptions` decomposition). It is the root
cause of high blast radius on nearly every feature PR and the structural
reason TD-8 exists. TD-1 should land before the v0.5 Team-layer work begins.

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

## v0.4.4 Release

### Theme

Distribution surface and automation ergonomics.

The goal of v0.4.4 is to make faultline usable without manual wiring in CI:
a stable exit code contract that scripts can rely on, a batch command that
surfaces recurring root causes across a build matrix, and a thin GitHub
Action that brings the tool into the ecosystem without requiring shell glue.

### Delivery Order

#### 1. Stable exit code contract

Faultline now defines a three-tier exit code contract:

| Code | Meaning |
|------|---------|
| `0`  | Success — analysis completed; guard: no findings; batch: all logs matched |
| `1`  | Operational finding — guard: findings emitted; batch: one or more logs unmatched; analyze: silent failure with `--fail-on-silent` |
| `2`  | Error — bad arguments, unreadable input, or processing failure |

**Breaking change from v0.4.3:** errors previously exited with code 1.
Scripts that currently check `$? -eq 1` to detect any failure must be updated
to distinguish code 1 (expected result) from code 2 (unexpected error).

**Sentinels:** `app.ErrGuardFindings`, `app.ErrSilentFailure`, and
`app.ErrBatchUnmatched` map to exit code 1. All other errors map to exit
code 2.

#### 2. `faultline batch`

New command: `faultline batch <file> [file ...]`

Analyzes multiple CI log files sequentially and groups matched diagnoses by
failure pattern. Core value: "3 of your 12 nightly builds share the same root
cause — `missing-executable`."

**Output modes:**

- Terminal (default): human-readable deduplication summary with pattern
  counts and affected files
- JSON (`--json`): `batch.v1` schema with `patterns`, `entries`,
  `unmatched_sources`, total/matched/unmatched counts

**Exit semantics:**

- `0` — all sources matched a playbook
- `1` — one or more sources unmatched (`ErrBatchUnmatched` sentinel)
- `2` — file open failure or analysis error

**Flags:** `--json`, `--format terminal|json`, `--playbooks`,
`--playbook-pack`, `--no-history`

**Not in scope for v0.4.4:** parallel execution, stdin aggregation,
per-file verbose output, markdown format. These can follow in v0.4.5.

#### 3. Official GitHub Action (`faultline-action`)

Thin wrapper around stable CLI artifacts. Documented target for v0.4.4
planning; implementation tracked in a separate repository.

**Design constraints:**

- The Action is a thin shell wrapper: no JS/TS runtime, no Docker image
  build in the critical path
- It downloads the versioned CLI binary from GitHub Releases (pinned SHA)
- It calls `faultline analyze` or `faultline batch` and maps exit codes to
  Action outcomes
- It does not add opinions beyond what the CLI already expresses
- The stable exit code contract (above) is a prerequisite

**Not in scope for v0.4.4:** SARIF output, GitHub annotations as the primary
output surface, auto-PR comments. The annotation path is already available
via `--ci-annotations` on the analyze command; the Action just passes the
flag through.

#### 4. Corpus coverage hardening — noisy log testing

The most impactful reliability work in v0.4.4 is proving that playbooks
hold up against real, dirty CI logs — not just clean, curated fixtures.

**Problem:** The fixture suite covers synthetic and well-structured logs
well. Production CI logs are noisier in ways that break pattern matching:
timestamps prepended to every line, interleaved output from parallel jobs,
ANSI/VT100 escape codes, progress-bar rewrites, platform banners, and
multi-kilobyte preambles before the first meaningful line. Several playbooks
currently over-fit to clean log structure and produce false negatives or
weak matches when that structure is absent.

**Primary goal — noisy log test class:**

Establish a dedicated `noisy` fixture class in `fixtures stats` that must
pass before any playbook ships. A noisy fixture is a real or synthetically
degraded log that deliberately exercises at least one noise type:

| Noise type | Description |
|------------|-------------|
| `timestamps` | ISO-8601 or Unix epoch prefix on every line |
| `ansi` | VT100 color/cursor escape sequences |
| `parallel-interleave` | Output from ≥ 2 concurrent jobs mixed in the same stream |
| `platform-banner` | GitHub Actions / GitLab CI / CircleCI job header blocks |
| `truncation` | Log cut mid-line or mid-block due to size limits |
| `progress-rewrite` | `\r`-based progress lines overwriting the same terminal row |

Each fixture must declare its noise type(s) so failures during `fixtures stats`
point directly to the pattern that broke under load.

**Secondary goal — lift `real` accuracy:**

- Collect real CI logs (GitHub Actions, GitLab CI, CircleCI) that exercised
  known failure modes and verify each matches the correct playbook at
  confidence ≥ 0.70
- For every playbook that falls below that bar, tighten or broaden evidence
  patterns, update `context_window`, or add a `noisy_log` fixture that must
  pass before the playbook ships
- Priority failure categories: `auth`, `dependency`, `network`, and
  `environment` — highest false-negative rate under noise

**Acceptance criteria:**

- `faultline fixtures stats --class noisy` top-1 accuracy ≥ 0.70 (new baseline, must exist before release)
- `faultline fixtures stats --class real` top-1 accuracy ≥ 0.80 (up from ~0.72)
- Zero regressions against existing `synthetic` and `real` baselines
- Every noisy fixture documents: origin CI platform, noise type(s), and the
  playbook it is intended to match

### Not v0.4.4

- SARIF output (planned v0.5)
- `faultline report` or any Team-layer commands
- parallel batch execution
- hosted or cloud-synced pack registry
