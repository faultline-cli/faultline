# Playbook Robustness

Faultline playbooks must stay deterministic as the catalog grows. Coverage is
therefore measured as behavioral evidence, not only as file count.

The robustness report answers two questions for each resolved playbook:

- false-positive risk: how likely the playbook is to match a nearby failure it
  should not own
- false-negative risk: how likely the playbook is to miss legitimate variants
  of its root cause

Run it with:

```sh
go run ./cmd coverage
go run ./cmd coverage --json
go run ./cmd coverage --playbook-pack path/to/pack --json
```

The `--playbook-pack` form is important. Robustness is computed after bundled
and extra packs are composed, so external playbook sets are evaluated against
the real catalog they will run with.

## Current Snapshot

Snapshot date: 2026-05-15.

Current bundled catalog:

- total bundled playbooks: 173
- positive fixture-backed playbooks: 170
- source-detector playbooks: 3, covered by source fixture tests
- fixture assertions: 392 positive, 16 negative
- strict top-1 fixtures: 87
- duplicate IDs: 0

The catalog now has broad positive fixture coverage and stronger strict top-1
coverage for previously low-scoring rules. The main remaining robustness gap is
adversarial coverage: too few fixtures assert that nearby playbooks must not
match.

Recently improved rules:

| Playbook | Score | FP risk | FN risk | Why it matters |
| --- | ---: | --- | --- | --- |
| `circleci-resource-class-oom` | 75 | high | low | OOM wording overlaps killed-process and Node memory failures, now backed by strict/noisy variants. |
| `config-file-missing` | 75 | high | low | Missing-file wording is broad, now backed by `.env`, settings, and noisy config variants. |
| `ipv6-ipv4-resolution` | 75 | high | low | DNS and network-address wording overlaps several network rules, now backed by IPv6-specific variants. |
| `node-missing-executable` | 75 | high | low | Inherits generic missing executable evidence, now backed by Node-only runtime variants. |
| `process-killed-no-logs` | 75 | high | low | Generic termination wording overlaps OOM, timeout, and provider resource failures, now backed by SIGKILL/SIGTERM variants. |
| `formatting-failure` | 75 | high | low | Formatter, linter, and generic quality-gate wording are close neighbors, now backed by Prettier/Black variants. |

Lowest current robustness scores:

| Playbook | Score | FP risk | FN risk | Why it remains next |
| --- | ---: | --- | --- | --- |
| `config-mismatch` | 62 | medium | high | Only one positive fixture and no strict top-1 fixture. |
| `dependency-drift` | 62 | medium | high | Only one positive fixture and several dependency-resolution neighbors. |
| `disk-full` | 62 | medium | high | Needs broader variants across host disk, container disk, and runner quota wording. |
| `dns-enotfound` | 62 | medium | high | Needs distinct ENOTFOUND variants that stay separate from generic DNS and IPv6 rules. |
| `entrypoint-misconfigured` | 62 | medium | high | Needs variants that separate entrypoint mistakes from missing executable and architecture mismatches. |
| `expired-credentials` | 62 | medium | high | Needs provider-token variants and auth-family near misses. |
| `github-actions-syntax` | 62 | medium | high | Needs workflow parser variants and CircleCI/GitLab syntax near misses. |
| `gitlab-ci-artifact-expired` | 62 | medium | high | Needs GitLab artifact-expiry variants separate from upload failures. |

Medium-risk one-fixture playbooks should be promoted next with real or noisy
fixtures plus at least one near-miss. Examples include `config-mismatch`,
`dependency-drift`, `disk-full`, `dns-enotfound`, `entrypoint-misconfigured`,
`expired-credentials`, `github-actions-syntax`,
`gitlab-ci-artifact-expired`, `go-data-race`, `jest-worker-crash`,
`multistage-build-missing-artifact`, `npm-peer-dependency-conflict`,
`orphaned-resources`, `package-manager-mismatch`, `path-case-mismatch`,
`pip-hash-mismatch`, `port-conflict`, and `ssh-key-auth`.

## Scoring Contract

The report is intentionally heuristic, but all inputs are deterministic:

- positive fixture count per playbook
- disallowed-playbook assertion count per playbook
- strict top-1 fixture count per playbook
- shared positive pattern conflicts from the composed catalog
- explicit `match.none` exclusions protecting nearby rules
- detector type

False-positive risk:

- `high`: seven or more shared positive pattern conflicts
- `medium`: three to six shared positive pattern conflicts
- `medium`: silent-failure playbooks by default, because absence and continuation
  signals are easy to over-match
- `low`: fewer shared positive conflicts

False-negative risk:

- `high`: zero or one positive fixture
- `medium`: two positive fixtures
- `medium`: source detector playbooks, because their source fixtures are tracked
  outside log fixture counts
- `low`: three or more positive fixtures

Confidence score starts at 86, subtracts 8 for each medium risk and 16 for each
high risk, then adds:

- 6 for five or more positive fixtures, or 3 for three or four positive fixtures
- 3 for three or more negative assertions
- 2 for at least one strict top-1 fixture

Scores are clamped to 45-95 and sorted ascending in the text report. The score
is not a statistical probability; it is a review priority signal.

## Improvement Plan

1. Raise the minimum release bar for every shipped log playbook:
   one canonical fixture, one real or noisy fixture, one strict top-1 assertion,
   and one near-miss fixture for each obvious confusable neighbor.

2. Promote adversarial fixtures before adding new playbooks. Prefer a small
   fixture that proves "this similar thing must not match" over another
   positive variant for an already-covered rule.

3. Tighten broad playbooks first:
   `config-file-missing`, `ipv6-ipv4-resolution`, `node-missing-executable`,
   `process-killed-no-logs`, `formatting-failure`, and `quality-gate-failure`.
   Move generic `match.any` phrases into `match.all` or `match.partial`, add
   `within_lines` when evidence should be local, and add `match.none` exclusions
   only when a concrete near-miss proves the boundary.

4. Treat confusable families as review units:
   missing files, missing executables, permissions, auth, DNS/network,
   timeout, killed/OOM, config/schema, quality/lint/format, and test runners.
   A playbook changed in one family should be checked against fixtures for its
   neighbors in the same family.

5. Keep extra packs on the same rails. Pack authors should ship fixtures beside
   their pack and run composed coverage with `--playbook-pack`. A pack is not
   ready for default use if it lowers bundled top-1 behavior or introduces broad
   shared patterns without near-miss protection.

6. Use the existing deterministic gates after every batch:

```sh
make review
make test
make fixture-check
```

Run `make cli-smoke` when user-facing output, examples, packaging, or release
paths change.

## Extensibility Rules

New playbook sets should be additive and reviewable:

- prefer extending a bundled playbook when the root cause is the same
- add a new playbook only for a distinct root cause with distinctive evidence
- keep matching logic in structured fields, never in prose
- include pack-local fixtures for positive, noisy, strict top-1, and near-miss
  behavior
- run robustness against the composed catalog, not the pack in isolation
- document any intentional broad overlap with the nearby playbook it protects
  or depends on

This keeps the catalog scalable without introducing fuzzy matching or
non-deterministic classification.
