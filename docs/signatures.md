# Signature Hashing

Faultline uses `signature_hash` as a deterministic recurrence key.

Under the locked product boundary, cross-repo recurrence and team-level
correlation belong to Faultline Team. This document describes the hashing
contract itself, which remains useful for single-repo history, Team
aggregation, and deterministic validation.

The goal is narrow and deterministic:

- same logical failure should produce the same `signature_hash`
- equivalent failures across noisy CI environments should normalize the same way
- distinct failures should remain distinct often enough to trust recurrence
- the normalization rules must stay inspectable and explainable

This is not a fuzzy-similarity system. It is a stable recurrence key.

## Audit Summary

Before `signature.v1`, the repository already had useful groundwork:

- `failure_id` already existed as the diagnosed playbook ID
- `internal/matcher` already extracted deterministic evidence from matched lines
- `internal/output` already exposed `signature_hash`, `input_hash`, and `output_hash`
- `internal/store` already persisted `signature_hash`, normalized signature text,
  recurrence timestamps, and occurrence counts
- `internal/engine/fingerprint.go` already produced a short backward-compatible
  `fingerprint` from top evidence

The gaps were mostly around role clarity and implementation boundaries:

- signature generation lived in `internal/store`, even though it is analysis
  domain logic rather than persistence logic
- the normalized payload was ad hoc newline text instead of a canonical,
  versioned structure
- the signature input used only top-level evidence strings and ignored existing
  structured trigger attributes such as `signal_id`, `file`, and `scope_name`
- normalization coverage was useful but under-documented and only lightly tested
- trace output did not surface the canonical payload used for hashing

Risky refactor areas were:

- `internal/app/store_support.go`, which injects recurrence fields into results
- `internal/store/sqlite.go`, which persists `signature_hash` and normalized
  signature material
- `internal/output/output_json.go`, where public automation-facing fields must
  remain stable
- `internal/trace` and `internal/output/output_trace.go`, where inspectability
  can improve without making default `analyze` noisy

## Field Roles

These hashes have different jobs and should not be conflated:

- `failure_id`: the diagnosed failure class or playbook ID
- `signature_hash`: the normalized recurring instance identity
- `input_hash`: the deterministic hash of the analyzed input log
- `output_hash`: the deterministic hash of the structured analysis output
- `fingerprint`: a short backward-compatible summary token for top-result output

`signature_hash` is suitable for:

- recurrence counts
- first-seen and last-seen tracking
- grouping related failures across repeated runs
- future flaky/stability groundwork
- determinism investigations alongside `input_hash` and `output_hash`

`signature_hash` is not:

- a replacement for `failure_id`
- a fuzzy search key
- a hidden classifier

## Signature Input

Faultline now builds a canonical payload in `internal/signature`:

```json
{
  "version": "signature.v1",
  "failure_id": "missing-executable",
  "detector": "log",
  "evidence": [
    "<workspace>/.github/workflows/ci.yml:<n> exec <runner>/node20/bin/node no such file or directory"
  ],
  "attributes": {
    "files": ["internal/api/handler.go"],
    "scope_names": ["serveuser"],
    "signal_ids": ["panic.handler"]
  }
}
```

Composition rules are explicit:

- `version` is always included
- `failure_id` is always included
- `detector` is included when present
- `evidence` is a sorted, deduplicated list of normalized evidence lines
- `attributes` is included only when trigger evidence exposes stable structured
  fields

Current structured attributes come only from trigger evidence:

- `signal_ids`
- `files`
- `scope_names`

That keeps signatures tied to core positive evidence instead of optional
mitigations, suppressions, or playbook prose.

## Normalization Pipeline

Normalization is deterministic and stage-ordered:

1. Strip ANSI escape sequences.
2. Normalize line endings and trim blank fragments.
3. Replace timestamps and dates with placeholders:
   - `<timestamp>`
   - `<date>`
   - `<time>`
4. Normalize path tokens:
   - temp paths become `<tmp>/...`
   - CI workspace paths become `<workspace>/...`
   - runner-managed tool paths become `<runner>/...`
   - home-directory paths become `<home>/...`
   - other absolute paths become `<path>/...`
5. Replace line and column suffixes with `:<n>`.
6. Replace UUID-like values with `<id>`.
7. Replace long SHA or hex-like identifiers with `<hex>`.
8. Replace IPv4 addresses with `<ip>`.
9. Replace large unstable numeric identifiers with `<n>`.
10. Collapse internal whitespace and lowercase the final line.

