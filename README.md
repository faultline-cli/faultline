# Faultline

Deterministic CI failure diagnosis. No guesswork. No AI.

Faultline analyzes failing CI runs with structured detectors and fixture-validated playbooks to produce reproducible, evidence-backed diagnoses.

- 🔁 Deterministic results: same input, same output
- 🧪 Fixture-backed validation: detections are regression tested
- 🏠 Local-first execution: build logs stay on your machine by default
- 🤖 Agent-legible output: stable JSON and deterministic workflow artifacts

```text
# CI log
exec /__e/node20/bin/node: no such file or directory

# Faultline: deterministic diagnosis
[1] missing-executable (confidence: 84%)
[2] path-lookup-failure (confidence: 22%)
[3] runner-image-upgrade (confidence: 8%)

Evidence:
  - exec /__e/node20/bin/node: no such file or directory
  - /__e/node20/bin/node: required runtime not found in PATH

Fix:
  - Install the missing runtime in the CI image
  - Pin the runner to an image that includes the expected binary
  - Verify the configured path after runner upgrades
```

Faultline is a deterministic diagnosis engine built on 77 bundled playbooks and 103 real-fixture regression proofs. Each diagnosis pulls evidence directly from matched log lines: nothing generated, nothing guessed.

## What else is already built 🧩

Faultline includes additional surfaces that are complete but intentionally outside the default onboarding workflow:

| Command | Purpose |
| --- | --- |
| `trace` | Rule-by-rule evaluation showing why each playbook matched or was rejected |
| `replay` | Deterministically re-render saved analysis artifacts |
| `compare` | Diff two analysis artifacts to show what changed between runs |
| `inspect` | Scan repository source for local failure risks |
| `guard` | Quiet, high-confidence local prevention checks |

These commands are supported and tested. They ship alongside the core `analyze` -> `workflow` path but stay outside the default first-run narrative until they meet promotion criteria in the release boundary.

Experimental provider-backed delta remains hidden behind explicit opt-in:

```bash
FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1 \
faultline analyze build.log --json --bayes --delta-provider github-actions
```

## Designed for humans and agents 🤝

Faultline is designed to be legible to both humans and automated agents.

It produces stable output contracts for both:

- 👩‍💻 **Human engineers** get terminal output with evidence, confidence, and fix steps—immediately actionable without interpretation.
- 🤖 **Automation and agents** get deterministic JSON via `--json --mode agent` that stays stable release-to-release.

```bash
# Engineer workflow
faultline analyze ci.log

# Agent/automation workflow
faultline analyze ci.log --json --mode agent > diagnosis.json
faultline workflow ci.log --json --mode agent > workflow.json
```

No ML, no LLM, no probabilistic ranking in the core product path. `--bayes` stays assistive: it reranks already-detected candidates and explains the ranking, it never creates new matches.

## Command Maturity Model 🧭

Faultline organizes commands by maturity to provide structural clarity for contributors:

| Tier | Commands | Description |
| --- | --- | --- |
| **Stable** | `analyze`, `workflow`, `list`, `explain`, `fix` | Default release boundary. Ship-ready, tested, documented. |
| **Complete** | `trace`, `replay`, `compare`, `inspect`, `guard` | Full feature parity. Tested and documented, but not the first-run story. |
| **Experimental** | `--delta-provider github-actions` (behind `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1`) | Opt-in only. Requires equivalent test coverage before promotion. |

Tier promotion follows the release-boundary contract: deterministic command coverage, fixture-backed regression proof, checked-in snapshots, and release-check integration.

## Try it now 🚀

Start locally with the core analyze -> workflow path:

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
faultline workflow ci.log --json --mode agent
```

That is the default product path: diagnose the failing log first, then turn the top diagnosis into a deterministic follow-up artifact.

Prefer a pinned release instead of latest:

```bash
VERSION=v0.3.0 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
```

For automation, use stable JSON output contracts:

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
cat failing-ci.log | ./faultline workflow --json --mode agent
```

When the local flow is working for you, the same commands drop cleanly into GitHub Actions:

```yaml
name: ci

on:
  push:
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # ... your build/test steps that write build.log ...
      - name: Analyze failure with Faultline
        if: failure()
        run: |
          VERSION=v0.3.0 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
          faultline analyze build.log --format markdown --ci-annotations > faultline-summary.md
          faultline analyze build.log --json --bayes > faultline-analysis.json
          faultline workflow build.log --json --mode agent > faultline-workflow.json
      - name: Upload Faultline artifacts
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: faultline
          path: |
            faultline-summary.md
            faultline-analysis.json
            faultline-workflow.json
```

GitHub Actions is a strong follow-up path because it preserves the same deterministic CLI contract while attaching the artifacts to the failure where your team already looks.

