# Faultline

Faultline is a deterministic CLI that explains CI failures from logs and repository scans.

It is built for engineers who own broken pipelines, release jobs, and deployment failures and need a fast first answer they can trust. CI failures are repetitive, noisy, and expensive to reread by hand; Faultline turns that first-pass triage into something consistent, local, and reviewable.

## Why it is useful

- Diagnose a CI log from a file or stdin.
- Show the exact evidence lines behind the diagnosis.
- Return concrete fix and validation steps instead of vague summaries.
- Inspect a repository tree for source-level failure risks.
- Emit stable text, markdown, and JSON for humans and automation.
- Stay deterministic: same input, same playbooks, same result.
- Avoid LLM drift, hosted analysis services, and hidden heuristics.

## Example

Real CI log input:

```text
> docker pull mcr.microsoft.com/mssql/server:2017-latest-ubuntu
Error response from daemon: Get https://mcr.microsoft.com/v2/: Forbidden

> docker --debug pull mcr/microsoft.com/mssql/server:2017-latest-ubuntu
Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'

> docker --debug pull mcr.microsoft.com/mssql/server
Using default tag: latest
Error response from daemon: Get https://mcr.microsoft.com/v2/: Forbidden
```

Command:

```bash
faultline analyze build.log --format markdown --mode detailed
```

Output:

```markdown
# Docker registry authentication failure

- ID: `docker-auth`
- Confidence: 33%
- Category: auth
- Severity: high
- Score: 2.00
- Detector: log
- Stage: test

## Summary

CI could not authenticate to the container registry before an image pull or push.

## Evidence

- Error response from daemon: pull access denied for mcr/microsoft.com/mssql/server, repository does not exist or may require 'docker login'
```

## Quick Start

Build from source and run the bundled examples:

```bash
make build
./bin/faultline analyze examples/docker-auth.log
./bin/faultline analyze examples/missing-executable.log
./bin/faultline analyze examples/runtime-mismatch.log
```

Minimal usage:

```bash
# Analyze a log file
faultline analyze build.log

# Read from stdin
cat build.log | faultline analyze

# Emit stable JSON for automation
faultline analyze build.log --json

# Show only the fix steps for the top diagnosis
faultline fix build.log --format markdown

# Inspect a repository for source-level findings
faultline inspect .
```

Docker:

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/docker-auth.log
```

Release archive:

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze examples/docker-auth.log
```

## Commands

| Command | Purpose |
| --- | --- |
| `analyze [file]` | Diagnose a CI log from a file or stdin |
| `fix [file]` | Print the fix steps for the top diagnosis |
| `inspect [path]` | Inspect a repository for source-level failure risks |
| `workflow [file]` | Generate a deterministic follow-up workflow |
| `list` | List available playbooks |
| `explain <id>` | Show the details of one playbook |

Common flags:

| Flag | Description |
| --- | --- |
| `--json` | Emit machine-readable JSON |
| `--format raw\|markdown` | Select the human-readable output format |
| `--mode quick\|detailed` | Control output detail for human-readable results |
| `--top N` | Show the top N ranked results |
| `--git` | Enrich analysis with recent local git context |
| `--repo <path>` | Choose the repository path used by `--git` |

## Playbooks

Playbooks are deterministic rules plus operator-facing guidance.

- Structured fields decide whether a log or repository state matches.
- Evidence lines show why that playbook won.
- Summary, diagnosis, fix, and validation fields turn the match into an actionable response.

This keeps the engine explicit and reviewable while still producing useful operator output.

Additional playbook packs with extended coverage are planned.

## Credibility

- Real regression corpus under `fixtures/real/` built from public CI failures.
- Deterministic engine with stable ranking and evidence-first output.
- Real-world example logs under `examples/` with checked-in expected output.
- Source inspection support for repository-level failure risks.
- Release and regression workflow that exercises tests, overlap review, and delivery paths.

## Feedback

The highest-value contribution is a real failure that Faultline should explain better.

- Open an issue with the failing log snippet, expected diagnosis, and relevant context.
- Add or refine fixtures when a failure should be preserved in regression tests.
- If Faultline misses a recurring CI failure, send that case. Those misses are the fastest way to improve the tool.

Raw ingestion artifacts belong in `fixtures/staging/` as a local review queue only. Sanitize them first, then promote accepted cases into `fixtures/real/`.

## License

Faultline is licensed under Apache-2.0. See `LICENSE` for the full text.

## Development

```bash
make build
make test
make review
```

Helpful references:

- `examples/README.md` for runnable sample logs and expected outputs
- `docs/architecture.md` for package boundaries and runtime flow
- `docs/playbooks.md` for playbook authoring rules
- `docs/distribution.md` for release and Docker packaging
- `docs/detectors.md` for detector expectations
- `docs/adr/README.md` for architectural decisions
- `CONTRIBUTING.md` for contribution and fixture-sanitization guidance
- `SECURITY.md` for vulnerability reporting expectations
- `CODE_OF_CONDUCT.md` for project participation expectations
- `LICENSE` for the Apache-2.0 terms


