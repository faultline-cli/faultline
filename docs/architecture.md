# Architecture

Faultline keeps the shipped CLI surface stable, but the runtime is split into
explicit deterministic layers:

- `internal/cli` owns Cobra command definitions, flags, stdin/file handling,
  and handing structured options into the app layer.
- `internal/app` owns command use-cases such as analyze, inspect, fix, list,
  explain, workflow, and fixture-corpus operations.
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

That keeps today’s repo layout working while giving the starter repository a
clean default and letting premium packs live in a separate repository loaded
through an external directory.

Additional packs can be composed on top of the bundled starter catalog through
the `FAULTLINE_PLAYBOOK_PACKS` environment variable or repeatable
`--playbook-pack` flags. Faultline also auto-loads any packs installed under
`~/.faultline/packs/`, which is the persistent user-level upgrade path for
premium playbooks. A full `--playbooks` override still resolves a single custom
catalog root and does not combine with extra packs.

Starter pack composition should stay generous for adoption: broad coverage for
common CI failures across popular ecosystems, plus a minimal source-detector
baseline so `inspect` is useful without an extra install. Premium packs should
concentrate on provider-specific depth, advanced deployment or operations
workflows, and deeper source or security rules.

This same `~/.faultline/packs/` convention is used by the Docker image at
`/home/faultline/.faultline/packs`, so a mounted user directory can enable the
same premium pack set in both local and containerized runs.

## Detector boundary

Detectors stay explicit and separate:

- `log` consumes normalized log lines and lightweight log context
- `source` consumes a repository snapshot plus optional changed-file metadata

Both emit the shared `model.Result` shape so ranking, output, workflow, and
history remain stable across command surfaces.

## Rendering boundary

Human-facing longform content is stored in markdown-capable playbook fields such
as `summary`, `diagnosis_markdown`, `fix_markdown`, and
`validation_markdown`.

- markdown is presentation content only
- structured playbook fields still drive matching and ranking
- `--format markdown` emits markdown source from the same deterministic content model
- `--json` bypasses terminal styling entirely
- non-TTY and no-color environments fall back to plain output
