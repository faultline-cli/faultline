# Faultline

Faultline is a deterministic CLI that turns noisy CI failures into a ranked diagnosis with evidence and concrete next steps.

It is built for engineers who own broken pipelines, release jobs, or deploy logs and want a fast answer they can trust. Instead of re-reading the same auth, dependency, runtime, and deploy failures by hand, Faultline gives you a repeatable first-pass triage path that works locally, in CI, and inside automation.

## Why Teams Use It

- Diagnose CI failures from a log file or stdin.
- Show exact evidence lines that triggered the diagnosis.
- Return practical fix and validation steps instead of vague summaries.
- Inspect a repository tree for source-level failure risks.
- Emit stable text, markdown, JSON, and workflow output.
- Preserve deterministic behavior: same input, same playbooks, same result.
- Avoid LLM drift and hidden server-side logic.

Additional playbook packs with extended coverage are planned.

## Example

Real log excerpt:

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
faultline analyze build.log --format markdown
```

Faultline output:

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
```

## Quick Start

### Build from source

```bash
make build
./bin/faultline analyze examples/docker-auth.log
```

### Use a release archive

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze examples/docker-auth.log
```

### Docker

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/docker-auth.log
```

### Minimal usage

```bash
# Analyze a log file
faultline analyze build.log

# Read from stdin
cat build.log | faultline analyze

# Stable JSON for automation
faultline analyze build.log --json

# Show only the fix steps for the top diagnosis
faultline fix build.log

# Inspect a repository for source-level findings
faultline inspect .

# Build a deterministic follow-up workflow
faultline workflow build.log --mode agent --git --repo .
```

## Commands

| Command | Description |
| --- | --- |
| `analyze [file]` | Diagnose a CI log from a file or stdin |
| `fix [file]` | Print only the fix steps for the top diagnosis |
| `inspect [path]` | Inspect a repository tree for source-level failure risks |
| `workflow [file]` | Build a deterministic follow-up workflow |
| `list` | List available playbooks |
| `explain <id>` | Show full detail for one playbook |
| `packs install <dir>` | Install an extra playbook pack for automatic loading |
| `packs list` | List installed extra packs |
| `fixtures ingest` | Fetch public CI failure snippets into staging |
| `fixtures review` | Review staged fixtures with predicted matches |
| `fixtures promote` | Promote reviewed fixtures into the real corpus |
| `fixtures stats` | Report regression metrics across fixture corpora |

Common flags:

| Flag | Default | Description |
| --- | --- | --- |
| `--top N` | `1` | Show top N ranked results |
| `--mode quick\|detailed` | `quick` | Human output verbosity |
| `--format raw\|markdown` | `raw` | Human-readable output shape |
| `--json` | `false` | Emit stable JSON |
| `--playbooks <dir>` | auto | Replace the active catalog with one directory |
| `--playbook-pack <dir>` | none | Add one or more extra pack roots |
| `--git` | `false` | Enrich analysis with recent local git context |
| `--since <window>` | `30d` | History window for `--git` |
| `--repo <path>` | `.` | Repository path used by `--git` |

## Playbooks

Playbooks are deterministic rules plus operator-facing guidance.

- The structured fields decide whether a failure matches.
- The evidence lines explain why the match won.
- The diagnosis, fix, and validation sections turn the match into an actionable response.

This keeps Faultline reviewable and stable: the matching logic stays explicit, while the human guidance stays concise and useful.

## Credibility

- Real regression corpus under `fixtures/real/` built from public CI failures.
- Deterministic engine with stable ranking and evidence-first output.
- Source inspection support for repository-level failure risks.
- Fixture promotion and review flow that keeps new coverage testable.
- Release workflow that checks tests, playbook review, packaging, and smoke paths.

## Try The Examples

The `examples/` directory contains runnable samples built from real failure patterns:

- `examples/docker-auth.log`
- `examples/missing-executable.log`
- `examples/runtime-mismatch.log`

Each sample has a matching expected output file so you can compare behavior quickly.

```bash
./bin/faultline analyze examples/docker-auth.log
./bin/faultline analyze examples/missing-executable.log
./bin/faultline analyze examples/runtime-mismatch.log
```

## How Detection Works

1. Read log input from a file path or stdin, or walk a repository tree for source inspection.
2. Normalize the input into stable lines or source signals.
3. Load the bundled catalog plus any installed or explicitly composed extra packs.
4. Validate playbook structure and review overlap conflicts deterministically.
5. Match and rank with explicit rules, not hidden heuristics.
6. Render concise text, markdown, JSON, or workflow output.

The product constraints live in `SYSTEM.md`.

## Development

```bash
make build
make test
make review
make fixture-check
```

Useful targets:

- `make bench` for deterministic load and analysis benchmarks
- `make release-snapshot VERSION=v0.1.0` to build release archives
- `make smoke-release VERSION=v0.1.0` to verify a built release archive
- `make docker-smoke IMAGE=faultline-smoke` to verify the Docker delivery path

## Contributing

The highest-value contribution is a real failure that Faultline should explain better.

- Submit an issue with the failing log snippet, expected diagnosis, and surrounding context.
- Add or refine fixtures when a failure pattern should be preserved in regression tests.
- Run `make test` and `make review` before opening a change.

If you have a recurring CI failure that Faultline misses, open an issue or send a fixture candidate. Those cases are the fastest path to making the tool more useful.

## Docs

- `docs/architecture.md`: core package boundaries and runtime flow
- `docs/playbooks.md`: playbook authoring rules and pack composition
- `docs/distribution.md`: release artifacts and delivery workflow
- `docs/detectors.md`: source-detector model and expectations
- `docs/adr/README.md`: architectural decision records
