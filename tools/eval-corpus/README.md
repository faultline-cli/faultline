# faultline-eval

An evaluation harness for measuring how well faultline's playbooks match real
CI failure logs. The tool is deterministic, requires no external services, and
is designed to run inside the same repository as the playbooks it evaluates.

## Overview

`faultline-eval` runs faultline against a corpus of labelled log fixtures and
produces stable, machine-readable coverage metrics. The workflow is:

```
ingest → manifest → run → report → gaps → compare → badge
```

Each step is an independent subcommand. Run them in order during CI or
selectively during playbook authoring.

## Subcommands

### `ingest`

Reads a raw fixture dataset (CSV, JSONL, or TSV) and writes a normalised JSONL
corpus file that all subsequent commands consume.

```sh
faultline-eval ingest \
  --config dataset.yaml \
  --input  failures.csv \
  --out    corpus.jsonl
```

The config file describes which column contains the raw log text and optional
metadata columns (source, tag, severity). See `ingest --help` for schema
details.

### `manifest`

Computes a deterministic version fingerprint for a corpus file. Use this to
detect when the corpus has changed between CI runs.

```sh
faultline-eval manifest \
  --corpus corpus.jsonl \
  --out    corpus.manifest.json \
  --corpus-id ci-failures-v3
```

The manifest records a SHA-256 content hash of the sorted fixture IDs. The same
corpus always produces the same hash regardless of file ordering or the machine
it runs on.

### `run`

Executes faultline against every fixture in the corpus and writes results to a
JSONL file.

```sh
faultline-eval run \
  --corpus  corpus.jsonl \
  --out     results.jsonl
```

Each result records whether faultline matched, which playbook matched (if any),
duration, and the first error-line snippet extracted from the log.

### `report`

Reads a results file and prints a human-readable or machine-readable summary.

```sh
faultline-eval report --results results.jsonl            # text table
faultline-eval report --results results.jsonl --json     # JSON summary
faultline-eval report --results results.jsonl --markdown # Markdown table
```

The report includes overall match-rate, per-playbook hit counts, and a list of
top unmatched tags.

### `gaps`

Clusters unmatched results by normalised error signature and generates playbook
stubs for the most common gaps.

```sh
faultline-eval gaps \
  --results     results.jsonl \
  --fixtures    corpus.jsonl \
  --out         gaps/ \
  --max-samples 5
```

Output directory structure:

```
gaps/
  clusters.json          # machine-readable cluster list
  clusters.md            # human-readable markdown summary
  stubs/
    cluster-001.yaml     # playbook stub for the top gap
    cluster-002.yaml
    ...
  samples/
    cluster-001/         # up to N sample log files per cluster
      sample-01.log
      ...
```

#### How clustering works

1. For each unmatched result, extract `FirstLineSnippet` (the first non-empty
   line of the log).
2. Strip variable tokens: ISO-8601 timestamps, UUIDs, hex IDs longer than 12
   chars, file paths, URLs, and bare numbers.
3. Compute SHA-256 of the cleaned line. The first 8 hex characters form a
   stable bucket key.
4. After grouping, sort buckets by count descending and re-number them
   `cluster-001`, `cluster-002`, and so on.

The clustering is entirely deterministic — the same results file always
produces the same clusters.

#### What the stubs are for

Each stub is a skeleton playbook YAML pre-populated with the suspected failure
class, representative error line, and sample fixture IDs. Copy it into
`playbooks/bundled/`, fill in the `patterns` and `remediation` fields, and run
`make test` to validate it against the corpus samples.

#### Promoting a cluster into a playbook

1. Inspect `gaps/clusters.md` to find the highest-count cluster that is not yet
   covered.
2. Copy `gaps/stubs/cluster-NNN.yaml` to `playbooks/bundled/my-new-failure.yaml`.
3. Review `gaps/samples/cluster-NNN/` to understand the actual log format.
4. Fill in `patterns`, `severity`, and `remediation` in the playbook.
5. Run `faultline-eval run --corpus corpus.jsonl --out results-new.jsonl`.
6. Run `faultline-eval compare --baseline results.jsonl --current results-new.jsonl`.
7. If the compare gate passes and coverage increased, commit the new playbook.