![Faultline missing executable demo](docs/readme-assets/missing-executable.gif)

The demo shows the default high-signal flow in one pass: deterministic diagnosis first, deterministic workflow handoff second.

## Why Faultline ✅

Most tools try to guess what went wrong.

Faultline does not guess at detection.

- 🎯 Deterministic pattern matching
- 📊 Ranked diagnoses with explicit evidence
- 🧭 Differential diagnosis for close calls: likely cause, alternatives, and ruled-out rivals
- ⚖️ Optional `--bayes` reranking that stays deterministic, explainable, and additive
- 🧩 Deterministic workflow artifacts for local follow-through and agent handoff
- 🛡️ Quiet `guard` checks for high-confidence local prevention
- 🧯 Structured fix steps instead of vague advice
- 🚫 No LLMs, no opaque ranking, and no non-reproducible output
- 🔍 Audit-friendly output with evidence pulled directly from the log
- ✅ Faultline only emits a diagnosis when the match clears its confidence threshold

That makes it reliable in CI, explainable to engineers, and safe to automate against.

## Handles 🛠️

- 🐳 Docker and registry authentication failures
- 🧰 Missing executables, PATH problems, and command invocation errors
- 🧬 Runtime version mismatches across Node, Python, Ruby, and Go
- 📦 Dependency install, resolver, and lockfile failures
- 🗄️ Cache corruption and dependency drift
- 🔐 Permission issues and filesystem access failures
- 🧾 CI config errors, bad working directories, and missing build inputs
- 🌿 Git checkout and runner failures
- 🔑 Environment variable problems, invalid secrets, and expired credentials
- 🌐 DNS, TLS, timeout, and connection failures
- 🏗️ Compile, lint, test, and deploy failures

The goal is not to catch everything. It is to reliably catch what is already known and explain it clearly.

## Built on real failures 📚

- 📚 77 bundled playbooks under `playbooks/bundled/`
- 🧪 103 accepted real fixtures in the checked-in regression corpus
- ⚙️ Deterministic ranking, conflict review, and regression gates
- 🤖 Stable terminal, JSON, and workflow output for automation

