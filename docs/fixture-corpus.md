# Fixture Corpus

Faultline's trust boundary is the checked-in corpus, not a vague accuracy claim. The current snapshot below reflects the accepted real fixtures and bundled playbooks in this repository.

## Current Snapshot

- Bundled playbooks: 67
- Accepted real fixtures: 73
- Top-1 match rate: 100% (73/73)
- Top-3 match rate: 100% (73/73)
- Unmatched fixtures: 0
- False positives: 0
- Weak matches: 7

These numbers describe the checked-in regression corpus only. They are useful because they are deterministic, reviewable, and reproducible from the repository state.

## Why It Matters

- The corpus is built from accepted real-world CI failures under `fixtures/real/`.
- Ranking changes are gated by regression checks instead of hidden online learning.
- Playbook coverage, thresholds, and baseline behavior are visible in version control.
- Automation consumers can audit the proof artifact that backs the product's trust story.

## Regenerate And Check

Build the CLI first:

```bash
make build
```

Re-run the real-fixture regression gate:

```bash
./bin/faultline fixtures stats --class real --check-baseline
```

Inspect the same report as JSON:

```bash
./bin/faultline fixtures stats --class real --json --check-baseline
```

Recompute the broader combined corpus if needed:

```bash
./bin/faultline fixtures stats --class all
```

## Source Of Truth

- Bundled playbooks live under `playbooks/bundled/`.
- Accepted real fixtures live under `fixtures/real/`.
- The checked-in regression baseline is `fixtures/real/baseline.json`.
- The fixture commands are wired through `faultline fixtures stats`.
