# GitHub Action Contract

Faultline's core CLI stays provider-agnostic. A GitHub-specific integration should be a thin wrapper over stable CLI commands and output artifacts, not a fork of the product logic.

For most teams, this wrapper is the strongest follow-up path after validating the local CLI flow. It should stay a thin integration layer over the default product path, not redefine it.

## Contract Surface

The recommended surfaces for a separate `faultline-action` repository are:

- human summary: `faultline analyze <logfile> --format markdown`
- machine-readable diagnosis: `faultline analyze <logfile> --json`
- deterministic next-step handoff: `faultline workflow <logfile> --json --mode agent`
- experimental failure delta against the last successful run on the same branch:
  `FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1 faultline analyze <logfile> --json --delta-provider github-actions`

These contracts already exist in the CLI and should remain the integration boundary.
GitHub annotations, summaries, and check-run presentation should be generated
by the wrapper from JSON or markdown artifacts rather than by the core CLI.

`workflow.v1` is the deterministic handoff contract for downstream scripts and agents. Additive fields are allowed; silent removals or renames are not.

Current additive workflow JSON may also carry `ranking_hints`, `delta_hints`,
`metrics_hints`, and `policy_hints` when the underlying analysis has enough
explicit context. Those fields should remain optional and omitted when absent.

## Design Rules

- Keep GitHub-specific wiring out of `internal/`.
- Use file input or stdin capture, not provider-specific parsing inside the CLI.
- Treat the CLI JSON and workflow payloads as the source of truth for artifacts.
- Let GitHub choose presentation, upload, and threshold policy outside the core binary.
- Do not duplicate matching, ranking, or workflow logic in the action repository.

## Recommended Action Flow

1. Capture the failing log into a file inside the workflow job.
2. Run `faultline analyze` to produce markdown and JSON artifacts, including ranking metadata.
3. Optionally enable the experimental provider-backed delta path with `FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1 --delta-provider github-actions` and pass `GITHUB_TOKEN`, `GITHUB_REPOSITORY`, `GITHUB_REF_NAME`, and `GITHUB_RUN_ID`.
4. Run `faultline workflow --json --mode agent` to produce the deterministic follow-up artifact.
5. Publish the markdown summary and upload the JSON outputs as workflow artifacts.
6. Optionally gate follow-up automation based on deterministic confidence and playbook thresholds in the action repository, not in core CLI logic.

## Example Commands

Using a local binary:

```bash
faultline analyze build.log --format markdown > faultline-summary.md
faultline analyze build.log --json > faultline-analysis.json
FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1 \
GITHUB_TOKEN="$GITHUB_TOKEN" \
GITHUB_REPOSITORY="$GITHUB_REPOSITORY" \
GITHUB_REF_NAME="$GITHUB_REF_NAME" \
GITHUB_RUN_ID="$GITHUB_RUN_ID" \
faultline analyze build.log --json --delta-provider github-actions > faultline-analysis-delta.json
faultline workflow build.log --json --mode agent > faultline-workflow.json
```

Using Docker:

```bash
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --format markdown > faultline-summary.md
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --json > faultline-analysis.json
docker run --rm -v "$PWD":/workspace faultline workflow /workspace/build.log --json --mode agent > faultline-workflow.json
```

## Example Wrapper Step

```yaml
- name: Analyze failing job with Faultline
  if: failure()
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    faultline analyze build.log --format markdown > faultline-summary.md
    faultline analyze build.log --json > faultline-analysis.json
    FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1 faultline analyze build.log --json --delta-provider github-actions > faultline-analysis-delta.json
    faultline workflow build.log --json --mode agent > faultline-workflow.json
```

## Compatibility Notes

- Keep `workflow.v1` stable unless an explicit breaking version is introduced.
- Additive JSON fields are acceptable; silent field removals or renames are not.
- Bayesian reranking is enabled by default and must stay additive and explainable; it is a ranking aid, not a second matcher. Pass `--bayes=false` to disable.
- If GitHub summaries or annotations need policy thresholds, keep those decisions in the action repository rather than the core CLI.
- Legacy `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1` still enables this path, but new integrations should use `FAULTLINE_EXPERIMENTAL_PROVIDER_DELTA=1`.
