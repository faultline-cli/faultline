---
name: eval-corpus
description: Use this skill when working with the faultline-eval evaluation harness: ingesting a labelled dataset into a normalised corpus, running coverage measurement against faultline playbooks, producing a coverage report, identifying high-value gap clusters, comparing baseline versus current results in CI, and generating playbook stubs from gap data. Trigger it for requests about running the eval pipeline, measuring playbook coverage, finding gaps, or promoting eval results as baselines.
---

# Eval Corpus

This skill is for running `faultline-eval`, the deterministic log evaluation
harness found in `tools/eval-corpus/`. Use it whenever the task involves:

- Ingesting a raw dataset (CSV/JSONL) into a normalised fixture corpus
- Running faultline against a corpus to measure playbook coverage
- Generating coverage reports (text, JSON, Markdown)
- Identifying unmatched failure clusters and generating playbook stubs
- Comparing baseline vs. current results as a CI regression gate
- Generating badge artifacts for README embedding

Do not use it for general playbook authoring, fixture ingestion via
`faultline fixtures ingest`, or the internal real-fixture regression gate
(`make fixture-check`). Those workflows use different tooling.

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`tools/eval-corpus/README.md`](../../../tools/eval-corpus/README.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- The dataset config YAML in `fixtures/datasets/` that you're working with

## Pipeline Overview

```
ingest → manifest → run → report → gaps → compare → badge
```

Each subcommand is independent. Run them in order for a full evaluation.

## Build

`faultline-eval` is a separate binary from the main `faultline` CLI. Build it once
before running any subcommand:

```sh
go build -o bin/faultline-eval ./tools/eval-corpus
```

Or use the Makefile target:

```sh
make eval-build
```

## Subcommand Reference

### ingest

Reads a raw CSV (or JSONL) dataset and writes a normalised JSONL corpus.

```sh
bin/faultline-eval ingest \
  --config fixtures/datasets/<config>.yaml \
  --out    eval-work/corpus.jsonl
```

**Required config fields:**

```yaml
input:
  type: csv
  path: fixtures/datasets/<file>.csv

parsing:
  log_field: <column-name>   # column containing the log/error text
  id_field: <column-name>    # stable row identifier (optional)
  timestamp_field: <column>  # event timestamp (optional)

processing:
  dedupe: true
  max_log_size: 200kb
```

### run

Runs faultline detection on every fixture and writes results.

```sh
bin/faultline-eval run \
  --corpus  eval-work/corpus.jsonl \
  --out     eval-work/results.jsonl \
  --workers 8
```

### report

Summarises coverage from a results file.

```sh
bin/faultline-eval report --results eval-work/results.jsonl
bin/faultline-eval report --results eval-work/results.jsonl --markdown
bin/faultline-eval report --results eval-work/results.jsonl --json
```

Key metrics: overall match rate, per-playbook hit counts, top unmatched tags.

### gaps

Clusters unmatched fixtures and generates playbook stubs.

```sh
bin/faultline-eval gaps \
  --results     eval-work/results.jsonl \
  --fixtures    eval-work/corpus.jsonl \
  --out         eval-work/gaps \
  --max-samples 5
```

Output:
- `eval-work/gaps/clusters.jsonl` — machine-readable cluster list
- `eval-work/gaps/cluster-summary.md` — human-readable summary
- `eval-work/gaps/samples/<cluster>/` — representative log files
- `eval-work/gaps/playbook-stubs/` — YAML stubs ready for review

### compare

CI regression gate: fail if coverage drops below a threshold.

```sh
bin/faultline-eval compare \
  --baseline baseline-results.jsonl \
  --current  eval-work/results.jsonl \
  --out      eval-work/comparison \
  --fail-on-coverage-drop \
  --min-match-rate 0.65
```

### badge

Writes a compact coverage summary for README embedding.

```sh
bin/faultline-eval badge \
  --results        eval-work/results.jsonl \
  --out            eval-work/coverage-summary \
  --corpus-version ci-v1
```

## Working With Large Datasets

The Travis Torrent dataset (`fixtures/datasets/final-2017-01-25.csv`) is 3.4 GB
with ~3.9 M rows. Pre-filter it before ingesting to keep run times reasonable:

```sh
# Extract header + failed rows only
head -1 fixtures/datasets/final-2017-01-25.csv > eval-work/failed.csv
awk -F',' 'NR>1 && $65=="\"failed\"" && $60!="\"\"" { print }' \
  fixtures/datasets/final-2017-01-25.csv >> eval-work/failed.csv
```

Column 65 is `tr_status`, column 60 is `tr_log_tests_failed`. Adjust if needed.

Use `fixtures/datasets/travis-torrent-2017.yaml` as the ingest config.

## Promoting a Gap Cluster Into a Playbook

1. Review `eval-work/gaps/cluster-summary.md` for the top unmatched cluster.
2. Copy the corresponding stub from `eval-work/gaps/playbook-stubs/` to `playbooks/bundled/`.
3. Inspect sample log files in `eval-work/gaps/samples/<cluster>/`.
4. Fill in `patterns`, `severity`, and `remediation`.
5. Re-run `faultline-eval run` and `compare` to verify improvement.
6. Commit the new playbook when `make test` passes.

## Guardrails

- Always use `dedupe: true` to avoid inflating coverage metrics with duplicates.
- Set `max_log_size` to prevent extremely large log fields from skewing results.
- Use `--fail-on-coverage-drop` in CI to catch regressions early.
- Store the baseline `results.jsonl` as a CI artifact and download it at PR time.
- Never commit `eval-work/` output — it is transient.
