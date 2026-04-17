# Faultline

Deterministic, audit-friendly CI automation for known failures. No AI guessing - ranked diagnoses, workflow handoffs, and checked-in fix steps.

```text
# CI log
exec /__e/node20/bin/node: no such file or directory

# Faultline
[1] missing-executable (confidence: 84%)

Diagnosis:
Required executable or runtime binary missing.

Fix:
- Install the missing runtime or tool in the CI image
- Pin the runner or action to an image that includes the expected binary
- Verify the configured path still exists after recent runner or action upgrades
```

Faultline analyzes CI logs against a deterministic library of real-world failure playbooks. Same input, same playbook set, same answer every time.

It is built for repetitive failures that waste engineering time: missing credentials, version drift, lockfile mistakes, missing executables, runner problems, flaky tests, network failures, and other known CI breakages. Faultline runs locally or in CI, makes no network calls during analysis, and keeps ML or LLM systems out of the product path.

Works on any CI log, including GitHub Actions, GitLab CI, and similar systems.

Run it when CI fails:

```bash
faultline analyze ci.log
faultline workflow ci.log
faultline analyze ci.log --bayes
faultline analyze ci.log --json
faultline workflow ci.log --json --mode agent
faultline guard .
```

`analyze` gives you a ranked diagnosis with evidence. `workflow` turns the same result into a deterministic next-step artifact that engineers, scripts, and agents can follow without inventing their own glue. `guard` uses the same evidence model for quiet, high-confidence local checks before CI.

## Try it now

Install the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
```

Prefer a pinned release instead of latest:

```bash
VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
```

For automation, use JSON output:

```bash
faultline analyze ci.log --json
faultline workflow ci.log --json --mode agent
```

Build from source if you want to work from the repo directly. Requires Go 1.25+.

```bash
git clone https://github.com/faultline-cli/faultline
cd faultline
go build -o faultline ./cmd
./faultline analyze examples/missing-executable.log
```

Pipe any failing log directly into the analyzer:

```bash
cat failing-ci.log | ./faultline analyze
cat failing-ci.log | ./faultline analyze --json
```

![Faultline missing executable demo](docs/readme-assets/missing-executable.gif)

## Why Faultline

Most tools try to guess what went wrong.

Faultline does not guess at detection.

- Deterministic pattern matching
- Ranked diagnoses with explicit evidence
- Differential diagnosis for close calls: likely cause, alternatives, and ruled-out rivals
- Optional `--bayes` reranking that stays deterministic, explainable, and additive
- Deterministic workflow artifacts for local follow-through and agent handoff
- Quiet `guard` checks for high-confidence local prevention
- Structured fix steps instead of vague advice
- No LLMs, no opaque ranking, and no non-reproducible output
- Audit-friendly output with evidence pulled directly from the log
- Faultline only emits a diagnosis when the match clears its confidence threshold

That makes it reliable in CI, explainable to engineers, and safe to automate against.

## Handles

- Docker and registry authentication failures
- Missing executables, PATH problems, and command invocation errors
- Runtime version mismatches across Node, Python, Ruby, and Go
- Dependency install, resolver, and lockfile failures
- Cache corruption and dependency drift
- Permission issues and filesystem access failures
- CI config errors, bad working directories, and missing build inputs
- Git checkout and runner failures
- Environment variable problems, invalid secrets, and expired credentials
- DNS, TLS, timeout, and connection failures
- Compile, lint, test, and deploy failures

The goal is not to catch everything. It is to reliably catch what is already known and explain it clearly.

## Built on real failures

- 77 bundled playbooks under `playbooks/bundled/`
- 84 accepted real fixtures in the checked-in regression corpus
- Deterministic ranking, conflict review, and regression gates
- Stable terminal, JSON, and workflow output for automation

The current corpus snapshot and validation commands are published in [`docs/fixture-corpus.md`](docs/fixture-corpus.md).
Repository-specific agent operating loops for fixture curation, unmatched-log triage, playbook refinement, and deterministic verification are published in [`docs/agent-workflows.md`](docs/agent-workflows.md).

## Try it in 60 seconds

Build the CLI and run it on a checked-in sample log:

```bash
make build
./bin/faultline analyze examples/missing-executable.log
cat examples/missing-executable.log | ./bin/faultline workflow --no-history
```

Or use Docker without installing Go:

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/missing-executable.log
```

