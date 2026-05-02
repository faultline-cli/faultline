# Repository Cleanup Assessment

Date: 2026-05-02

## Scope

This pass keeps the default release story centered on `analyze`, `workflow`,
`list`, `explain`, and `fix`. Cleanup decisions favor deterministic coverage,
fixture-backed playbooks, generated docs that match shipped YAML, and maintainer
tools that stay outside the first-run narrative.

## Least Valuable Playbooks

The lowest-value bundled playbooks were the 20 removed during the playbook triage
pass:

- low-signal cache and topology log rules: `remote-cache-miss`,
  `topology-ownership-mismatch`, `topology-failure-clustered`
- absence-of-run or overlapping CI workflow rules: `step-skipped-unexpectedly`,
  `workflow-not-triggered`
- niche documentation or project-specific suites: `link-checker-failure`,
  `asciidoctor-jbehave-test-failure`, `flexget-test-failure`,
  `mdanalysis-test-failure`, `nikola-build-test-failure`,
  `nupic-test-failure`, `openai-gym-test-failure`,
  `pip-install-test-failure`, `python-parser-test-failure`,
  `readthedocs-build-test-failure`, `sentry-elasticsearch-test-failure`,
  `translate-toolkit-test-failure`, `youtube-dl-test-failure`
- weak root-cause inference rules: `clock-drift`, `random-seed-not-fixed`

Remaining low-confidence areas are not removal candidates yet, but they should
not be promoted in the product narrative until fixture evidence improves:

- the silent-failure family, currently 0 / 8 with positive fixture evidence in
  `go run ./cmd coverage`
- source-detector playbooks that are validated outside the real-log fixture
  accounting path
- provider-specific CI rules for Azure Pipelines, CircleCI, GitLab scheduler
  state, and Jenkins until each has accepted real fixtures

## Least Valuable Modules

No Go package is currently dead: `go list ./...` resolves every package, and the
companion surfaces are wired through CLI or tests. The least valuable modules
are therefore maintenance-risk candidates rather than deletion candidates:

- `internal/engine/delta`: explicit experimental provider API path; high
  complexity and network behavior, but gated by environment variables.
- `internal/authoring`: useful maintainer scaffold support, but hidden and not a
  release-path dependency for normal users.
- `internal/metrics` and `internal/policy`: advisory outputs with limited
  current user-facing value until recurrence and Team-layer decisions mature.
- `tools/eval-corpus`: valuable for catalog strategy, but large and separate
  from the release-critical CLI path.
- `internal/coverage`: important as a maintainer gate, but it currently needs
  better accounting for silent and source-detector fixtures.

## Required Update

- Keep the generated failure catalog at 173 playbooks and the real-fixture
  baseline at 215 accepted fixtures.
- Resolve the one weak real-fixture match:
  `stackexchange-stackoverflow-66973433-s4-52f19765e92d43ae`.
- Close the 57-playbook positive-fixture backlog reported by
  `go run ./cmd coverage`, prioritizing high-frequency classes before
  provider-specific long tail.
- Teach coverage reporting to account for silent-detector and source-detector
  regression fixtures so bundled non-log surfaces do not look uncovered forever.
- Re-run the large GitHub Actions evaluation against the post-triage 173-playbook
  catalog before making new public coverage claims.

## Technical Debt Surface

- Generated docs previously allowed orphaned playbook pages to remain after
  playbook deletion. `make docs-check` now flags stale generated docs, and
  `make docs-generate` prunes them.
- The command surface is intentionally broad for companion workflows. Help text
  and docs should continue to keep non-default commands out of onboarding.
- The coverage gate measures fixture evidence more directly than historical docs
  did, but it still undercounts non-log detector proof.
- Some internal planning docs describe older versions or design options. Keep
  them under `docs/internal/` and avoid linking them as user-facing truth.

## Next Steps

1. Promote fixture evidence in small batches for the 57 remaining uncovered
   playbooks, running `make review`, `make test`, and `make fixture-check` after
   each batch.
2. Add coverage accounting for silent and source fixtures before expanding those
   playbook families.
3. Re-evaluate the large external corpus after the playbook removal pass and
   update `docs/fixture-corpus.md` only with clearly scoped claims.
4. Keep `internal/engine/delta`, `internal/metrics`, `internal/policy`, and
   `internal/authoring` hidden or companion-only until they have a stronger
   release-grade story.
