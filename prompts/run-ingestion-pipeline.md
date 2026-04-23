# Run Ingestion Pipeline

Use this workflow when the task is to bring new public failure evidence into Faultline through the repository's fixture intake path.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- [`docs/agent-workflows.md`](../docs/agent-workflows.md)
- [`fixtures/staging/README.md`](../fixtures/staging/README.md)

## Goal

Run the full deterministic ingestion pipeline from public URLs to one of two outcomes:

- rejected from staging
- promoted into `fixtures/real/` with explicit expectations

Prefer a varied batch over a single-source sweep.

## Required Approach

1. Confirm each source is public and suitable for ingestion.
2. Check the current corpus mix before choosing URLs:
   - `./bin/faultline fixtures stats --class real --json`
   - bias the run toward underrepresented adapters or failure classes when the current corpus is skewed
3. Build a mixed candidate batch across multiple adapters when possible:
   - `github-issue`
   - `gitlab-issue`
   - `stackexchange-question`
   - `discourse-topic`
   - `reddit-post`
4. Prefer breadth over depth:
   - take one or two strong candidates from a source before harvesting more from the same thread
   - avoid spending the whole run on one issue, one subreddit post, or one discussion thread unless it clearly yields distinct failure signatures
5. Run `faultline fixtures ingest --adapter ... --url ...` for each candidate URL.
6. Review staged results with `faultline fixtures review`.
7. Reject duplicates, near-duplicates without new signal, setup-only snippets, workaround-only snippets, and cases that still need sanitization.
8. For accepted cases, promote with `faultline fixtures promote <staging-id> --expected-playbook <id>` plus any needed `--strict-top-1`, `--disallow`, or `--expected-stage` guards.
9. Rebuild and run `./bin/faultline fixtures stats --class real --check-baseline`.
10. If a promoted fixture exposes weak matching or a confusable neighbor, continue with [`refine-playbook-from-fixture.md`](./refine-playbook-from-fixture.md) before stopping.
11. If the investigation surfaces a repository-local risk that belongs to a bundled `source` playbook rather than log matching, add or update repository fixtures under `internal/engine/testdata/source/` and extend the source-playbook regression tests in the same pass instead of forcing that signal into the real-log corpus.

## Command Skeleton

```bash
faultline fixtures ingest --adapter github-issue --url <public-url>
faultline fixtures ingest --adapter stackexchange-question --url <public-url>
faultline fixtures ingest --adapter reddit-post --url <public-url>

faultline fixtures review

faultline fixtures promote <staging-id> --expected-playbook <id>
make build
./bin/faultline fixtures stats --class real --check-baseline
```

## Acceptance Bar

- the batch includes varied sources when available, not just repeated pulls from one thread
- the run is biased toward underrepresented adapters or failure classes when current corpus stats show a gap
- staging output is sanitized
- the promoted fixture has an explicit expected playbook
- the real corpus still passes its deterministic baseline gate
- any new ambiguity is handled immediately, not deferred
- any repository-inspection risk uncovered during intake is handed off to source-playbook fixture coverage rather than left implicit

## Source Selection Rules

- Prefer public reports with direct failure evidence over discussions with only speculation.
- Prefer one strong case each from several sources over many similar snippets from one source.
- Treat additional snippets from the same URL as guilty until proven useful.
- Promote repeated-source snippets only when they add a distinct failure boundary, not just more wording around the same one.

## Deliverable

- the exact ingestion and promotion commands run
- the source mix used
- promoted fixture IDs or rejected staging IDs
- any follow-on playbook refinement required by the new evidence
- any source-playbook fixture or regression follow-up required by the new evidence
