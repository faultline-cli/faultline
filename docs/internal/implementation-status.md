# Implementation Status

Faultline is currently implemented as a CLI-first deterministic product with both log-analysis and source-detector surfaces.

## Shipped Scope

- CLI entrypoint in [`cmd/`](../../cmd)
- command wiring in [`internal/cli/`](../../internal/cli) and use-case orchestration in [`internal/app/`](../../internal/app)
- deterministic analysis engine in [`internal/engine/`](../../internal/engine)
- bundled and extra-pack catalog loading in [`internal/playbooks/`](../../internal/playbooks)
- deterministic matching and scoring in [`internal/matcher/`](../../internal/matcher) and [`internal/scoring/`](../../internal/scoring)
- log and source detector implementations in [`internal/detectors/`](../../internal/detectors)
- terminal, markdown, JSON, workflow, replay, compare, and trace rendering in [`internal/output/`](../../internal/output) and [`internal/renderer/`](../../internal/renderer)
- local repository enrichment in [`internal/repo/`](../../internal/repo)
- optional single-repo local history in [`internal/store/`](../../internal/store) and recurrence signatures in [`internal/signature/`](../../internal/signature)
- additive workflow hints, metrics, and policy outputs in [`internal/workflow/`](../../internal/workflow), [`internal/metrics/`](../../internal/metrics), and [`internal/policy/`](../../internal/policy)
- fixture ingestion, sanitization, review, promotion, and stats in [`internal/fixtures/`](../../internal/fixtures)
- hidden maintainer authoring scaffold support in [`internal/authoring/`](../../internal/authoring)

## Current Public Surface

The repository currently ships these user-visible commands:

- core path: `analyze`, `workflow`, `list`, `explain`, `fix`
- companion surfaces: `trace`, `replay`, `compare`, `inspect`, `guard`, `packs`

Important current behavior:

- `analyze` supports `terminal`, `markdown`, and `json` output
- `workflow` supports `local` and `agent` modes and emits `workflow.v1` JSON
- `inspect` and `guard` expose the source detector through repository-local checks
- `replay` and `compare` operate on saved analysis artifacts without re-running matching
- `packs install` and `packs list` support optional extra catalog composition
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
- `make review` remains clean after playbook or pattern changes
- `make cli-smoke` passes against checked-in examples and companion-command snapshots
- `make release-check VERSION=<tag>` passes before a release cut

## Notes On Scope

- Faultline is no longer a service or frontend product; the CLI is the product.
- The repository includes local single-repo history/store support, but the
  locked product boundary treats cross-repo correlation, aggregation,
  reporting, and recurring-failure coordination as Team-layer value.
- Experimental or hidden paths such as provider-backed delta, hooks, and authoring helpers remain outside the default onboarding story even though their implementation exists in the repository.
