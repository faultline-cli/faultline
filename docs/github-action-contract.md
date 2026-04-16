# GitHub Action Contract

Faultline's core CLI stays provider-agnostic. A GitHub-specific integration should be a thin wrapper over stable CLI commands and output artifacts, not a fork of the product logic.

## Contract Surface

The recommended surfaces for a separate `faultline-action` repository are:

- human summary: `faultline analyze <logfile> --format markdown`
- machine-readable diagnosis: `faultline analyze <logfile> --json`
- deterministic next-step handoff: `faultline workflow <logfile> --json --mode agent`
- optional evidence-fusion metadata: `faultline analyze <logfile> --json --bayes`

These contracts already exist in the CLI and should remain the integration boundary.

## Design Rules

- Keep GitHub-specific wiring out of `internal/`.
- Use file input or stdin capture, not provider-specific parsing inside the CLI.
- Treat the CLI JSON and workflow payloads as the source of truth for artifacts.
- Let GitHub choose presentation, upload, and threshold policy outside the core binary.
- Do not duplicate matching, ranking, or workflow logic in the action repository.

## Recommended Action Flow

1. Capture the failing log into a file inside the workflow job.
2. Run `faultline analyze` to produce markdown and JSON artifacts.
3. Optionally run `faultline analyze --json --bayes` when the action wants additive ranking or delta payloads.
4. Run `faultline workflow --json --mode agent` to produce the deterministic follow-up artifact.
5. Publish the markdown summary and upload the JSON outputs as workflow artifacts.

## Example Commands

Using a local binary:

```bash
faultline analyze build.log --format markdown > faultline-summary.md
faultline analyze build.log --json > faultline-analysis.json
faultline analyze build.log --json --bayes > faultline-analysis-bayes.json
faultline workflow build.log --json --mode agent > faultline-workflow.json
```

Using Docker:

```bash
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --format markdown > faultline-summary.md
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --json > faultline-analysis.json
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --json --bayes > faultline-analysis-bayes.json
docker run --rm -v "$PWD":/workspace faultline workflow /workspace/build.log --json --mode agent > faultline-workflow.json
```

## Notes For v0.2.0

- Keep `workflow.v1` stable unless an explicit breaking version is introduced.
- Additive JSON fields are acceptable; silent field removals or renames are not.
- `--bayes` must stay additive and explainable; it is a ranking aid, not a second matcher.
- If GitHub summaries or annotations need policy thresholds, keep those decisions in the action repository rather than the core CLI.
