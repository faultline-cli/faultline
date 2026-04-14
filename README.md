# Faultline

Deterministic CI failure analysis. No AI guessing.

When CI fails, you do not need another vague suggestion. You need a diagnosis, evidence, and fix steps you can trust.

Faultline analyzes CI logs and matches them against a deterministic library of real-world failure playbooks. Same input, same playbook set, same answer every time.

It is built for the repetitive failures that waste engineering time: missing credentials, version drift, lockfile mistakes, missing executables, runner problems, flaky tests, network failures, and other known CI breakages. Faultline runs locally or in CI, makes no network calls during analysis, and keeps ML or LLM systems out of the product path.

![Faultline terminal modes demo](docs/readme-assets/docker-auth-modes.gif)

## Example

Input log:

```text
pull access denied for ghcr.io/org/image
unauthorized: authentication required
```

Faultline diagnosis:

```text
Docker registry authentication failure (docker-auth) [33% confidence]
Severity: high

Summary
-------

CI could not authenticate to the container registry before an image pull or push.
```

Follow-up fix view:

```text
1. Verify the registry username, token, or password configured in CI secrets.
2. Ensure the registry login step runs before any docker pull or docker push command.
3. Confirm the token has the correct repository scope for the image being accessed.
```

This is the point of Faultline: take opaque CI output and turn it into a deterministic diagnosis you can act on immediately.

## Try it in 10 seconds

Pipe a failing log directly into the analyzer:

```bash
cat ci.log | faultline analyze
```

Or run the main flows directly:

```bash
faultline analyze ci.log
faultline fix ci.log
faultline inspect .
```

## Why Faultline

Most tools try to guess what went wrong.

Faultline does not guess.

- Deterministic pattern matching
- Ranked diagnoses with explicit evidence
- Structured fix steps instead of vague advice
- No LLMs or probabilistic output in the execution path

That makes it reliable in CI, explainable to engineers, and safe to automate against.

## Built on real failures

- 67 bundled playbooks under `playbooks/bundled/`
- 70 accepted real fixtures in the checked-in regression corpus
- Deterministic ranking, conflict review, and regression gates
- Stable terminal and JSON output for automation

## What Faultline catches

Faultline already detects a wide range of common CI and CD failures, including:

- Docker registry auth failures and missing login steps
- Image pull failures and registry access problems
- Missing binaries and PATH or execution errors in CI images
- Permission denied failures and runner file access issues
- Node, Python, Ruby, and Go runtime version mismatches
- npm, pnpm, yarn, pip, and Poetry install or lockfile failures
- Cache corruption and dependency drift
- DNS, TLS certificate, timeout, and connection-refused network failures
- Missing environment variables, invalid secrets, and expired credentials
- Incorrect working directories and missing build inputs
- TypeScript, eslint, syntax, and compile failures
- Flaky tests, timeout failures, data races, and snapshot mismatches
- Artifact upload, pipeline timeout, and runner disk-full failures
- ImagePullBackOff, CrashLoopBackOff, health-check, and deploy failures

The goal is not to catch everything. It is to reliably catch what is already known and explain it clearly.

## Try it in 60 seconds

Build the CLI and run it on a checked-in sample log:

```bash
make build
./bin/faultline analyze examples/docker-auth.log
```

Or use Docker without installing Go:

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/docker-auth.log
```

What you get back is a ranked diagnosis with evidence, not a generic summary. The bundled Docker auth example starts like this:

```md
# Docker registry authentication failure

- ID: `docker-auth`
- Confidence: 33%
- Category: auth
- Severity: high

## Summary

CI could not authenticate to the container registry before an image pull or push.
```

Tagged releases publish tarballs on the GitHub Releases page. If you are browsing the repository before a new release is cut, use `make build` or Docker for the fastest first run.

## Why it exists

CI failures are often repetitive, noisy, opaque, and slower to diagnose than they should be.

Faultline is built for engineers who want:

- deterministic results from explicit rules
- evidence pulled directly from the log
- fast local diagnosis without uploading build data
- stable terminal and JSON output for automation

It is intentionally narrow. Faultline does not try to explain every possible failure. It aims to be fast, repeatable, and trustworthy on failures it knows.

## Why trust it

- Same input and playbook set produce the same result every time.
- Evidence is pulled directly from matched log lines.
- Fix steps come from checked-in playbooks, not probabilistic generation.
- JSON output stays stable for automation and agent workflows.
- Analysis runs locally without shipping build logs to a hosted service.

## What it does

- Analyze CI logs from a file or stdin.
- Inspect a repository for source-level failure risks.
- Render concise terminal, markdown, or stable JSON output.
- Generate deterministic follow-up workflows from the analysis result.

## Install options

### Build from source

Requires Go 1.25+.

```bash
make build
./bin/faultline analyze examples/docker-auth.log
```

### Docker

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/docker-auth.log
```

