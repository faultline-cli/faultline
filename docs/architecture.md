# Architecture

Faultline keeps the shipped CLI surface stable, but the runtime is split into
explicit deterministic layers:

- `internal/cli` owns Cobra command definitions, flags, stdin/file handling,
  and handing structured options into the app layer.
- `internal/app` owns command use-cases such as analyze, inspect, fix, list,
  explain, workflow, guard, and fixture-corpus operations.
- `internal/engine` owns analysis orchestration and depends on explicit
  collaborators for playbook catalogs, detector lookup, history persistence,
  source loading, and git enrichment.
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
  serialization.
- `internal/renderer` owns terminal-aware human rendering, including plain
  fallback, markdown rendering, and restrained ANSI styling.

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
   `--bayes` is enabled
3. output, workflow, and guard consume the final deterministic ordering

That boundary matters:

- detectors remain authoritative
- scoring is assistive, not a second matcher
- same input and same repo snapshot still produce the same output
- ranking and delta payloads are additive and explainable

## Rendering boundary

Human-facing longform content is stored in markdown-capable playbook fields such
as `summary`, `diagnosis`, `fix`, and
`validation`.

- markdown is presentation content only
- structured playbook fields still drive matching and ranking
- CLI commands render the same deterministic content model to terminal or markdown output
- `--format json` and `--json` emit the structured machine-readable form
- non-TTY and no-color environments fall back to plain output