### `compare`

Compares two results files (baseline vs. current) and exits non-zero if any
configured gate condition fails. Use this as a CI regression gate.

```sh
faultline-eval compare \
  --baseline  results-main.jsonl \
  --current   results-pr.jsonl \
  --out        comparison \
  --min-match-rate 0.70 \
  --fail-on-coverage-drop \
  --fail-on-new-nondeterminism
```

Output files written to `--out`:

- `comparison.json` — machine-readable delta
- `comparison.md` — markdown report suitable for a PR comment

#### Gate conditions

| Flag | Behaviour |
|------|-----------|
| `--min-match-rate F` | Fail if current match-rate < F |
| `--fail-on-coverage-drop` | Fail if current match-rate < baseline match-rate |
| `--fail-on-new-nondeterminism` | Fail if any fixture produces inconsistent results that did not in baseline |

All gate failures are enumerated in `FailReasons`; the command reports all of
them at once rather than stopping at the first.

#### Example GitHub Actions step

```yaml
- name: Evaluate playbook coverage
  run: |
    faultline-eval run --corpus corpus.jsonl --out results-current.jsonl
    faultline-eval compare \
      --baseline  results-main.jsonl \
      --current   results-current.jsonl \
      --fail-on-coverage-drop \
      --min-match-rate 0.65
```

Store `results-main.jsonl` as a build artifact on the default branch and
download it at the start of each PR run.

### `badge`

Writes a compact coverage summary artifact (JSON + Markdown) suitable for
embedding in a README or attaching to a release.

```sh
faultline-eval badge \
  --results        results.jsonl \
  --out            coverage-summary \
  --corpus-version ci-v3 \
  --corpus-hash    $(jq -r .overall_corpus_hash corpus.manifest.json) \
  --deterministic  pass \
  --top            5
```

Output files:

- `coverage-summary.json` — machine-readable badge data
- `coverage-summary.md` — one-line markdown badge + summary table

## Adding new datasets

Implement the `RecordReader` interface:

```go
type RecordReader interface {
    Read() (Record, error)
    Close() error
}
```

Then add a `case` for the new format in `ingest/reader.go:NewReader()`. The
interface is the only extension point — no other file changes are needed.

## Why coverage ≠ accuracy

Match-rate measures _coverage_: what fraction of CI failures does faultline
recognise at all? It does **not** measure _accuracy_: whether the matched
playbook is the most appropriate one for a given failure.

Accuracy requires labelled fixtures — fixtures that carry an `ExpectedFailureID`
field identifying the canonical playbook. When a fixture has an
`ExpectedFailureID`, `faultline-eval` can classify the match as:

| Outcome | Meaning |
|---------|---------|
| `TP` (true positive) | Matched, and the matched playbook is the expected one |
| `FP` (false positive) | Matched, but a different playbook matched instead |
| `FN` (false negative) | Not matched, but a match was expected |
| `unlabelled` | No `ExpectedFailureID` — contributes to coverage only |

Unlabelled fixtures are counted in coverage but never in accuracy calculations.
This prevents artificially inflating precision/recall with data that has not
been reviewed by a human.

To label a fixture, add `expected_failure_id` (and optionally
`expected_failure_class`, `expected_severity`, `expected_remediation_category`)
to the fixture's JSONL record. Run `faultline-eval run` again to pick up the
labels.

## Known limitations

- Playbooks are re-loaded from disk for each `AnalyzeReader` call inside the
  runner. For large corpora this is acceptably fast, but a caching layer would
  improve throughput when running the same corpus many times.
- The clustering heuristic operates on `FirstLineSnippet` only. Multi-line
  errors where the first line is a generic wrapper (e.g., `Error:` with the
  detail on the second line) may undercluster. Add a second-line fallback if
  this becomes a problem in practice.
- `compare --fail-on-new-nondeterminism` requires two separate `run` passes
  over the corpus. The determinism check compares the two result files; it does
  not re-execute faultline internally.
