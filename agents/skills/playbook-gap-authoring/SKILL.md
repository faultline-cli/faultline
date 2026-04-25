---
name: playbook-gap-authoring
description: Use this skill when the task is a coverage improvement sprint driven by the eval corpus gap report — starting from an existing results.jsonl, identifying the highest-value unmatched clusters, deciding whether each gap warrants a new playbook or an extension to an existing one, authoring or extending the playbook, verifying the change against samples, and re-running the eval to measure improvement. Trigger it for requests like "close coverage gaps", "extend coverage from the eval corpus", "author playbooks from the gap report", or "improve the match rate". Do not trigger it for general playbook authoring from fixtures, for the ingestion pipeline, or for CI baseline comparisons — use the eval-corpus, new-playbook-authoring, or baseline-regression skills for those.
---

# Playbook Gap Authoring

This skill covers the full cycle of closing coverage gaps identified by the
`faultline-eval` gap report: reading the cluster summary, triaging root causes,
deciding extend vs. new, authoring pattern changes, verifying against samples,
and measuring the impact with a re-eval.

Use it when:

- You have a `results.jsonl` from `faultline-eval run` and want to close gaps
- You want to work through the top-N gap clusters systematically
- You need to decide whether a gap merits extending an existing playbook or
  authoring a new one
- You want to measure the coverage delta after pattern changes

Do not use it for:

- Initial corpus ingestion or eval pipeline setup (use `eval-corpus`)
- Detailed fixture authoring (use `fixture-generation`)
- Full new-playbook lifecycle from scratch (use `new-playbook-authoring`)
- CI baseline comparison (use `baseline-regression`)

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- [`eval-work/gaps/cluster-summary.md`](../../../eval-work/gaps/cluster-summary.md)
  (or the specific gaps folder: `gaps2/`, `gaps3/`, etc.)
- The playbook YAML files for the gap candidates identified in triage

## Workflow

### Phase 1 — Establish Baseline

1. Run the report to get current coverage:
   ```sh
   ./bin/faultline-eval report --results eval-work/results.jsonl
   ```
   Record the totals: matched, unmatched, match rate.

2. Regenerate the gap clusters from the **current** results (always regenerate;
   stale gap directories from previous sessions can mislead triage):
   ```sh
   ./bin/faultline-eval gaps \
     --results   eval-work/results.jsonl \
     --fixtures  eval-work/corpus.jsonl \
     --out       eval-work/gaps \
     --max-samples 5
   ```
   > The `faultline-eval run` command buffers **all** results in memory, sorts
   > by FixtureID for determinism, then writes the file in a single pass. The
   > output file stays at 0 bytes until the run completes. Plan accordingly for
   > large corpora (46 k fixtures takes ~10 min at --workers 4).

3. Read `eval-work/gaps/cluster-summary.md`. Note the cluster sizes. Focus on
   the top 5–10 by fixture count; the gains diminish quickly below that.

### Phase 2 — Triage Each Cluster

For each candidate cluster (top-N by fixture count):

4. Inspect the sample log files:
   ```sh
   cat eval-work/gaps/samples/cluster-NNN/sample-001.log
   cat eval-work/gaps/samples/cluster-NNN/sample-002.log
   ```

5. Run the current playbooks against a sample:
   ```sh
   cat eval-work/gaps/samples/cluster-NNN/sample-001.log | ./bin/faultline analyze
   ```

6. Classify the gap (choose one):
   - **Extend**: The root cause maps to an existing playbook but the current
     patterns miss this format (e.g. a condensed vs. verbose test output format).
     The fix is new `match.any` entries and/or `match.none` exclusions.
   - **New**: The root cause is a distinct failure class with no existing playbook
     analog. Follow the `new-playbook-authoring` skill for this path.
   - **Skip**: Too few samples or too noisy/env-specific to generalise. Document
     and move to the next cluster.

   > For the Travis Torrent corpus specifically: the `tr_log_tests_failed` field
   > stores test failures in a condensed format — space/`#`-separated package
   > paths and timing tokens — not the verbose test runner output. A gap that
   > looks like "Go test failures" may just be missing patterns for the condensed
   > format (e.g. `github.com/`, `Test (`, `#coordinator`). Always read 3–5
   > samples before deciding New vs. Extend.

