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
2. Build a mixed candidate batch across multiple adapters when possible:
   - `github-issue`
   - `gitlab-issue`
   - `stackexchange-question`
   - `discourse-topic`
   - `reddit-post`
3. Prefer breadth over depth:
   - take one or two strong candidates from a source before harvesting more from the same thread
   - avoid spending the whole run on one issue, one subreddit post, or one discussion thread unless it clearly yields distinct failure signatures
4. Run `faultline fixtures ingest --adapter ... --url ...` for each candidate URL.
5. Review staged results with `faultline fixtures review`.
6. Reject duplicates, near-duplicates without new signal, setup-only snippets, workaround-only snippets, and cases that still need sanitization.
7. For accepted cases, promote with `faultline fixtures promote <staging-id> --expected-playbook <id>` plus any needed `--strict-top-1`, `--disallow`, or `--expected-stage` guards.
8. Rebuild and run `./bin/faultline fixtures stats --class real --check-baseline`.
9. If a promoted fixture exposes weak matching or a confusable neighbor, continue with [`refine-playbook-from-fixture.md`](./refine-playbook-from-fixture.md) before stopping.

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
- staging output is sanitized
- the promoted fixture has an explicit expected playbook
- the real corpus still passes its deterministic baseline gate
- any new ambiguity is handled immediately, not deferred

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