The rules intentionally preserve semantically meaningful content such as:

- tool and executable names
- package and dependency names
- failure phrases and error text
- exit codes
- relative repository paths in structured trigger attributes
- detector-specific trigger IDs

### Before / After Examples

```text
2026-04-22T12:05:31Z /home/runner/work/app/app/.github/workflows/ci.yml:118:
exec /__e/node20/bin/node: no such file or directory
```

becomes:

```text
<timestamp> <workspace>/.github/workflows/ci.yml:<n> exec <runner>/node20/bin/node no such file or directory
```

```text
/tmp/build-9812/output.log:43: request 9fd46ec5-6c4f-4f0c-a11d-f3c96f172d63 failed for commit 71E944493FA59840 on 10.24.6.9
```

becomes:

```text
<tmp>/output.log:<n> request <id> failed for commit <hex> on <ip>
```

## Hashing

Faultline hashes the canonical payload with:

- algorithm: SHA-256
- encoding: lowercase hex
- truncation: none

The hash input is the canonical JSON payload produced by the versioned
signature builder. The JSON is emitted without HTML escaping so the stored
payload stays human-readable and stable.

## Layer Integration

- CLI layer:
  `analyze` remains quiet by default; the advanced `trace` surface now exposes
  the signature hash and canonical payload for inspection.
- App layer:
  signature computation still happens before persistence and result rendering so
  recurrence fields are available to store and output layers.
- Engine layer:
  the engine still owns log analysis and `fingerprint`, but signature
  generation now lives beside the domain model in `internal/signature` instead
  of in the store.
- Detectors layer:
  detectors do not need custom hashing hooks; they can improve signatures by
  exposing better trigger evidence through existing structured fields.
- Playbooks layer:
  playbook prose does not contribute to signatures.
- Output layer:
  `signature_hash` remains in stable analysis JSON; trace output now shows the
  canonical payload used to compute it.
- Store layer:
  recurrence still keys off `signature_hash`; `normalized_signature` now stores
  the canonical versioned payload.

## Versioning

Normalization behavior is versioned in-band through the canonical payload:

- current version: `signature.v1`
- stored inside the persisted `normalized_signature` JSON
- included in the hash input itself

That means future signature versions can coexist in history without ambiguous
interpretation. If Faultline changes the signature rules later, the new version
will naturally produce a different canonical payload and therefore a different
hash.

No store-specific branching is required to interpret historical rows: inspect
the stored payload version.

## Tests

The signature suite covers:

- unit tests for timestamps, UUIDs, SHAs, IPs, and path normalization
- path normalization across Unix and Windows-style paths
- multiline evidence handling
- canonical payload snapshot coverage
- stability across equivalent noisy variants
- fixture-driven variant matrices across CI-style path, timestamp, and runner
  noise
- store-backed end-to-end recurrence checks that run noisy variants through
  `analyze`, persistence, `history`, and `verify-determinism`
- distinctness checks for meaningfully different causes that still map to the
  same playbook family
- distinctness for unrelated failures
- fixture-driven end-to-end checks using real CI logs
- trace output assertions for signature inspectability

The dedicated noisy-variant matrix now covers cases such as:

- same missing-executable failure across Linux, Windows, and hosted-toolcache
  path variants
- same `npm ci` lockfile mismatch across runner and workspace differences
- same Node.js version mismatch across workspace, temp-path, and toolcache
  wrapper noise
- same missing environment variable across GitHub Actions and Windows runner
  path variants
- different missing environment variables that stay distinct even though they
  map to the same `env-var-missing` playbook
- same dependency-drift playbook with different conflicting packages, which
  should stay distinct instead of collapsing into one signature

## Known Limitations

- signatures are only as good as the evidence extracted for the winning result
- detector families with richer structured trigger evidence will produce better
  signatures than detectors that only expose free-form text
- the current path strategy preserves only the most meaningful tail segments of
  absolute paths; this is deliberate, but it can still merge some edge cases
- `signature.v1` does not attempt stack-specific parsing of test names, package
  coordinates, or exit semantics beyond what is already present in extracted
  evidence

## Deferred Items

- optional stack-specific normalization extensions, gated by explicit versioning
- richer structured trigger fields for detectors that can expose package names,
  test identifiers, or executable identities safely
- store queries that aggregate recurrence by signature version
- optional explicit debugging surfaces beyond `trace` if operators want
  signature-only inspection without the full rule trace