What you get back is a ranked diagnosis with evidence, not a generic summary. When you want the follow-through artifact as well, the same example also produces a deterministic workflow plan:

```text
WORKFLOW  missing-executable · Required executable or runtime binary missing  [local · workflow.v1]
Source: stdin
Evidence:
  - exec /__e/node20/bin/node: no such file or directory
Focus files:
  - Dockerfile
  - .github/workflows/ci.yml
```

The full checked-in snapshots live in:

- `examples/missing-executable.workflow.local.txt`
- `examples/missing-executable.workflow.agent.json`

Designed to run inside CI pipelines, the bundled missing-executable diagnosis example starts like this:

```md
# Required executable or runtime binary missing

- ID: `missing-executable`
- Confidence: 84%
- Category: build
- Severity: high

## Summary

The job tried to launch a required tool or runtime binary, but that executable was missing from the image, runner, or expected path.
```

## Why it exists

CI failures are often repetitive, noisy, opaque, and slower to diagnose than they should be.

Faultline is built for engineers who want:

- deterministic results from explicit rules
- evidence pulled directly from the log
- fast local diagnosis without uploading build data
- stable terminal, JSON, and workflow output for automation

It is intentionally narrow. Faultline does not try to explain every possible failure. It aims to be fast, repeatable, and trustworthy on failures it knows. Designed to minimise false positives: better no result than a wrong one.

## Why trust it

- Same input and playbook set produce the same result every time.
- Evidence is pulled directly from matched log lines.
- Fix steps come from checked-in playbooks, not probabilistic generation.
- `--bayes` never creates new matches; it only reranks already-detected candidates and explains why.
- Repo signals only participate when Faultline has explicit repository context (`--git` or guard/inspect). They are additive: they never create matches, only enrich an existing diagnosis with config drift, CI config changes, large-commit blast radius, hotspot files and directories, co-change pairs, hotfix or revert indicators, and CODEOWNERS-derived ownership boundary signals (boundary crossing, upstream component changes, ownership mismatch, failure clustering).
- JSON and workflow output stay stable for automation and agent workflows.
- Analysis runs locally without shipping build logs to a hosted service.

## What it does

- Analyze CI logs from a file or stdin.
- Rerank close calls with `--bayes`.
- Explain why the winning diagnosis beat nearby alternatives.
- Surface likely drift causes only when repo context is explicit: config file changes, CI pipeline edits, large blast-radius commits, and hotspot patterns from recent history.
- Turn the top diagnosis into a deterministic workflow handoff.
- Inspect a repository for source-level failure risks.
- Run quiet, high-confidence local checks with `guard`.
- Render concise terminal, markdown, or stable JSON output.
- Use checked-in playbooks and real-fixture regression gates as the trust boundary.

## Install options

### One-command installer

This is the default install path.

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
```

### Build from source

Requires Go 1.25+.

```bash
git clone https://github.com/faultline-cli/faultline
cd faultline
go build -o faultline ./cmd
./faultline analyze examples/missing-executable.log
```

### Docker

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/missing-executable.log
```

### Release archive

Release archives are published on the GitHub Releases page.

