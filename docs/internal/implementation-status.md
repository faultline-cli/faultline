# Implementation Status

Faultline is currently implemented as a CLI-first deterministic product centered on log analysis.

## Shipped Scope

- CLI entrypoint in [`cmd/`](../../cmd)
- command wiring in [`internal/cli/`](../../internal/cli) and use-case orchestration in [`internal/app/`](../../internal/app)
- deterministic analysis engine in [`internal/engine/`](../../internal/engine)
- bundled and extra-pack catalog loading in [`internal/playbooks/`](../../internal/playbooks)
- deterministic matching and scoring in [`internal/matcher/`](../../internal/matcher) and [`internal/scoring/`](../../internal/scoring)
- log and source detector implementations in [`internal/detectors/`](../../internal/detectors)
- terminal, markdown, JSON, and workflow rendering in [`internal/output/`](../../internal/output) and [`internal/renderer/`](../../internal/renderer)
- local repository enrichment in [`internal/repo/`](../../internal/repo)
- optional single-repo local history in [`internal/store/`](../../internal/store) and recurrence signatures in [`internal/signature/`](../../internal/signature)
- additive workflow hints in [`internal/workflow/`](../../internal/workflow)
- fixture ingestion, sanitization, review, promotion, and stats in [`internal/fixtures/`](../../internal/fixtures)

## Current Public Surface

The repository currently ships these user-visible commands:

- core path: `analyze`, `batch`, `workflow`, `list`, `explain`, `fix`
- source inspection path: `inspect`
- companion surface: `report`
- hidden maintainer workflows: `fixtures`

Important current behavior:

- `analyze` supports `terminal`, `markdown`, and `json` output
- `batch` analyzes multiple local CI logs independently and groups matches by failure pattern
- `inspect` scans local source trees through deterministic source detectors
- `workflow` supports `local` and `agent` modes and emits `workflow.v1` JSON
- `report` reads only the local forensic store and remains a bounded companion surface
- explicit `--playbook-pack` flags support optional local catalog composition
- hidden maintainer-only `fixtures` commands remain available for corpus curation

## Repository State

- the repository structure matches the deterministic CLI architecture described in [`SYSTEM.md`](../../SYSTEM.md)
- release archives bundle the binary plus `playbooks/bundled/`
- Docker packaging follows the same bundled-playbook contract
- the checked-in examples under [`examples/`](../../examples) are used for snapshot-backed CLI smoke validation
- the checked-in corpus under [`fixtures/real/`](../../fixtures/real) is the regression proof for shipped matching behavior

## Validation Baseline

The repository is in the expected current state when these remain true:

- `go test ./...` passes
- `make fixture-check` passes on the accepted real corpus baseline
- `make bayes-check` passes with no ranking regressions
- `make review` remains clean after playbook or pattern changes
- `make docs-check` confirms generated failure docs are current
- `make cli-smoke` passes against checked-in examples and command snapshots
- `make release-check VERSION=<tag>` passes before a release cut

## Notes On Scope

- Faultline is no longer a service or frontend product; the CLI is the product.
- The repository includes local single-repo history/store support, but the
  locked product boundary treats cross-repo correlation, aggregation,
  reporting, and recurring-failure coordination as Team-layer value.
- Hidden paths are limited to maintainer fixture curation and release gates.
