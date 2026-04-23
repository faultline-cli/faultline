# Fixture Corpus

Faultline's trust boundary is the checked-in corpus, not a vague accuracy claim. The current snapshot below reflects the accepted real fixtures and bundled playbooks in this repository.

## Current Snapshot

- Bundled playbooks: 101
- Accepted real fixtures: 145
- Top-1 match rate: 100% (145/145)
- Top-3 match rate: 100% (145/145)
- Unmatched fixtures: 0
- False positives: 0
- Weak matches: 0 (0.0%)
- Fixture metadata validation: required for real and staging corpora
- Corpus fingerprint drift: release-gated through `fixtures/real/baseline.json`
- Test corpus files: 32 (release-gated through `corpus_test.go`)
- Source-detector regression fixtures: 8 repository scenarios under `internal/engine/testdata/source/`

These numbers describe the checked-in regression corpus only. They are useful because they are deterministic, reviewable, and reproducible from the repository state.

## Why It Matters

- The corpus is built from accepted real-world CI failures under `fixtures/real/`.
- Ranking changes are gated by regression checks instead of hidden online learning.
- Playbook coverage, thresholds, and baseline behavior are visible in version control.
- Automation consumers can audit the proof artifact that backs the product's trust story.

## Coverage By Failure Class

Bundled playbook coverage by category (from `playbooks/bundled/`):

| Category | Bundled Playbooks |
| --- | --- |
| build | 33 |
| ci | 20 |
| test | 13 |
| runtime | 14 |
| deploy | 8 |
| auth | 7 |
| network | 6 |

Accepted real fixtures mapped through expected playbooks (from `fixtures/real/`):

| Category | Accepted Real Fixtures |
| --- | --- |
| build | 38 |
| network | 29 |
| ci | 32 |
| runtime | 21 |
| auth | 10 |
| deploy | 9 |
| test | 6 |

This table is intended as public proof coverage, not a claim that unknown failures are solved.

## Release Snapshot Trend

Starting snapshot table for release-over-release tracking:

| Snapshot | Bundled Playbooks | Accepted Real Fixtures | Test Corpus | Top-1 | Top-3 | Unmatched | False Positive |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-04-17 baseline (`fixtures/real/baseline.json`) | 77 | 103 | 16 | 100% | 100% | 0 | 0 |
| 2026-04-18 corpus expansion | 77 | 103 | 30 | 100% | 100% | 0 | 0 |
| 2026-04-22 v0.4 refinement pass | 100 | 113 | 30 | 100% | 100% | 0 | 0 |
| 2026-04-22 test-gap closure pass | 100 | 116 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-22 adapter-diversity pass | 100 | 118 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-23 ingestion pass | 100 | 120 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-23 mixed-source ingestion pass | 100 | 123 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-23 auth-and-lock ingestion pass | 100 | 126 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-24 mixed-source ingestion + source-boundary refinement | 101 | 145 | 32 | 100% | 100% | 0 | 0 |

Append one row per release cut so corpus growth and match stability stay visible over time.

## Coverage Observations

- The shipped corpus is still strongest on build, network, and CI failures, with 99 of 145 accepted fixtures concentrated in those three categories.
- Provider diversity is still uneven: the accepted real corpus is 87 GitHub fixtures, 25 Stack Exchange fixtures, 14 GitLab fixtures, 9 Reddit fixtures, and 10 Discourse fixtures.
- The `test` category is no longer a single-fixture blind spot, but it is still the thinnest accepted real slice relative to the 13 bundled test playbooks. Future ingestion should keep biasing toward high-signal test failures before adding more test rules.
- Source-detector rules are now regression-gated separately from the real log corpus. That keeps the trust boundary honest, but it also means source-surface expansion should come with paired repository fixtures, not just more YAML. The shipped source surface now includes 8 repository regression scenarios, including negative fixtures for virtualenv and test-only noise boundaries.
- Signature and recurrence behavior is also pressure-tested separately through
  the noisy variant corpus under `internal/signature/testdata/variants/` plus
  store-backed app tests. That keeps recurrence grouping and output
  determinism reviewable without inflating the release-gated real corpus.