### Phase 3 — Extend or Author

**For Extend (editing an existing playbook):**

7. Open the playbook YAML:
   ```sh
   cat playbooks/bundled/log/<category>/<playbook-id>.yaml
   ```

8. Add new `match.any` patterns grounded in real sample lines.

9. Add `match.none` exclusions for any false-positive risk introduced by broad
   new patterns (e.g. adding `"github.com/"` to a Go test playbook requires
   excluding Python-adjacent lines like `"yt-dl.org"`, `"yt-dlp"`, and
   `"please report this issue on https://github.com"`).

10. Verify that each added pattern normalises correctly. Faultline normalises via
    `strings.ToLower` + `strings.TrimSpace` + whitespace collapse. Test with:
    ```sh
    cat eval-work/gaps/samples/cluster-NNN/sample-001.log | ./bin/faultline analyze
    ```
    Expected: the playbook now matches at some confidence level.

11. Spot-check false-positive risk on a dissimilar cluster sample:
    ```sh
    cat eval-work/gaps/samples/cluster-MMM/sample-001.log | ./bin/faultline analyze
    ```
    Ensure the playbook does not fire on unrelated failure categories.

**For New (authoring a new playbook):**

12. Follow the `new-playbook-authoring` skill from step 1. Establish root-cause
    justification from the gap cluster samples before writing YAML.

### Phase 4 — Sample Verification Loop

13. Batch-verify all samples for the cluster:
    ```sh
    for f in eval-work/gaps/samples/cluster-NNN/*.log; do
      echo -n "$f: "
      cat "$f" | ./bin/faultline analyze 2>/dev/null | grep -oP '(?<=playbook=)\S+|No known' | head -1
    done
    ```
    A good result: most samples now match the target playbook.

### Phase 5 — Full Eval and Delta Measurement

14. Run a full eval with the updated playbooks. Use `--workers` to parallelise:
    ```sh
    ./bin/faultline-eval run \
      --corpus  eval-work/corpus.jsonl \
      --out     eval-work/results-new.jsonl \
      --workers 8
    ```
    > Playbooks are loaded from the filesystem at runtime (not compiled in).
    > No rebuild is needed between pattern edits and eval runs.

15. Report the new coverage:
    ```sh
    ./bin/faultline-eval report --results eval-work/results-new.jsonl
    ```
    Compare to the baseline from step 1. Record the delta (matched count, rate).

16. Regenerate gaps from the new results to see remaining work:
    ```sh
    ./bin/faultline-eval gaps \
      --results   eval-work/results-new.jsonl \
      --fixtures  eval-work/corpus.jsonl \
      --out       eval-work/gaps-new \
      --max-samples 5
    ```

17. Commit when `make test` passes and the delta is confirmed positive.

## Guardrails

- Always regenerate gap clusters from current results before triaging — stale gap
  directories from previous sessions lead to wasted work on already-resolved gaps.
- Every `match.any` phrase must be grounded in a real sample line. Do not add
  patterns speculatively.
- Every `match.none` exclusion must be grounded in a real confusable case you can
  demonstrate. Document the rationale in a comment if the connection is non-obvious.
- Broad patterns (short phrases, common tokens) need `match.none` guards. Verify
  false-positive risk before running the full eval.
- The `faultline-eval run` file is written atomically at the end of the run; the
  output file stays at 0 bytes until all fixtures are processed.
- Do not commit `eval-work/` output (results JSONL, gap clusters, comparison
  artifacts). These are transient. Only commit playbook YAML changes.
- Run `make test` before declaring the sprint complete.
- When the top-N cluster gains are small (< 22 fixtures per top cluster), the
  remaining unmatched population is long-tail noise. Assess whether further
  coverage work has diminishing returns.

## Deliverable

Report:

- The coverage baseline (matched/unmatched/rate before)
- The improvements made: which playbooks were extended or created, and what
  patterns were added/excluded
- The coverage after the sprint (matched/unmatched/rate after, delta)
- The top remaining gap cluster sizes (to calibrate whether further work is
  warranted)
- `make test` result
