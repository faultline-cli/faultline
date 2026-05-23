# Architecture

Faultline keeps the shipped CLI surface stable, but the runtime is split into
explicit deterministic layers:

- `internal/cli` owns Cobra command definitions, flags, stdin/file handling,
  the internal command-surface manifest, and handing structured options into
  the app layer.
- `internal/app` owns command use-cases such as analyze, fix, list, explain,
  workflow, inspect, report, batch, and maintainer fixture workflows.
  Command methods should stay thin: shared log-analysis orchestration lives in
  `internal/app/analysis_pipeline.go`, while source-analysis companion flows
  live in `internal/app/source_analysis.go`.
- `internal/artifact` owns construction of the first-class `FailureArtifact`
  used for storage, replay, and remediation handoff.
- `internal/store` owns optional durable local forensic memory, deterministic
  signature hashing, SQLite persistence, and explicit schema migrations.
- `internal/engine` owns analysis orchestration and depends on explicit
  collaborators for playbook catalogs, detector lookup, source loading, and
  git enrichment. It does not own persistence.
- `internal/workflow` owns deterministic workflow handoff generation for the
  top-level `faultline workflow` command.
- `internal/fixtures` owns deterministic fixture corpora, public-source
  ingestion adapters, normalization, review metadata, promotion flow, and
  regression statistics.
- `internal/coverage` owns coverage reports over the resolved playbook catalog.
  It derives fixture evidence from corpus expectations, including positive
  matches, near-miss or disallowed-playbook assertions, and strict top-1
  requirements. The CLI should not infer coverage from fixture filenames alone.
- `internal/detectors` owns the detector registry plus the distinct `log` and
  `source` detector implementations.
- `internal/playbooks` owns catalog resolution, YAML loading, validation, and
  deterministic review helpers. Bundled overlap review is gated by a checked-in
  baseline so intentional shared patterns are explicit.
- `internal/scoring` owns the optional Bayesian-inspired evidence-fusion layer
  used for additive reranking explanations and delta diagnosis.
- `internal/output` owns command-facing output selection plus JSON/workflow
  serialization, focused views (`--view summary|evidence|fix|raw`), and
  evidence-only views.
- `internal/renderer` owns terminal-aware human rendering, including quick
  (default) and detailed modes, plain fallback, markdown rendering, and
  restrained ANSI styling.
- `internal/trace` owns rule-by-rule trace payloads used by internal
  diagnostics and tests.

## Playbook boundary

The playbook catalog resolves directories in this order:

1. `FAULTLINE_PLAYBOOK_DIR`
2. `playbooks/bundled`
3. legacy `playbooks`
4. `/playbooks/bundled`
5. `/playbooks`

That keeps today’s repo layout working while giving the repository a clean
default and letting extra packs live in separate directories loaded through an
external path.

Additional packs can be composed on top of the bundled catalog through
the `FAULTLINE_PLAYBOOK_PACKS` environment variable or repeatable
`--playbook-pack` flags. A full `--playbooks` override still resolves a single
custom catalog root and does not combine with extra packs.

Packs can carry deterministic provenance in analysis JSON:

- pack name
- version
- source URL or local source path
- pinned ref when available
- playbook count contributed by that pack

This provenance is additive. It does not change matching; it makes the loaded
catalog auditable in analysis JSON.

Bundled catalog composition should stay generous for adoption: broad coverage for
common CI failures across popular ecosystems. Extra packs can concentrate on
provider-specific depth, advanced deployment or operations workflows, and deeper
source or security rules.

Playbooks can now also form a deterministic inheritance graph across the
composed pack set through `extends: <playbook-id>`. The boundary stays narrow:

- inheritance is local to the already-loaded pack graph
- duplicate playbook IDs are still rejected
- child playbooks inherit and extend structured match, workflow, scoring, and
  explanatory content from the parent
- pack order still matters because later packs may contribute inheriting rules
  that target earlier bundled or team playbooks

The same pack boundary can also carry a deterministic match-fragment graph
through `faultline-matchers.yaml`. Those overlays:

- define reusable named match fragments for `match.any`, `match.all`,
  `match.none`, and `match.partial`
- let playbooks compose those fragments through `match.use`
- support explicit `playbook:<id>` references when a playbook should reuse the
  fully composed match graph of another rule
- resolve after inheritance, reject composition cycles, and expand back into a
  flat deterministic playbook match set before ranking

This keeps reusable sub-patterns first-class without turning the matcher into a
runtime DSL or a second hidden decision engine.

This is the intended scaling path for playbooks:

- reuse signal fragments for shared evidence instead of repeating literal
  patterns
- inherit a parent playbook when the diagnosis stays the same and only the
  surrounding constraints or remediation need to specialize
- use partial groups for signal clusters that are individually weak but
  collectively decisive

That keeps the catalog composable without losing the deterministic one-rule,
one-root-cause discipline.

Pack composition is explicit per run. Mounted pack directories work in local
and containerized runs by passing the same `--playbook-pack` path or
`FAULTLINE_PLAYBOOK_PACKS` value.

## Fixture ingestion boundary

Fixture ingestion is split deliberately:

- site adapters stay explicit because each source has different URL parsing,
  endpoint layout, and response schema
- HTTP transport and JSON fetch behavior should be shared when possible

That means Faultline should avoid a single generic "web adapter" abstraction
for GitHub, GitLab, Stack Exchange, Reddit, and Discourse. The reusable layer
is the transport, not the source-specific extraction logic.

## Detector boundary