### Release archive

Tagged releases publish tarballs named `faultline_<version>_<os>_<arch>.tar.gz` on the GitHub Releases page.

```bash
VERSION=<published-tag>
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

If you move the binary elsewhere, keep `playbooks/bundled/` beside it or set `FAULTLINE_PLAYBOOK_DIR`.

## First run examples

The repository includes runnable sample logs and expected markdown output.

```bash
./bin/faultline analyze examples/docker-auth.log
./bin/faultline analyze examples/missing-executable.log
./bin/faultline analyze examples/runtime-mismatch.log
./bin/faultline analyze examples/docker-auth.log --format markdown
./bin/faultline fix examples/docker-auth.log --format markdown
./bin/faultline explain docker-auth
```

More runnable examples and output snapshots are documented in `examples/README.md`.

## Core commands

| Command | Purpose |
| --- | --- |
| `analyze [file]` | Diagnose a CI log from a file or stdin |
| `fix [file]` | Print fix steps for the top diagnosis |
| `inspect [path]` | Scan a repository for source-level findings |
| `explain <id>` | Show the full playbook for one diagnosis |
| `list` | List bundled and installed playbooks |
| `workflow [file]` | Generate a deterministic follow-up workflow |
| `completion` | Generate shell completion scripts |

Useful flags:

| Flag | Description |
| --- | --- |
| `--json` | Emit machine-readable JSON |
| `--format terminal\|markdown\|json` | Choose the output format |
| `--mode quick\|detailed` | Control human-readable output detail |
| `--top N` | Show the top N ranked diagnoses |
| `--git` | Enrich analysis with recent local git context |
| `--repo <path>` | Choose the repository used by `--git` |

Advanced usage:

- `packs` installs and manages optional extra playbook packs after the bundled catalog is no longer enough.

## How it works

1. Faultline normalizes the input log into stable lines.
2. It loads deterministic YAML playbooks from the bundled catalog and any optional installed packs.
3. It matches explicit patterns, ranks results with stable rules, and extracts supporting evidence.
4. It returns a diagnosis, evidence, fix steps, and validation guidance.

The same input and playbook set should produce the same result every time.

## Support matrix

| Capability | Supported |
| --- | --- |
| Local log files | Yes |
| Stdin input | Yes |
| Stable JSON output | Yes |
| Docker usage | Yes |
| CI usage | Yes |
| Local repo inspection | Yes |
| Network calls during analysis | No |

## Credibility checks

- `./bin/faultline fixtures stats --class real` currently reports 70 accepted real fixtures.
- The checked-in regression snapshot reports top-1 = 1.000, top-3 = 1.000, unmatched = 0.000, false_positive = 0.000.
- The bundled catalog currently ships 67 playbooks under `playbooks/bundled/`.
- Release validation runs `make test`, `make review`, `make fixture-check`, release archive smoke tests, and Docker smoke tests.

These numbers describe the checked-in regression corpus, not the full space of CI failures.

## Repository guide

- `examples/README.md` shows runnable sample logs and expected output.
- `docs/product-spec.md` describes user-facing product behavior and positioning.
- `docs/implementation-status.md` captures the current CLI-only repository status.
- `docs/architecture.md` explains package boundaries and runtime flow.
- `docs/playbooks.md` documents playbook authoring and pack composition.
- `docs/distribution.md` covers release and Docker packaging.
- `docs/detectors.md` describes detector behavior.
- `docs/adr/README.md` indexes architectural decisions.
- `CONTRIBUTING.md` covers contribution and fixture-sanitization rules.

## Development

```bash
make build
make test
make review
make demo-assets
```

`make demo-assets` regenerates the README GIFs and screenshots from the VHS tapes under `docs/readme-assets/tapes/`.

## Feedback

The most useful issue is a sanitized CI failure that Faultline should diagnose better. Include the smallest log excerpt that reproduces the problem, the expected diagnosis, and what Faultline returned instead.

Raw ingestion artifacts belong in `fixtures/staging/` only as a local review queue. Sanitize them before promotion into `fixtures/real/`.

## License

Faultline is licensed under MIT. See `LICENSE`.


