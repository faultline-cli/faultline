# Release Verification Matrix

Faultline's release contract is deterministic: public behavior is validated by
tests, fixture gates, playbook review, and CLI smoke snapshots before a release.

| Surface | Required checks |
| --- | --- |
| `analyze`, `batch`, `fix`, JSON output, artifact fields | `go test ./...`, `make fixture-check`, `make bayes-check`, `make cli-smoke` |
| `workflow` handoff | `go test ./internal/workflow ./internal/app ./cmd`, `make cli-smoke` |
| `inspect` and source detector checks | `go test ./internal/detectors/... ./internal/engine ./internal/app ./cmd`, `make cli-smoke` |
| `list`, `explain`, catalog composition | `go test ./internal/playbooks ./internal/cli ./cmd`, `make cli-smoke` |
| Playbook patterns, scoring, and generated catalog docs | `go test ./internal/matcher ./internal/scoring ./internal/engine`, `make fixture-check`, `make bayes-check`, `make review`, `make docs-check` |
| Hidden maintainer fixture workflows | `go test ./internal/fixtures ./internal/app ./internal/cli`, `make fixture-check` |
| Release archives | `make release-check VERSION=<tag>` or `make release-verify VERSION=<tag>` |

`make release-check` and `make release-verify` both run the normal
release-hardening sequence, including fixture, Bayes, playbook-review,
generated-docs, and CLI-smoke gates. With `VERSION=dev`,
`make release-verify` skips archive smoke tests; with a tag it also builds and
smokes the release archive. Set `WITH_DOCKER=1` to include Docker smoke
validation.