```bash
VERSION=v0.2.0
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

If you move the binary elsewhere, keep `playbooks/bundled/` beside it or set `FAULTLINE_PLAYBOOK_DIR`.

## First run examples

The repository includes runnable sample logs and expected markdown output.

```bash
./bin/faultline analyze examples/missing-executable.log
./bin/faultline analyze examples/runtime-mismatch.log
./bin/faultline analyze examples/docker-auth.log
./bin/faultline analyze examples/missing-executable.log --format markdown
./bin/faultline analyze examples/missing-executable.log --json --bayes
cat examples/missing-executable.log | ./bin/faultline workflow --no-history
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --bayes --no-history
./bin/faultline guard .
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --no-history
./bin/faultline fix examples/missing-executable.log --format markdown
./bin/faultline explain missing-executable
```

More runnable examples and output snapshots are documented in `examples/README.md`.

## Core commands

| Command | Purpose |
| --- | --- |
| `analyze [file]` | Diagnose a CI log from a file or stdin |
| `fix [file]` | Print fix steps for the top diagnosis |
| `inspect [path]` | Scan a repository for source-level findings |
| `guard [path]` | Emit only high-confidence local prevention findings |
| `explain <id>` | Show the full playbook for one diagnosis |
| `list` | List bundled and installed playbooks |
| `packs` | Install and list optional extra playbook packs |
| `workflow [file]` | Generate a deterministic follow-up workflow |
| `completion` | Generate shell completion scripts |

Useful flags:

| Flag | Description |
| --- | --- |
| `--json` | Emit machine-readable JSON |
| `--format terminal\|markdown\|json` | Choose the output format |
| `--mode quick\|detailed` | Control human-readable output detail |
| `--top N` | Show the top N ranked diagnoses |
| `--bayes` | Apply deterministic Bayesian-inspired reranking |
| `--ci-annotations` | Emit GitHub Actions annotations during analysis |
| `--delta-provider github-actions` | Compare against the last successful GitHub Actions run on the same branch |
| `--git` | Enrich analysis with recent local git context (config drift, CI changes, large commits, hotspots, hotfix and revert signals, CODEOWNERS ownership boundary signals) |
| `--repo <path>` | Choose the repository used by `--git` |

Advanced usage:

- `packs` installs and manages optional extra playbook packs after the bundled catalog is no longer enough.

## How it works

1. Faultline normalizes the input log into stable lines.
2. It loads deterministic YAML playbooks from the bundled catalog and any optional installed packs.
3. It matches explicit patterns, extracts supporting evidence, and ranks results with stable rules.
4. When `--bayes` is enabled, it reranks only the already-matched candidates and adds explainable ranking hints.
5. When repo context is explicit (`--git`), it attaches additive history signals to the diagnosis: recently changed config and dependency files, CI pipeline file edits, hotspot directories and files, co-change pairs, large blast-radius commits, hotfix and revert patterns, and author breadth across the commit window. It also parses CODEOWNERS, builds an ownership graph, and derives topology signals for ownership boundary crossings, upstream component changes, ownership mismatches, and localised failure clustering.
6. It returns a diagnosis, evidence, fix steps, workflow hints, and validation guidance.

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
| Local guard checks | Yes |
| Network calls during analysis | No |

## Credibility checks

- `./bin/faultline fixtures stats --class real` currently reports 84 accepted real fixtures and a `weak_match` rate of `0.119` (10/84).
- The checked-in regression snapshot reports top-1 = 1.000, top-3 = 1.000, unmatched = 0.000, false_positive = 0.000.
- The bundled catalog currently ships 77 playbooks under `playbooks/bundled/`.
- Release validation runs `make test`, `make review`, `make fixture-check`, release archive smoke tests, and Docker smoke tests.

These numbers describe the checked-in regression corpus, not the full space of CI failures.

## Repository guide

- `examples/README.md` shows runnable sample logs and expected output.
- `docs/fixture-corpus.md` publishes the checked-in regression snapshot and regeneration commands.
- `docs/failures/README.md` indexes search-targeted CI failure pages tied to Faultline diagnoses.
- `docs/architecture.md` explains package boundaries and runtime flow.
- `docs/github-action-contract.md` documents the provider-agnostic CLI contract for a thin GitHub Action wrapper.
- `docs/playbooks.md` documents playbook authoring and pack composition.
- `docs/distribution.md` covers release and Docker packaging.
- `docs/releases/v0.2.0-preview.md` drafts the v0.2.0 positioning and release notes.
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

The most useful issue is a sanitized CI failure that Faultline should diagnose better. Have a failure this doesn't catch? Open an issue with the log. Include the smallest log excerpt that reproduces the problem, the expected diagnosis, and what Faultline returned instead.

Raw ingestion artifacts belong in `fixtures/staging/` only as a local review queue. Sanitize them before promotion into `fixtures/real/`.

## License

Faultline is licensed under MIT. See `LICENSE`.
