# Architecture

Faultline keeps the shipped CLI surface stable, but the runtime is split into
explicit deterministic layers:

- `internal/cli` owns Cobra command definitions, flags, stdin/file handling,
  and handing structured options into the app layer.
- `internal/app` owns command use-cases such as analyze, inspect, fix, list,
  explain, workflow, guard, compare, replay, trace, and fixture-corpus operations.
- `internal/compare` owns deterministic diffing of two saved analysis artifacts
  into a structured `Report` (diagnosis change, evidence delta, repo-context
  delta, and delta-signal changes).
- `internal/engine` owns analysis orchestration and depends on explicit
  collaborators for playbook catalogs, detector lookup, history persistence,
  source loading, and git enrichment.
- `internal/engine/delta` owns explicit provider-backed failure delta
  resolution and minimal cross-run extraction such as changed files and newly
  failing tests.
- `internal/fixtures` owns deterministic fixture corpora, public-source
  ingestion adapters, normalization, review metadata, promotion flow, and
  regression statistics.
- `internal/detectors` owns the detector registry plus the distinct `log` and
  `source` detector implementations.
- `internal/playbooks` owns catalog resolution, YAML loading, validation, and
  deterministic review helpers.
- `internal/scoring` owns the optional Bayesian-inspired evidence-fusion layer
  used for additive reranking explanations and delta diagnosis.
- `internal/output` owns command-facing output selection plus JSON/workflow
  serialization, focused views (`--view summary|evidence|fix|raw`), compare
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

For local validation against an external pack checkout, the repository can use
the ignored symlink at `playbooks/packs/extra-local` or an explicit
`EXTRA_PACK_DIR` value. The corresponding deterministic checks are
`make extra-pack-check` and `make extra-pack-review`.

Bundled catalog composition should stay generous for adoption: broad coverage for
common CI failures across popular ecosystems, plus a minimal source-detector
baseline so `inspect` is useful without an extra install. Extra packs can
concentrate on provider-specific depth, advanced deployment or operations
workflows, and deeper source or security rules.

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
- `--view summary|evidence|fix|raw` selects a focused slice of the human-readable output
  without changing the underlying analysis; `summary` and `raw` map to quick and
  detailed rendering modes respectively; `evidence` and `fix` emit narrow single-purpose
  slices of the top result
- non-TTY and no-color environments fall back to plain output

## Compare boundary

`faultline compare` is a deterministic companion for diffing two saved analysis
artifacts (produced by `faultline analyze --json` or `faultline replay --json`).
It does not re-run analysis; it only compares the stored payloads.

- diagnosis change is detected by comparing the top `failure_id` across both artifacts
- evidence, repo context, and delta-signal fields are diffed as ordered string sets
- output is stable and machine-readable with `--json`, or human-readable via terminal and markdown format
- the compare report is intentionally narrow: it surfaces what changed, not why
