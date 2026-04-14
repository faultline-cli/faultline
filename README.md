# Faultline

CI failed. Faultline gives you a deterministic diagnosis from the log, the exact evidence that matched, and concrete fix steps you can act on immediately.

It is built for the repetitive failures that waste engineering time: missing credentials, version drift, lockfile mistakes, missing executables, runner problems, flaky tests, network failures, and other known CI breakages. Faultline runs locally or in CI, makes no network calls during analysis, and keeps ML or LLM systems out of the product path.

![Faultline terminal modes demo](docs/readme-assets/docker-auth-modes.gif)

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

CI failures are often repetitive, noisy, and slower to diagnose than they should be.

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

## What it catches today

Bundled playbooks currently cover 67 deterministic diagnoses across common CI failure categories, including:

- Auth: Docker registry auth, Git auth, SSH key auth, missing environment variables, AWS credential failures.
- Build: runtime mismatch, missing executable, dependency drift, lockfile drift, pip install failure, TypeScript compile failure, eslint failure, Go compile errors, missing build inputs, cache corruption.
- Network: DNS resolution failures, TLS certificate errors, blocked egress, connection refused, network timeout, rate limiting.
- Runtime: port already in use, permission denied, Docker daemon unavailable, disk full, OOM killed, resource limits.
- Test: flaky tests, test timeouts, coverage gate failures, snapshot mismatches, data races, missing test fixtures, order dependency.
- Deploy and CI: image pull backoff, CrashLoopBackOff, health-check failure, artifact upload failures, secrets unavailable, runner disk full, pipeline timeout.

The repository ships three small runnable sample logs for quick smoke tests and 112 accepted real fixtures for regression coverage.

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

- `./bin/faultline fixtures stats` currently reports 112 accepted real fixtures.
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


