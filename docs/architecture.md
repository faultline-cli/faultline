# Architecture

Faultline keeps the shipped CLI surface stable, but the runtime is split into
explicit deterministic layers:

- `internal/cli` owns Cobra command definitions, flags, stdin/file handling,
  and handing structured options into the app layer.
- `internal/app` owns command use-cases such as analyze, inspect, fix, list,
  explain, workflow, guard, compare, replay, trace, and fixture-corpus operations.
- `internal/store` owns optional durable local forensic memory, deterministic
  signature hashing, SQLite persistence, and explicit schema migrations.
- `internal/authoring` owns the hidden maintainer-only scaffold flow that turns
  a sanitized log into a deterministic candidate playbook YAML.
- `internal/compare` owns deterministic diffing of two saved analysis artifacts
  into a structured `Report` (diagnosis change, evidence delta, repo-context
  delta, and delta-signal changes).
- `internal/engine` owns analysis orchestration and depends on explicit
  collaborators for playbook catalogs, detector lookup, source loading, and
  git enrichment. It does not own persistence.
- `internal/engine/delta` owns explicit provider-backed failure delta
  resolution and minimal cross-run extraction such as changed files and newly
  failing tests.
- `internal/fixtures` owns deterministic fixture corpora, public-source
  ingestion adapters, normalization, review metadata, promotion flow, and
  regression statistics.
- `internal/hooks` owns constrained playbook hook execution, policy gating,
  typed hook handlers, and additive confidence refinement.
- `internal/metrics` owns deterministic reliability metric calculation from
  explicit local history and optional supplied history artifacts.
- `internal/detectors` owns the detector registry plus the distinct `log` and
  `source` detector implementations.
- `internal/playbooks` owns catalog resolution, YAML loading, validation, and
  deterministic review helpers.
- `internal/policy` owns the advisory recommendation layer derived from metrics.
- `internal/scoring` owns the optional Bayesian-inspired evidence-fusion layer
  used for additive reranking explanations and delta diagnosis.
- `internal/output` owns command-facing output selection plus JSON/workflow
  serialization, focused views (`--view summary|evidence|trace|fix|raw`), compare
  formatting, and evidence-only views.
- `internal/renderer` owns terminal-aware human rendering, including quick
  (default) and detailed modes, plain fallback, markdown rendering, and
  restrained ANSI styling.
- `internal/trace` owns per-playbook rule-by-rule trace payloads used by
  `faultline trace` and `faultline analyze --trace`.

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
`--playbook-pack` flags. Faultline also auto-loads any packs installed under
`~/.faultline/packs/`, which is the persistent user-level install path for
extra playbook packs. A full `--playbooks` override still resolves a single
custom catalog root and does not combine with extra packs.

Installed packs record a small manifest at install time so the analysis object
can carry deterministic provenance:

- pack name
- version
- source URL or local source path
- pinned ref when available
- playbook count contributed by that pack

This provenance is additive. It does not change matching; it makes the loaded
catalog auditable in analysis JSON and `packs list`.

For local validation against an external pack checkout, the repository can use
the ignored symlink at `playbooks/packs/extra-local` or an explicit
`EXTRA_PACK_DIR` value. The corresponding deterministic checks are
`make extra-pack-check` and `make extra-pack-review`.

Bundled catalog composition should stay generous for adoption: broad coverage for
common CI failures across popular ecosystems, plus a minimal source-detector
baseline so `inspect` is useful without an extra install. Extra packs can
concentrate on provider-specific depth, advanced deployment or operations
workflows, and deeper source or security rules.

The same pack boundary now carries optional hook overlays through
`faultline-hooks.yaml` at the pack root. Those overlays:

- define reusable named typed hooks
- attach verify, collect, or remediate hooks to existing playbooks by ID
- allow later packs to disable or override earlier hook definitions
- remain additive to the playbook catalog instead of redefining matching logic

Hook execution itself stays explicit and local. The analysis engine still owns
matching and ranking; hooks are an evidence refinement layer that runs only
when the user opts in through the hidden hooks flag.

