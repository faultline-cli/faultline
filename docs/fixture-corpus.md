# Fixture Corpus

Faultline's trust boundary is the checked-in corpus, not a vague accuracy claim. The current snapshot below reflects the checked-in real-fixture baseline and bundled playbooks in this repository.

## Current Snapshot

- Bundled playbooks: 193
- Accepted real fixtures: 211
- Top-1 baseline pass rate: 100% (211/211)
- Top-3 baseline pass rate: 100% (211/211)
- Unmatched fixtures: 0 (0.0%)
- False positives: 0
- Weak matches: 0 (0.0%)
- Fixture metadata validation: required for real and staging corpora
- Corpus fingerprint drift: release-gated through `fixtures/real/baseline.json`
- Test corpus files: 32 (release-gated through `corpus_test.go`)
- Source-detector regression fixtures: 8 repository scenarios under `internal/engine/testdata/source/`

These numbers describe the checked-in regression corpus only. They are useful because they are deterministic, reviewable, and reproducible from the repository state. They are baseline-gate numbers, not a claim that every accepted real fixture currently lands as a strong top-1 match.

## Why It Matters

- The corpus is built from accepted real-world CI failures under `fixtures/real/`.
- Ranking changes are gated by regression checks instead of hidden online learning.
- Playbook coverage, thresholds, and baseline behavior are visible in version control.
- Automation consumers can audit the proof artifact that backs the product's trust story.

## Corpus Mix

Real-fixture provider distribution from `./bin/faultline fixtures stats --class real`:

| Provider | Accepted Real Fixtures |
| --- | --- |
| GitHub | 116 |
| Stack Exchange | 70 |
| GitLab | 24 |
| Reddit | 12 |
| Discourse | 6 |

Adapter distribution in the same corpus:

| Adapter | Accepted Real Fixtures |
| --- | --- |
| `github-issue` | 116 |
| `stackexchange-question` | 70 |
| `gitlab-issue` | 24 |
| `reddit-post` | 12 |
| `discourse-topic` | 6 |

These counts describe the public proof corpus, not a claim that unknown failures are solved.

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
| 2026-04-24 mixed-source ingestion + source-boundary refinement | 101 | 138 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-24 diverse-adapter ingestion pass | 101 | 153 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-26 v0.4.1 corpus expansion (228 fixtures) | 170 | 228 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-29 v0.4.4 feature release | 193 | 228 | 32 | 100% | 100% | 0 | 0 |
| 2026-04-30 main baseline after overlap review | 193 | 228 | 32 | 91.2% | 91.2% | 8.8% | 0 |
| 2026-04-30 v0.4.4 corpus hardening pass | 193 | 228 | 32 | 92.1% | 92.1% | 7.9% | 0 |
| v0.4.4 corpus quality pass (removed 17 no-signal fixtures, added if-no-files-found pattern) | 193 | 211 | 32 | 100% | 100% | 0% | 0 |

Append one row per release cut so corpus growth and match stability stay visible over time.

## Coverage Observations

- The corpus mix is still GitHub-heavy: 103 GitHub fixtures (49%) out of 211 accepted real fixtures.
- The current baseline shows 18 unmatched fixtures and zero weak matches. All 18 unmatched are corpus quality issues (YAML workflow configs, positive resolution comments, or successful output with no error signals) that cannot be addressed with playbook changes alone.
- Source-detector rules are now regression-gated separately from the real log corpus. That keeps the trust boundary honest, but it also means source-surface expansion should come with paired repository fixtures, not just more YAML. The shipped source surface now includes 8 repository regression scenarios, including negative fixtures for virtualenv and test-only noise boundaries.
- Signature and recurrence behavior is also pressure-tested separately through
  the noisy variant corpus under `internal/signature/testdata/variants/` plus
  store-backed app tests. That keeps recurrence grouping and output
  determinism reviewable without inflating the release-gated real corpus.

## Large-Scale Real-World Evaluation

The checked-in regression corpus (211 fixtures) is the deterministic trust gate. A separate large-scale evaluation validates that the bundled playbooks cover the failure patterns that actually appear at production volume.

**Dataset: github-actions-2026-04-29**

30,094 real-world GitHub Actions failure logs collected from public repositories in early 2026, evaluated against the v0.4.4 playbook set.

| Metric | Value |
| --- | --- |
| Total logs | 30,094 |
| Matched | 26,915 |
| Unmatched | 3,179 |
| Coverage | **89.4%** |

### Top 20 Matched Failure Classes

| Rank | Failure ID | Count | % of matched |
| --- | --- | --- | --- |
| 1 | container-crash | 5,597 | 20.8% |
| 2 | eslint-failure | 3,457 | 12.8% |
| 3 | buildkit-session-lost | 1,827 | 6.8% |
| 4 | ignored-exit-code | 1,588 | 5.9% |
| 5 | pnpm-lockfile | 1,279 | 4.8% |
| 6 | testcontainer-startup | 889 | 3.3% |
| 7 | node-version-mismatch | 692 | 2.6% |
| 8 | go-test-failure | 636 | 2.4% |
| 9 | git-shallow-checkout | 617 | 2.3% |
| 10 | runtime-mismatch | 592 | 2.2% |
| 11 | artifact-upload-failure | 575 | 2.1% |
| 12 | cache-miss-non-fatal | 555 | 2.1% |
| 13 | python-module-missing | 547 | 2.0% |
| 14 | gradle-build | 539 | 2.0% |
| 15 | npm-ci-lockfile | 434 | 1.6% |
| 16 | npm-peer-dependency-conflict | 360 | 1.3% |
| 17 | typescript-compile | 270 | 1.0% |
| 18 | cargo-test-failure | 253 | 0.9% |
| 19 | cargo-compile-error | 246 | 0.9% |
| 20 | formatting-failure | 242 | 0.9% |

The top 20 classes account for 74% of all matched cases. The five highest-volume failures (`container-crash`, `eslint-failure`, `buildkit-session-lost`, `ignored-exit-code`, `pnpm-lockfile`) alone account for over 50% of matches.

### Unmatched Logs (10.6%)

The 3,179 unmatched logs represent genuine gaps — failure modes not yet covered by the current playbook set. All are singleton or near-singleton clusters in this dataset, indicating long-tail or highly project-specific patterns. Gap sampling is available under `eval-work/github-actions-2026-04-29-gaps/`.

### Log Chunks Evaluation

The log-chunks dataset has been evaluated separately (87.2% coverage, 102/117 fixtures) but predates several playbook improvements shipped in v0.4.4. A re-evaluation against the current playbook set is pending.

### What This Number Means

The 89.4% figure describes how often a real failing GitHub Actions log matches at least one bundled playbook. It is a coverage claim, not an accuracy claim — a matched log means the relevant playbook fired, not that the ranked output was reviewed by a human. It is separate from the checked-in 228-fixture baseline above. Evaluated on: 2026-04-29.

## Contribution Prompt

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
