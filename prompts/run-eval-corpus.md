# Run Eval Corpus Pipeline

Evaluate faultline's playbook coverage against a labelled log dataset using the
`faultline-eval` harness in `tools/eval-corpus/`.

## Prerequisites

- `bin/faultline-eval` is built (`make eval-build` or `go build -o bin/faultline-eval ./tools/eval-corpus`)
- A dataset config YAML exists in `fixtures/datasets/` (see `fixtures/datasets/travis-torrent-2017.yaml` for the bundled Travis Torrent dataset)
- `eval-work/` directory is created for output artefacts

## Step 1 — Pre-filter the raw dataset (large CSVs only)

For the Travis Torrent dataset (`final-2017-01-25.csv`, 3.4 GB), extract only
the rows that are relevant before ingesting:

```sh
mkdir -p eval-work
# Keep header + failed builds that have test failure content
head -1 fixtures/datasets/final-2017-01-25.csv > eval-work/failed.csv
# Column 65 = tr_status, column 60 = tr_log_tests_failed
awk -F',' 'NR>1 && $65 ~ /failed/ && $60 !~ /^""?$/ { print }' \
  fixtures/datasets/final-2017-01-25.csv >> eval-work/failed.csv
printf "Filtered rows: %s\n" "$(wc -l < eval-work/failed.csv)"
```

For other datasets, skip this step and point the config directly at the source file.

## Step 2 — Ingest into a normalised corpus

```sh
bin/faultline-eval ingest \
  --config fixtures/datasets/travis-torrent-2017.yaml \
  --input  eval-work/failed.csv \
  --out    eval-work/corpus.jsonl
printf "Corpus fixtures: %s\n" "$(wc -l < eval-work/corpus.jsonl)"
```

Review a sample to sanity-check the extracted log content:

```sh
head -3 eval-work/corpus.jsonl | python3 -m json.tool | head -40
```

## Step 3 — Compute corpus manifest

```sh
bin/faultline-eval manifest \
  --corpus    eval-work/corpus.jsonl \
  --out       eval-work/corpus.manifest.json \
  --corpus-id travis-torrent-2017
cat eval-work/corpus.manifest.json
```

## Step 4 — Run faultline against the corpus

```sh
bin/faultline-eval run \
  --corpus  eval-work/corpus.jsonl \
  --out     eval-work/results.jsonl \
  --workers 8
printf "Evaluated fixtures: %s\n" "$(wc -l < eval-work/results.jsonl)"
```

Use more workers for faster runs on large corpora (`--workers 16`). The run is
CPU-bound and embarrassingly parallel.

## Step 5 — Generate coverage report

```sh
# Human-readable table
bin/faultline-eval report --results eval-work/results.jsonl

# JSON summary (for further processing)
bin/faultline-eval report --results eval-work/results.jsonl --json \
  > eval-work/coverage.json

# Markdown (for PR comments)
bin/faultline-eval report --results eval-work/results.jsonl --markdown \
  > eval-work/coverage.md
```

Key metrics to note:
- **Overall match rate**: percentage of fixtures matched by at least one playbook
- **Per-playbook hit count**: which playbooks are doing the most work
- **Top unmatched tags**: most common failure patterns with no playbook

## Step 6 — Identify coverage gaps

```sh
bin/faultline-eval gaps \
  --results     eval-work/results.jsonl \
  --fixtures    eval-work/corpus.jsonl \
  --out         eval-work/gaps \
  --max-samples 5
```

Review output:

```sh
cat eval-work/gaps/cluster-summary.md
ls eval-work/gaps/playbook-stubs/
```

## Step 7 — Compare against a baseline (CI regression gate)

If a baseline results file exists (e.g. stored as a CI artifact):

```sh
bin/faultline-eval compare \
  --baseline  <baseline-results.jsonl> \
  --current   eval-work/results.jsonl \
  --out       eval-work/comparison \
  --fail-on-coverage-drop \
  --min-match-rate 0.65
cat eval-work/comparison.md
```

## Step 8 — Generate badge artefact

```sh
CORPUS_HASH=$(python3 -c "import json; print(json.load(open('eval-work/corpus.manifest.json'))['overall_corpus_hash'])")
bin/faultline-eval badge \
  --results        eval-work/results.jsonl \
  --out            eval-work/coverage-summary \
  --corpus-version travis-torrent-2017 \
  --corpus-hash    "$CORPUS_HASH"
cat eval-work/coverage-summary.md
```

## Acting on Results

After reviewing the gap report:

1. Identify the cluster with the highest count that has no matching playbook.
2. Copy `eval-work/gaps/playbook-stubs/cluster-NNN.yaml` to `playbooks/bundled/`.
3. Read sample logs in `eval-work/gaps/samples/cluster-NNN/`.
4. Fill in `patterns`, `severity`, and `remediation`.
5. Re-run steps 4–6 and verify coverage improved.
6. Run `make test` to validate no regressions.
7. Commit the new playbook and re-baseline if the compare gate passes.

## Notes on the Travis Torrent Dataset

`final-2017-01-25.csv` (`fixtures/datasets/`) is the Travis Torrent research
dataset. It contains CI build metadata extracted from Travis CI logs, not raw
log output. The `tr_log_tests_failed` column holds parsed test failure
identifiers (e.g., `features/filtering/if_and_unless.feature`). Expect a low
initial match rate — the gap report is the primary output of interest.

Compatible ingest config: `fixtures/datasets/travis-torrent-2017.yaml`.
