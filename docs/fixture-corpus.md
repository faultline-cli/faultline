# Fixture Corpus

Faultline's trust boundary is the checked-in corpus, not a vague accuracy claim. The current snapshot below reflects the accepted real fixtures and bundled playbooks in this repository.

## Current Snapshot

- Bundled playbooks: 77
- Accepted real fixtures: 103
- Top-1 match rate: 100% (103/103)
- Top-3 match rate: 100% (103/103)
- Unmatched fixtures: 0
- False positives: 0
- Weak matches: 11 (10.7%)
- Fixture metadata validation: required for real and staging corpora
- Corpus fingerprint drift: release-gated through `fixtures/real/baseline.json`

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
| build | 29 |
| runtime | 13 |
| test | 10 |
| ci | 8 |
| network | 6 |
| deploy | 6 |
| auth | 5 |

Accepted real fixtures mapped through expected playbooks (from `fixtures/real/`):

| Category | Accepted Real Fixtures |
| --- | --- |
| build | 33 |
| network | 28 |
| ci | 19 |
| runtime | 11 |
| auth | 7 |
| deploy | 4 |
| test | 1 |

This table is intended as public proof coverage, not a claim that unknown failures are solved.

## Release Snapshot Trend

Starting snapshot table for release-over-release tracking:

| Snapshot | Bundled Playbooks | Accepted Real Fixtures | Top-1 | Top-3 | Unmatched | False Positive |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-04-17 baseline (`fixtures/real/baseline.json`) | 77 | 103 | 100% | 100% | 0 | 0 |

Append one row per release cut so corpus growth and match stability stay visible over time.

## Contribution Prompt

If Faultline misses a failure class, contribute a sanitized public log:

1. Open an issue with an anonymized failing snippet and environment context.
2. Include a public source URL when possible (issue, discussion, or forum thread).
3. Avoid secrets, private hostnames, and internal repository names.
4. Mark expected diagnosis if known to speed triage.

Maintainers should route accepted cases through the deterministic ingest/review/promote flow.

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
- The checked-in regression baseline is `fixtures/real/baseline.json`.
- The fixture commands are wired through `faultline fixtures stats`.
- Source provenance and adapter counts are included in `faultline fixtures stats` output.