The current corpus snapshot and validation commands are published in [`docs/fixture-corpus.md`](docs/fixture-corpus.md).
Coverage by failure class and release-to-release proof snapshots are published in [`docs/fixture-corpus.md#coverage-by-failure-class`](docs/fixture-corpus.md#coverage-by-failure-class).
Repository-specific agent operating loops for fixture curation, unmatched-log triage, playbook refinement, and deterministic verification are published in [`docs/agent-workflows.md`](docs/agent-workflows.md).
The refined shipping surface is published in [`docs/release-boundary.md`](docs/release-boundary.md).

## See the checked-in example ⏱️

Want a repo-local proof run? Build the CLI and run the analyze -> workflow handoff on a checked-in sample log:

```bash
make build
./bin/faultline analyze examples/missing-executable.log --format markdown --mode quick
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

- 📄 `examples/missing-executable.workflow.local.txt`
- 🤖 `examples/missing-executable.workflow.agent.json`
- 🔁 `examples/missing-executable.replay.expected.md`
- 🔍 `examples/missing-executable.trace.expected.md`
- 🆚 `examples/missing-vs-runtime.compare.expected.md`

For a concise playbook-driven demo set with different failure classes:

- 🧪 `./bin/faultline analyze examples/missing-executable.log --format markdown --mode quick`
- 🧪 `./bin/faultline analyze examples/runtime-mismatch.log --format markdown --mode quick`
- 🧪 `./bin/faultline analyze examples/docker-auth.log --format markdown --mode quick`

The bundled missing-executable diagnosis snapshot starts like this:

```md
# Required executable or runtime binary missing

- ID: `missing-executable`
- Confidence: 84%
- Category: build
- Severity: high

## Summary

The job tried to launch a required tool or runtime binary, but that executable was missing from the image, runner, or expected path.
```

## Why it exists 🎯

CI failures are often repetitive, noisy, opaque, and slower to diagnose than they should be.

Faultline is built for engineers who want:

- 🎯 deterministic results from explicit rules
- 🔍 evidence pulled directly from the log
- ⚡ fast local diagnosis without uploading build data
- 🤖 stable terminal, JSON, and workflow output for automation

It is intentionally narrow. Faultline does not try to explain every possible failure. It aims to be fast, repeatable, and trustworthy on failures it knows. Designed to minimise false positives: better no result than a wrong one.

## Release Boundary 📦

The default release story is intentionally small:

- 🔎 `analyze`
- 🧭 `workflow`
- 📚 `list`
- 🗂️ `explain`
- 🧯 `fix`

`inspect`, `guard`, and `packs` remain supported companion commands, but they are not the primary onboarding path. Provider-backed GitHub Actions delta remains experimental and hidden behind `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1`.

## Why trust it 🔒

- 🔁 Same input and playbook set produce the same result every time.
- 🔍 Evidence is pulled directly from matched log lines.
- 🧯 Fix steps come from checked-in playbooks, not probabilistic generation.
- ⚖️ `--bayes` never creates new matches; it only reranks already-detected candidates and explains why.
- 🧠 Repo signals only participate when Faultline has explicit repository context (`--git` or guard/inspect). They are additive: they never create matches, only enrich an existing diagnosis with config drift, CI config changes, large-commit blast radius, hotspot files and directories, co-change pairs, hotfix or revert indicators, and CODEOWNERS-derived ownership boundary signals (boundary crossing, upstream component changes, ownership mismatch, failure clustering).
- 🤖 JSON and workflow output stay stable for automation and agent workflows.
- 🏠 Analysis runs locally without shipping build logs to a hosted service.

## What it does ⚙️

- 🔎 Analyze CI logs from a file or stdin.
- 📈 Rerank close calls with `--bayes`.
- 🧭 Explain why the winning diagnosis beat nearby alternatives.
- 🔍 Trace rule-by-rule playbook evaluation with `faultline trace` or `faultline analyze --trace`.
- 🔁 Replay saved analysis artifacts with `faultline replay analysis.json`.
- 🆚 Compare saved analysis artifacts with `faultline compare previous.analysis.json current.analysis.json`.
- 🪟 Focus the live report with `--view summary|evidence|trace|fix|raw`.
- 🪟 Focus replayed analysis artifacts with `--view summary|evidence|fix|raw`.
- 🌡️ Surface likely drift causes only when repo context is explicit: config file changes, CI pipeline edits, large blast-radius commits, and hotspot patterns from recent history.
- 🧩 Turn the top diagnosis into a deterministic workflow handoff.
- 🗂️ Optionally inspect a repository for source-level failure risks.
- 🛡️ Optionally run quiet, high-confidence local checks with `guard`.
- 🖥️ Render concise terminal, markdown, or stable JSON output.
- 📐 Use checked-in playbooks and real-fixture regression gates as the trust boundary.

## Install options 📥

### One-command installer ⚡

This is the default local install path.

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
```

### Build from source 🧱

Requires Go 1.25+.

```bash
git clone https://github.com/faultline-cli/faultline
cd faultline
go build -o faultline ./cmd
./faultline analyze examples/missing-executable.log
```

### Docker 🐳

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/missing-executable.log
```

### Release archive 🗃️

Release archives are published on the GitHub Releases page.

```bash
VERSION=v0.3.0
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

If you move the binary elsewhere, keep `playbooks/bundled/` beside it or set `FAULTLINE_PLAYBOOK_DIR`.

### GitHub Actions follow-up path 🔁

After validating the local CLI flow, run the same deterministic outputs inside a failing workflow job:

```bash
VERSION=v0.3.0 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze build.log --format markdown --ci-annotations > faultline-summary.md
faultline analyze build.log --json --bayes > faultline-analysis.json
faultline workflow build.log --json --mode agent > faultline-workflow.json
```

This is the strongest next step when you want the diagnosis artifacts attached automatically to failed runs without changing the core product path.
The provider integration contract is documented in `docs/github-action-contract.md`.

## First run examples 🧪

The repository includes runnable sample logs and expected markdown output.

```bash
./bin/faultline analyze examples/missing-executable.log
cat examples/missing-executable.log | ./bin/faultline workflow --no-history
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --bayes --no-history
./bin/faultline analyze examples/runtime-mismatch.log
./bin/faultline analyze examples/docker-auth.log
./bin/faultline analyze examples/missing-executable.log --format markdown
./bin/faultline analyze examples/missing-executable.log --json --bayes
./bin/faultline fix examples/missing-executable.log --format markdown
./bin/faultline explain missing-executable
./bin/faultline guard .
```

More runnable examples and output snapshots are documented in `examples/README.md`.

## Command guide 🧰

Stable commands in the default release path:

| Command | Purpose |
| --- | --- |
| `analyze [file]` | Diagnose a CI log from a file or stdin |
| `workflow [file]` | Generate a deterministic follow-up workflow |
| `fix [file]` | Print fix steps for the top diagnosis |
| `list` | List bundled and installed playbooks |
| `explain <id>` | Show the full playbook for one diagnosis |

Supported companion commands:

| Command | Purpose |
| --- | --- |
| `trace [file]` | Show rule-by-rule evaluation and rejection context |
| `replay <analysis.json>` | Re-render a saved analysis artifact deterministically |
| `compare <before> <after>` | Diff two saved analysis artifacts |
| `inspect [path]` | Scan a repository for source-level findings |
| `guard [path]` | Emit only high-confidence local prevention findings |
| `packs` | Install and list optional extra playbook packs |

Useful flags:

| Flag | Description |
| --- | --- |
| `--json` | Emit machine-readable JSON |
| `--format terminal\|markdown\|json` | Choose the output format |
| `--mode quick\|detailed` | Control human-readable output detail |
| `--top N` | Show the top N ranked diagnoses |
| `--bayes` | Apply deterministic Bayesian-inspired reranking |
| `--ci-annotations` | Emit GitHub Actions annotations during analysis |
| `--git` | Enrich analysis with recent local git context (config drift, CI changes, large commits, hotspots, hotfix and revert signals, CODEOWNERS ownership boundary signals) |
| `--repo <path>` | Choose the repository used by `--git` |

Advanced usage:

- 📦 `packs` installs and manages optional extra playbook packs after the bundled catalog is no longer enough.
- 🛡️ `inspect` and `guard` provide bounded local-prevention checks using the same deterministic playbook model.
- 🧪 `FAULTLINE_EXPERIMENTAL_GITHUB_DELTA=1 faultline analyze ... --delta-provider github-actions` enables the hidden experimental provider-backed delta path.

## How it works 🧠

1. Faultline normalizes the input log into stable lines.
2. It loads deterministic YAML playbooks from the bundled catalog and any optional installed packs.
3. It matches explicit patterns, extracts supporting evidence, and ranks results with stable rules.
4. When `--bayes` is enabled, it reranks only the already-matched candidates and adds explainable ranking hints.
5. When repo context is explicit (`--git`), it attaches additive history signals to the diagnosis: recently changed config and dependency files, CI pipeline file edits, hotspot directories and files, co-change pairs, large blast-radius commits, hotfix and revert patterns, and author breadth across the commit window. It also parses CODEOWNERS, builds an ownership graph, and derives topology signals for ownership boundary crossings, upstream component changes, ownership mismatches, and localised failure clustering.
6. It returns a diagnosis, evidence, fix steps, workflow hints, and validation guidance.

The same input and playbook set should produce the same result every time.

## Support matrix 📋

| Capability | Supported |
| --- | --- |
| Local log files | Yes |
| Stdin input | Yes |
| Stable JSON output | Yes |
| Docker usage | Yes |
| CI usage | Yes |
| Local repo inspection | Yes |
| Local guard checks | Yes |
| Network calls during analysis | No by default; experimental provider delta only |

## Credibility checks ✅

- 🧪 `./bin/faultline fixtures stats --class real` currently reports 103 accepted real fixtures and a `weak_match` rate of `0.097` (10/103).
- 📊 The checked-in regression snapshot reports top-1 = 1.000, top-3 = 1.000, unmatched = 0.000, false_positive = 0.000.
- 📚 The bundled catalog currently ships 77 playbooks under `playbooks/bundled/`.
- ✅ Release validation runs `make test`, `make review`, `make fixture-check`, `make cli-smoke`, release archive smoke tests, and Docker smoke tests.

These numbers describe the checked-in regression corpus, not the full space of CI failures.

## Repository guide 🗺️

- 🧪 `examples/README.md` shows runnable sample logs and expected output.
- 📈 `docs/fixture-corpus.md` publishes the checked-in regression snapshot and regeneration commands.
- 🧭 `docs/failures/README.md` indexes search-targeted CI failure pages tied to Faultline diagnoses.
- 🏗️ `docs/architecture.md` explains package boundaries and runtime flow.
- 🔌 `docs/github-action-contract.md` documents the provider-agnostic CLI contract for a thin GitHub Action wrapper.
- 🧩 `docs/playbooks.md` documents playbook authoring and pack composition.
- 📦 `docs/distribution.md` covers release and Docker packaging.
- 📰 `docs/releases/v0.3.0.md` captures the shipped v0.3.0 release notes.
- 📰 `docs/releases/v0.2.0.md` captures the shipped v0.2.0 positioning and release notes.
- 🔍 `docs/detectors.md` describes detector behavior.
- 🗂️ `docs/adr/README.md` indexes architectural decisions.
- 🤝 `CONTRIBUTING.md` covers contribution and fixture-sanitization rules.

## Development 👩‍💻

```bash
make build
make test
make review
make demo-assets
```

`make demo-assets` regenerates the README GIFs and screenshots from the VHS tapes under `docs/readme-assets/tapes/`.

## Feedback 💬

The most useful issue is a sanitized CI failure that Faultline should diagnose better. Have a failure this doesn't catch? Open an issue with the log. Include the smallest log excerpt that reproduces the problem, the expected diagnosis, and what Faultline returned instead.

Raw ingestion artifacts belong in `fixtures/staging/` only as a local review queue. Sanitize them before promotion into `fixtures/real/`.

## License 📄

Faultline is licensed under MIT. See `LICENSE`.
