# Release Verification Matrix

Faultline's release contract is deterministic: public behavior is validated by
tests, fixture gates, playbook review, and CLI smoke snapshots before a release.

| Surface | Required checks |
| --- | --- |
| `analyze`, `fix`, JSON output, artifact fields | `go test ./...`, `make fixture-check`, `make cli-smoke` |
| `workflow`, `workflow explain/apply` | `go test ./internal/workflow ./internal/app ./cmd`, `make cli-smoke` |
| `list`, `explain`, `packs`, catalog composition | `go test ./internal/playbooks ./internal/cli ./cmd`, `make cli-smoke` |
| `trace`, `replay`, `compare`, focused views | `go test ./internal/output ./internal/trace ./internal/compare ./cmd`, `make cli-smoke` |
| `inspect`, `guard`, source detectors | `go test ./internal/detectors/... ./internal/engine ./internal/app`, `make cli-smoke` |
| Playbook patterns and scoring | `go test ./internal/matcher ./internal/scoring ./internal/engine`, `make fixture-check`, `make review` |
| Hidden maintainer fixture workflows | `go test ./internal/fixtures ./internal/app ./internal/cli`, `make fixture-check` |
| Release archives | `make release-verify VERSION=<tag>` |

`make release-verify` runs the normal release-hardening sequence. With
`VERSION=dev` it skips archive smoke tests; with a tag it also builds and smokes
the release archive. Set `WITH_DOCKER=1` to include Docker smoke validation.