## Contribution Prompt

If Faultline misses a failure class, contribute a sanitized public log:

1. Open an issue with an anonymized failing snippet and environment context.
2. Include a public source URL when possible (issue, discussion, or forum thread).
3. Avoid secrets, private hostnames, and internal repository names.
4. Mark expected diagnosis if known to speed triage.

Maintainers should route accepted cases through the deterministic ingest/review/promote flow:

```bash
# Ingest from a public source URL.
faultline fixtures ingest --adapter github-issue --url https://github.com/owner/repo/issues/123

# Sanitize known credential and identity patterns in the staging fixture.
faultline fixtures sanitize <staging-id>

# Preview replacements without modifying the file.
faultline fixtures sanitize <staging-id> --dry-run

# Review predicted matches and duplicate signals.
faultline fixtures review

# Draft a candidate playbook from the sanitized fixture when a real gap exists.
faultline fixtures scaffold --from-fixture <staging-id> --category <category>

# Promote to fixtures/real/ with a locked expectation.
faultline fixtures promote <staging-id> --expected-playbook <id>
```

### Sanitizer Scope and Limitations

`faultline fixtures sanitize` applies a deterministic, rule-based pass over the `raw_log` and `normalized_log` fields of a staging fixture. It masks:

- GitHub personal access tokens and app tokens
- AWS access key IDs
- `Authorization: Bearer/Token/Basic` header values
- Credentials embedded in URLs (`https://user:password@host`)
- Credential key=value and key: value pairs (`password=`, `secret=`, `api_key=`, `access_token=`, `auth_token=`, `private_key=`)
- JWT tokens
- PEM-encoded private key blocks
- Email addresses

The sanitizer does **not** catch:

- Internal hostnames, company-specific domain names, and service endpoint URLs
- Signed or pre-authorized URLs that do not contain obvious credential prefixes
- Customer or tenant identifiers embedded in paths or query strings
- Usernames, project names, or repository names that could identify a contributor's environment
- Any secret that does not match the above explicit patterns

Always inspect the fixture manually after running the sanitizer. The tool is a first-pass aide, not a guarantee of complete redaction.

### Scaffold Helper

`faultline fixtures scaffold` is a hidden maintainer helper that composes with
the staging workflow above. It can read from:

- `--from-fixture <staging-id>`
- `--log <path>`
- a positional log path
- stdin

The command applies the same deterministic sanitizer pass before extracting
candidate match patterns, then emits a draft YAML playbook with required fields
and `TODO` markers. It is useful for bootstrapping a new rule, but it does not
replace manual review, `make review`, or the fixture regression gates.

## Regenerate And Check

Build the CLI first:

```bash
make build
```

Re-run the real-fixture regression gate:

```bash
./bin/faultline fixtures stats --class real --check-baseline
```

Run the shipped CLI smoke checks against the checked-in examples:

```bash
make cli-smoke
```

Inspect the same report as JSON:

```bash
./bin/faultline fixtures stats --class real --json --check-baseline
```

Recompute the broader combined corpus if needed:

```bash
./bin/faultline fixtures stats --class all
```

Verify the bundled catalog size exposed by the CLI:

```bash
./bin/faultline list | tail -n +2 | wc -l
```

## Source Of Truth

- Bundled playbooks live under `playbooks/bundled/`.
- Accepted real fixtures live under `fixtures/real/`.
- Source-detector regression fixtures live under `internal/engine/testdata/source/`.
- Test corpus files live under `internal/engine/testdata/corpus/`.
- Signature and recurrence variant fixtures live under `internal/signature/testdata/variants/`.
- The test corpus validates playbook matching through `corpus_test.go` as part of `make test`.
- The checked-in regression baseline is `fixtures/real/baseline.json`.
- The fixture commands are wired through `faultline fixtures stats`.
- Source provenance and adapter counts are included in `faultline fixtures stats` output.