This same `~/.faultline/packs/` convention is used by the Docker image at
`/home/faultline/.faultline/packs`, so a mounted user directory can enable the
same installed pack set in both local and containerized runs.

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
- optional at runtime
- additive to the existing analysis path

The ownership split is deliberate:

- `internal/app` resolves store config, opens the store, handles graceful
  degradation, enriches results with history, and records completed runs
- `internal/store` hides SQL, migrations, and schema details behind a small
  interface plus a no-op fallback
- `internal/engine` stays deterministic and store-agnostic; it returns analysis
  results without querying or mutating on-disk state
- detectors remain stateless in v1 and do not read from the store directly

The store records durable forensic memory such as:

- top-diagnosis recurrence by `signature_hash`
- run-level `input_hash` and `output_hash`
- ranked playbook matches for longitudinal review
- hook execution results when hooks are enabled

The store does not become a generic raw-log warehouse. By default it stores
hashes, normalized signature material, minimal evidence excerpts, and small
structured summaries only.

## Scoring boundary

Faultline now has an explicit three-layer ranking model:

1. detectors decide which playbooks matched
2. `internal/scoring` may rerank those already-matched candidates when
   `--bayes` is enabled, and it only emits delta hints when repo-aware context
   is explicit
3. output, workflow, and guard consume the final deterministic ordering

That boundary matters:

- detectors remain authoritative
- scoring is assistive, not a second matcher
- same input and same repo snapshot still produce the same output
- ranking and delta payloads are additive and explainable
- changed files are suspicious context, not proof on their own
- provider-backed delta remains opt-in and should stay narrow: compare the
  current failing run against the last successful run on the same branch,
  extract deterministic diffs, and feed them back into the same scoring model

## Rendering boundary

Human-facing longform content is stored in markdown-capable playbook fields such
as `summary`, `diagnosis`, `fix`, and
`validation`.

- markdown is presentation content only
- structured playbook fields still drive matching and ranking
- CLI commands render the same deterministic content model to terminal or markdown output
- `--format json` and `--json` emit the structured machine-readable form
- `--view summary|evidence|trace|fix|raw` selects a focused slice of the human-readable output
  without changing the underlying analysis; `summary` and `raw` map to quick and
  detailed rendering modes respectively; `evidence` and `fix` emit narrow single-purpose
  slices of the top result; `trace` routes to deterministic rule-by-rule playbook evaluation
- replayed analysis artifacts support `summary|evidence|fix|raw`; trace replay requires a
  saved trace artifact or rerunning `faultline trace` on the original log
- non-TTY and no-color environments fall back to plain output

The stable analysis JSON schema is additive. Beyond the ranked results, it may
also include:

- `pack_provenance` when one or more packs contributed playbooks
- `metrics` when sufficient explicit history exists to compute TSS, FPC, or PHI
- `policy` when a deterministic advisory recommendation can be derived from
  those metrics
- `input_hash` and `output_hash` for repeated-run determinism checks
- result-level `signature_hash`, recurrence fields, and hook history summaries

Saved analysis artifacts preserve those fields on replay and compare.

When hook execution is enabled, result JSON may also include additive `hooks`
reports. These reports record:

- execution mode
- base confidence, total confidence delta, and final confidence
- per-hook status (`executed`, `blocked`, `skipped`, or `failed`)
- structured hook facts and captured evidence excerpts

Absent hook execution remains absent from output.

Workflow artifacts are derived from the same deterministic analysis object.
When present, JSON and text workflow output may also carry:

- `ranking_hints`
- `delta_hints`
- `metrics_hints`
- `policy_hints`

Absent data remains absent. Faultline does not invent placeholder values for
missing history or policy inputs.

## Compare boundary

`faultline compare` is a deterministic companion for diffing two saved analysis
artifacts (produced by `faultline analyze --json` or `faultline replay --json`).
It does not re-run analysis; it only compares the stored payloads.

- diagnosis change is detected by comparing the top `failure_id` across both artifacts
- evidence, repo context, and delta-signal fields are diffed as ordered string sets
- output is stable and machine-readable with `--json`, or human-readable via terminal and markdown format
- the compare report is intentionally narrow: it surfaces what changed, not why
