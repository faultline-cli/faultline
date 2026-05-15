# Local Store Groundwork

Faultline contains a local forensic store implementation.

Important product-boundary note: local single-repo history can remain part of
the product, but aggregation, reporting, and cross-repo recurring-failure
coordination are locked to Faultline Team.

The goal of the local store is narrow and practical: support single-repo
forensic recall, deterministic validation, and maintainers' investigation
without turning Faultline into a service, dashboard, or analytics system.

## Purpose

The local store implementation exists to support:

- recurrence tracking by normalized `signature_hash`
- first-seen and last-seen timestamps
- occurrence counts over time
- deterministic `input_hash` and `output_hash` recording for repeated-run checks
- ranked playbook match history
- local recurring-failure reports grouped by `failure_id`

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

`faultline analyze` writes matched failures to the local store by default. No
configuration is required; if SQLite cannot be opened, Faultline degrades to a
no-op store and still returns the diagnosis.

Local history enrichment in `analyze` output remains explicit opt-in through:

- `--history`
- `--store /path/to/store.db`
- `FAULTLINE_STORE` set to a concrete store path
- the hidden advanced `--metrics-history <jsonl>` flag when only explicit
  reliability metrics are needed

`--no-history` and `--no-store` remain hidden compatibility switches and force
the no-op store. Explicit history commands (`report`, `history`, `signatures`,
and `verify-determinism`) still open the default store path when no `--store`
is supplied, because reading local history is their whole job.

Single-repo companion surfaces can read from the same local store:

- `faultline history`
- `faultline report`
- `faultline history --signature <hash>`
- `faultline signatures`
- `faultline verify-determinism <logfile>`

## What Is Stored

Schema v2 keeps five small tables plus one additive artifact snapshot column on
the main run record:

- `analysis_runs`
- `findings`
- `signatures`
- `playbook_matches`

Stored data is intentionally minimal:

- run timestamp and input source path on `analysis_runs`
- playbook IDs, ranks, scores, and confidence values
- `signature_hash`
- normalized signature material
- `input_hash` and `output_hash`
- first-class `artifact_json` snapshots for replay, compare, and deterministic
  remediation handoff
- first-seen and last-seen times
- small evidence excerpts, retained as the first matched evidence line

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
- default `faultline analyze` writes to local SQLite, but normal output omits
  recurrence counters unless `--history`, `--store`, or `FAULTLINE_STORE` is set
- reliability metrics and advisory policy fields stay absent unless local
  history is explicitly surfaced or an explicit metrics history file provides
  enough input data
- disabling the store or leaving history off does not change match logic
- store failures do not block core analysis by default
- JSON remains stable and additive for automation

When the store is active, Faultline includes:

- top-level `input_hash`
- top-level `output_hash`
- top-level `artifact`
- result-level `signature_hash`
- recurrence fields such as `seen_before`, `occurrence_count`,
  `first_seen_at`, and `last_seen_at`

When enabled, the human-readable `analyze` and `trace` surfaces may also show
a compact history summary when local history exists for the winning result.
That enrichment is explicit and additive: it does not silently rerank the
diagnosis.

That history summary is intentionally concise:

- a short `signature_hash` prefix so users can pivot into `faultline history --signature <hash>`
- a recurrence line such as `seen 3 times over 7d in local history`
- explicit `first seen` and `last seen` timestamps

When reliability metrics are emitted, they are advisory context only. Faultline
does not trigger retries, quarantine suites, route CI jobs, or mutate pipeline
state from local metrics or policy recommendations.

To keep repeated-run verification useful, `output_hash` is computed from the
stable diagnosis payload before local-history counters and policy summaries are
applied. That means repeated runs of the same input can still be compared even
while `occurrence_count`, `first_seen_at`, or history-derived policy context
change over time.

## Reading History

History is intended to answer explicit questions:

- have we seen this exact normalized failure signature before?
- how often has this failure class won recently?
- did the same input produce stable structured output over time?

Interpret the fields narrowly:

- `signature_hash`: recurring normalized instance identity
- `occurrence_count`: number of stored top-ranked findings for that signature
- `first_seen_at` / `last_seen_at`: first and most recent stored occurrence
- `seen_before`: shorthand that the signature already existed before this run

`occurrence_count` is not a hidden severity score, and it does not change the
detector result by itself.

## History Commands

`faultline report` shows the retention-oriented view from the local store:

- grouped `failure_id`
- count of stored top-ranked occurrences
- last seen timestamp
- latest example evidence line

`faultline history` shows three additive views from the local store:

- recurring signatures
- playbook selection frequency, total match frequency, and recurring-run counts

`faultline history --signature <hash>` focuses on one signature and shows:

- recurrence metadata
- recent stored findings for that signature

`faultline signatures` is the narrowest surface: it just lists stored
signatures with counts and timestamps so users can pivot into
`faultline history --signature ...`.

`faultline verify-determinism <logfile>` computes the same canonical
`input_hash` used during analysis and reports whether stored runs of that input
produced one stable `output_hash` or drifted across multiple outputs.

## Maintainer Quality Feedback

The store now exposes lightweight maintainers' summaries rather than a
dashboard product:

- recurring signature counts to spot over-grouping or fragmentation regressions
- playbook `selected_count` and `matched_count` so maintainers can see when a
  rule often appears but rarely wins
- `non_selected_count` and `avg_rank` to highlight noisy runners-up that may
  need better exclusions or narrower wording
- recurring-run counts to show which rules are repeatedly winning on real inputs
- average selected confidence by playbook

These views are intentionally local, inspectable, and bounded. They are meant
to guide single-repo diagnosis and catalog maintenance, not to define the
Team-layer commercial boundary.

## Privacy

The store is local-first and inspectable.

It prefers hashes and normalized excerpts over raw inputs so it can be useful
without turning into a second copy of your CI log archive.

If you need stricter handling, leave history off or force the no-op store with
the hidden compatibility flags `--no-store` or `--no-history`.

## Boundary Note

Going forward:

- Core should not expand its product promise around local persistence
- Team should own cross-repo history correlation, aggregation, reporting, org
  policy, shared playbooks, and sync
- any remaining local-store behavior should stay hidden, companion-only, or
  maintainer-oriented unless the product boundary is explicitly revised

## Deferred Work

The current store is groundwork for future features, but those are not all
shipped now:

- retention policies
- flaky/stability heuristics beyond recurrence groundwork
- richer determinism verification across Faultline versions
- deeper inspect/guard-specific persistence policy