Detectors stay explicit and separate:

- `log` consumes normalized log lines and lightweight log context
- `source` consumes a repository snapshot plus optional changed-file metadata

Both emit the shared `model.Result` shape so ranking, output, workflow, and
history remain stable across command surfaces.

## Store boundary

The store is intentionally narrow:

- local only
- SQLite-backed by default
- opt-in at runtime for analysis-style commands
- additive to the existing analysis path

The ownership split is deliberate:

- `internal/app` resolves store config, opens the store when explicitly enabled,
  handles graceful degradation, enriches results with history, and records
  completed runs
- `internal/store` hides SQL, migrations, and schema details behind a small
  interface plus a no-op fallback
- `internal/engine` stays deterministic and store-agnostic; it returns analysis
  results without querying or mutating on-disk state
- detectors remain stateless in v1 and do not read from the store directly

When enabled, the store records durable forensic memory such as:

- top-diagnosis recurrence by `signature_hash`
- run-level `input_hash` and `output_hash`
- first-class `artifact_json` snapshots for deterministic failure artifacts
- ranked playbook matches for longitudinal review

The store does not become a generic raw-log warehouse. When active, it stores
hashes, normalized signature material, minimal evidence excerpts, and small
structured summaries only. When history is not explicitly enabled, the no-op
store keeps default output stable.

## Scoring boundary

Faultline now has an explicit three-layer ranking model:

1. detectors decide which playbooks matched
2. `internal/scoring` may rerank those already-matched candidates when
   Bayesian reranking is active, and it only emits delta hints when repo-aware context
   is explicit
3. output and workflow consume the final deterministic ordering

That boundary matters:

- detectors remain authoritative
- scoring is assistive, not a second matcher
- same input and same repo snapshot still produce the same output
- ranking and delta payloads are additive and explainable
- changed files are suspicious context, not proof on their own
- provider-backed delta is outside the shipped local diagnosis path

## Rendering boundary

Human-facing longform content is stored in markdown-capable playbook fields such
as `summary`, `diagnosis`, `fix`, and
`validation`.

- markdown is presentation content only
- structured playbook fields still drive matching and ranking
- CLI commands render the same deterministic content model to terminal or markdown output
- `--format json` and `--json` emit the structured machine-readable form
- `--view summary|evidence|fix|raw` selects a focused slice of the human-readable output
  without changing the underlying analysis; `summary` and `raw` map to quick and
  detailed rendering modes respectively; `evidence` and `fix` emit narrow single-purpose
  slices of the top result.
- Rule-by-rule evaluation and artifact re-rendering remain internal
  diagnostics, not separate shipped command surfaces.
- non-TTY and no-color environments fall back to plain output

The stable analysis JSON schema is additive. Beyond the ranked results, it may
also include:

- `pack_provenance` when one or more packs contributed playbooks
- `metrics` when sufficient explicit history exists to compute TSS, FPC, or PHI
- `policy` when a deterministic advisory recommendation can be derived from
  those metrics
- `input_hash` and `output_hash` when local history is enabled for
  repeated-run determinism checks
- result-level `signature_hash` and recurrence fields
  when local history is enabled

Saved analysis artifacts preserve those fields for deterministic handoff and
internal regression checks.

Analysis JSON now also carries an additive first-class `artifact` object that
collapses the winning diagnosis or structured unknown state into one stable
unit of computation. The artifact records:

- fingerprint
- matched playbook identity when present
- evidence and confidence
- environment and history context
- fix steps
- structured unknown clusters and playbook seed hints when unmatched
- remediation commands, patch suggestions, and CI config diff hints

Workflow artifacts are derived from the same deterministic analysis object.
When present, JSON and text workflow output may also carry:

- `ranking_hints`
- `delta_hints`
- `metrics_hints`
- `policy_hints`
- the embedded failure `artifact`
- a structured `remediation` block with commands, patch suggestions, and CI config diff hints

Absent data remains absent. Faultline does not invent placeholder values for
missing history or policy inputs.

## Architectural gates (pre-Team Phase 1)

Two structural decisions must be made before Team Phase 1 features land. They
are recorded here so they remain visible and do not drift into undecided state.

### `internal/app` sub-package decomposition

`internal/app` owns the shipped core use-cases plus maintainer fixture workflows.
As Team features arrive,
new use-cases must not land in `internal/app` without first establishing a
sub-package boundary. The pattern to follow:

- Extract one or two use-cases into `internal/app/<name>` sub-packages.
- Keep shared option types (`AnalyzeOptions`, `OutputOptions`, etc.) in
  `internal/app` for backward compatibility with `internal/cli` callers.
- New Team use-cases belong in `internal/app/team/<name>`, not in the top-level
  `internal/app` package.

**Status (2026-05-01):** not yet decomposed. Must be resolved before Team Phase 1
adds any new use-cases. Tracked as a blocking pre-condition for Team Phase 1.

### Core/Team boundary enforcement

The Core/Team commercial boundary is currently prose-documented (see
`docs/release-boundary.md`) and not enforced by Go build tags or package structure.

**Decision (2026-05-01):** accept prose documentation as sufficient for the
current Core-only surface. Enforce via code structure when Team Phase 1 begins:
Team use-cases must live in a distinct package path (e.g. `internal/app/team/`)
that is absent from all Core command import chains. A CI lint step verifying that
Core command files do not import `internal/app/team` should ship alongside the
first Team feature, not after.

Build tag enforcement is not required before Team Phase 1 begins. It is required
to ship with the first Team feature.
