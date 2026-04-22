# Local Store

Faultline now has an optional local forensic store.

The goal is narrow and practical: keep enough durable local memory to answer
questions like "have we seen this failure signature before?" without changing
Faultline into a service, dashboard, or analytics system.

## Purpose

The store exists to support:

- recurrence tracking by normalized `signature_hash`
- first-seen and last-seen timestamps
- occurrence counts over time
- deterministic `input_hash` and `output_hash` recording for repeated-run checks
- ranked playbook match history
- hook execution history when hooks are enabled

The store is not:

- a remote backend
- a daemon
- a telemetry pipeline
- a raw-log archive by default

## Backend

- primary backend: SQLite
- runtime mode: local only
- default path: `~/.faultline/store.db`
- graceful fallback: when disabled, unavailable, or corrupt, Faultline drops to
  a no-op store and continues analysis

Advanced CLI configuration is available through the hidden store controls:

- `--store auto`
- `--store off`
- `--store /path/to/store.db`
- `--no-store`
- `FAULTLINE_STORE`

`--no-history` remains a compatibility switch and also disables the store.

## What Is Stored

Schema v1 keeps five small tables:

- `analysis_runs`
- `findings`
- `signatures`
- `playbook_matches`
- `hook_results`

Stored data is intentionally minimal:

- playbook IDs, ranks, scores, and confidence values
- `signature_hash`
- normalized signature material
- `input_hash` and `output_hash`
- first-seen and last-seen times
- small evidence excerpts
- structured hook facts and hook evidence excerpts when hooks run

## What Is Not Stored By Default

- full raw logs
- arbitrary repository snapshots
- secrets on purpose
- telemetry or outbound sync state

## Signature Hashing

Recurrence is keyed by `signature_hash`, not by playbook ID alone.

Faultline computes that hash from:

- a versioned canonical signature payload
- the matched top-level failure ID
- normalized matched evidence lines
- structured trigger attributes when detectors already expose them

Normalization rules intentionally remove unstable noise before hashing:

- strip timestamps
- collapse whitespace
- normalize CI workspace, runner tool, temp, home, and absolute paths
- normalize file line and column numbers
- replace UUID-like tokens
- replace long hex identifiers and long numeric IDs
- replace dynamic IP addresses

The hash is deterministic SHA-256 over the canonical JSON payload.

The advanced `trace` surface exposes the computed signature and payload for
inspection. See [Signature Hashing](./signatures.md) for the full design and
normalization contract.

This makes the key stable across common CI noise while keeping unrelated
failures distinct enough for practical local recurrence tracking.

## Determinism Guarantees

The store is additive to the analysis path:

- the engine still analyzes input without reading SQL
- disabling the store does not change match logic
- store failures do not block core analysis by default
- JSON remains stable and additive for automation

When the store is active, Faultline includes:

- top-level `input_hash`
- top-level `output_hash`
- result-level `signature_hash`
- recurrence fields such as `seen_before`, `occurrence_count`,
  `first_seen_at`, and `last_seen_at`

## Privacy

The store is local-first and inspectable.

It prefers hashes and normalized excerpts over raw inputs so it can be useful
without turning into a second copy of your CI log archive.

If you need stricter handling, disable it with `--no-store` or `--no-history`.

## Deferred Work

The current store is groundwork for future features, but those are not all
shipped now:

- retention policies
- explicit history and batch commands
- flaky/stability heuristics beyond recurrence groundwork
- richer determinism verification across Faultline versions
- deeper inspect/guard-specific persistence policy
