# GitLab CI Contract

Faultline's core CLI stays provider-agnostic. A GitLab CI integration should be a thin wrapper over stable CLI commands and output artifacts, not a fork of product logic.

For teams running GitLab pipelines, this wrapper is a follow-up path after validating the local CLI flow. It should stay a thin integration layer over the default product path, not redefine it.

## Contract Surface

The recommended surfaces for a separate wrapper project or pipeline template are:

- human summary: `faultline analyze <logfile> --format markdown`
- machine-readable diagnosis: `faultline analyze <logfile> --json`
- deterministic next-step handoff: `faultline workflow <logfile> --json --mode agent`

These contracts already exist in the CLI and should remain the integration boundary.

`workflow.v1` is the deterministic handoff contract for downstream scripts and agents. Additive fields are allowed; silent removals or renames are not.

Current additive workflow JSON may also carry `ranking_hints`, `delta_hints`,
`metrics_hints`, and `policy_hints` when the underlying analysis has enough
explicit context. Those fields should remain optional and omitted when absent.

## Design Rules

- Keep GitLab-specific wiring out of `internal/`.
- Use file input or stdin capture, not provider-specific parsing inside the CLI.
- Treat CLI JSON and workflow payloads as the source of truth for artifacts.
- Let GitLab choose presentation, artifact upload, and threshold policy outside the core binary.
- Do not duplicate matching, ranking, or workflow logic in wrapper repositories.

## Recommended Pipeline Flow

1. Capture the failing job log into a file in the pipeline workspace.
2. Run `faultline analyze` to produce markdown and JSON artifacts, including ranking metadata.
3. Run `faultline workflow --json --mode agent` to produce the deterministic follow-up artifact.
4. Publish markdown summaries and upload JSON outputs as job artifacts.
5. Optionally gate follow-up automation with deterministic confidence and playbook thresholds in pipeline logic, not in core CLI logic.

## Example Commands

Using a local binary:

```bash
faultline analyze build.log --format markdown > faultline-summary.md
faultline analyze build.log --json > faultline-analysis.json
faultline workflow build.log --json --mode agent > faultline-workflow.json
```

Using Docker:

```bash
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --format markdown > faultline-summary.md
docker run --rm -v "$PWD":/workspace faultline analyze /workspace/build.log --json > faultline-analysis.json
docker run --rm -v "$PWD":/workspace faultline workflow /workspace/build.log --json --mode agent > faultline-workflow.json
```

## Example `.gitlab-ci.yml` Step

```yaml
faultline_analyze:
  stage: post
  when: on_failure
  script:
    - faultline analyze build.log --format markdown > faultline-summary.md
    - faultline analyze build.log --json > faultline-analysis.json
    - faultline workflow build.log --json --mode agent > faultline-workflow.json
  artifacts:
    when: always
    paths:
      - faultline-summary.md
      - faultline-analysis.json
      - faultline-workflow.json
```

## Compatibility Notes

- Keep `workflow.v1` stable unless an explicit breaking version is introduced.
- Additive JSON fields are acceptable; silent field removals or renames are not.
- Bayesian reranking is enabled by default and must stay additive and explainable; it is a ranking aid, not a second matcher. Pass `--bayes=false` to disable.
- Keep integration policy decisions in wrapper or pipeline repositories, not in core CLI logic.
- Provider-backed delta is intentionally outside the shipped core CLI contract.
