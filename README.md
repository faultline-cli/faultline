# Faultline

Faultline is a deterministic CLI for CI failure diagnosis.

It reads a build log or repository tree, matches the input against a curated
playbook library, and returns the most likely failure with evidence pulled
directly from the input plus concrete remediation steps. Same input, same
playbooks, same output.

## What Faultline Does

- Diagnose failed CI jobs from logs or stdin.
- Explain the likely root cause in plain language.
- Show exact evidence lines instead of fuzzy summaries.
- Suggest concrete follow-up and validation steps.
- Inspect source trees with deterministic source-detector playbooks.
- Emit stable text, markdown, JSON, and workflow output.

## Why It Is Different

Faultline is intentionally narrow.

- Deterministic runtime only. No ML or LLM calls in product logic.
- Evidence-first matching grounded in exact lines or source signals.
- Stable ranking and output suitable for automation.
- Installable playbook packs instead of hidden server-side logic.

This repository is the starter product. It ships with a strong bundled catalog
for common CI failures and supports deeper premium packs on top.

## Quick Example

Instead of reading thousands of log lines to find the one failure that matters,
Faultline returns the top diagnosis, the evidence that triggered it, and the
shortest credible fix path.

```bash
$ faultline analyze build.log
Top match: docker-auth
Confidence: 0.96

Evidence:
  - denied: requested access to the resource is denied
  - unauthorized: authentication required

Likely cause:
  The registry token is missing, expired, or scoped to a different registry.

Fix:
  1. Run docker login against the target registry.
  2. Confirm the CI secret has pull or push scope for the repository.
  3. Re-run the failing image pull or push step.
```

Source inspection uses the same deterministic model:

```bash
$ faultline inspect .
Top match: panic-in-http-handler
```

## Installation

### Release archive

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

Release archives contain:

```text
faultline_<version>_<os>_<arch>/
  faultline
  playbooks/
    bundled/
  README.md
```

If you move the binary elsewhere, keep `playbooks/bundled/` beside it or set
`FAULTLINE_PLAYBOOK_DIR`.

### Build from source

```bash
make build
./bin/faultline analyze build.log
```

### Docker

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/build.log
```

## Quickstart

```bash
# Analyze a log file
faultline analyze build.log

# Read from stdin
cat build.log | faultline analyze

# Stable JSON for automation
faultline analyze build.log --json

# Markdown for issue templates or handoff notes
faultline analyze build.log --format markdown

# Add recent git context
faultline analyze build.log --git --since 30d --repo .

# Show only the fix steps for the top diagnosis
faultline fix build.log

# Inspect a repository for source-risk findings
faultline inspect .

# Build a deterministic follow-up workflow
faultline workflow build.log --mode agent --git --repo .
```

## Core Commands

| Command | Description |
| --- | --- |
| `analyze [file]` | Ranked diagnosis from a file or stdin |
| `inspect [path]` | Ranked source-risk findings from a repository tree |
| `fix [file]` | Print only the fix steps for the top diagnosis |
| `workflow [file]` | Build a deterministic local or agent-ready follow-up plan |
| `list` | List available playbooks from the active catalog |
| `explain <id>` | Show full detail for one playbook |
| `packs install <dir>` | Install an extra playbook pack into `~/.faultline/packs/` |
| `packs list` | List locally installed extra packs |

Common flags:

| Flag | Default | Description |
| --- | --- | --- |
| `--top N` | `1` | Show top N ranked results |
| `--mode quick\|detailed` | `quick` | Human output verbosity |
| `--format raw\|markdown` | `raw` | Human-readable output shape |
| `--json` | `false` | Emit stable JSON |
| `--ci-annotations` | `false` | Emit GitHub Actions warning lines |
| `--playbooks <dir>` | auto | Replace the active catalog with one directory |
| `--playbook-pack <dir>` | none | Compose one or more extra pack roots on top of the starter catalog |
| `--git` | `false` | Enrich analysis with recent local git context |
| `--since <window>` | `30d` | History window for `--git` |
| `--repo <path>` | `.` | Repository path used by `--git` |

Environment variables:

| Variable | Description |
| --- | --- |
| `FAULTLINE_PLAYBOOK_DIR` | Full catalog override |
| `FAULTLINE_PLAYBOOK_PACKS` | Additional pack roots separated by the platform path separator |
| `FAULTLINE_PLAIN=1` | Force plain text output |
| `NO_COLOR` | Disable ANSI colors |

## How Detection Works

Faultline keeps the runtime explicit and reviewable:

1. Read log input from a file path or stdin, or walk a repository tree for
   source inspection.
2. Normalize the input into stable lines or source signals.
3. Load the bundled starter catalog plus any installed or explicitly composed
   extra packs.
4. Validate playbook structure and review overlap conflicts deterministically.
5. Match and rank with explicit rules, not hidden heuristics.
6. Render concise text, markdown, JSON, or workflow output.

The authoritative product constraints live in [SYSTEM.md](SYSTEM.md).

## Bundled Playbooks vs Premium Playbooks

Faultline now ships with an exact split.

| Package | Count | What it covers |
| --- | --- | --- |
| Bundled starter | 60 playbooks | Common auth, network, generic build, generic runtime, generic test, Node.js, TypeScript, Python, Go, core container and Kubernetes failures, starter source-detector coverage |
| Premium repository | 83 playbooks | Deeper JVM, Ruby, Rust, and .NET build stacks, provider-specific CI workflows, Terraform and Helm workflows, richer test-runner coverage, CORS, database pool exhaustion, Node.js runtime debugging, advanced source detectors |

Coverage matrix:

| Area | Bundled starter | Premium upgrade | Why the split works |
| --- | --- | --- | --- |
| Auth and secrets | Core registry, git, SSH, env, and cloud credential failures | Provider- and platform-specific auth remains in premium packs | Starter handles the failures most teams hit first; premium adds environment-specific depth |
| Build and dependency failures | Common lockfile, compiler, install, syntax, runtime-version, and quality-gate issues across Node.js, TypeScript, Python, and Go | JVM, Ruby, Rust, and .NET ecosystem playbooks such as Gradle, Maven, RuboCop, Bundler, Cargo, and .NET build failures | Starter is broad and immediately useful; premium adds stack-specific expertise with denser remediation |
| CI and runner behavior | Pipeline timeout, runner disk exhaustion, artifact transfer, and secret availability | CI-provider-specific pipeline rules live in the premium repo | Starter covers universal CI pain points; premium covers vendor-specific workflows |
| Deploy and infrastructure | Config mismatch, container crash, image pull failure, health check failure, CrashLoopBackOff, and deploy-time port conflicts | Helm and Terraform workflows, plus deeper platform operations rules | Starter gives a real deploy aha moment; premium adds higher-leverage ops workflows |
| Network | Connection refused, DNS, SSL, egress blocking, timeouts, and rate limits | CORS, gRPC, proxy, and webhook-specialized rules | Starter handles broad transport failures; premium covers application-edge and platform nuances |
| Runtime | OOM, permissions, env vars, generic disk pressure, port binding, segfaults, and resource limits | Database connection pool exhaustion and Node.js unhandled rejection debugging | Starter captures the common failures; premium covers service-level debugging depth |
| Test | Timeout, snapshot mismatch, flaky behavior, order and parallelism conflicts, fixture gaps, database isolation, coverage gate, and Go data races | Jest, Vitest, pytest, and RSpec runner-specialized rules | Starter covers generic test failure classes; premium accelerates framework-specific diagnosis |
| Source inspection | Two high-signal starter rules for panic handling and dropped errors | Deeper security and concurrency source rules live in the premium repo | Starter proves `inspect` is useful; premium turns it into a deeper code-risk product |

Starter is meant to be genuinely useful on day one. Premium is designed to be a
meaningful upgrade, not a ransom layer.

Starter keeps the highest-frequency and broadest-appeal failures, including:

- auth and secret failures
- lockfile and dependency drift failures
- common Node.js, TypeScript, Python, and Go failures
- pipeline timeouts, disk exhaustion, permission errors, and OOM kills
- connection refused, DNS, SSL, rate limits, and network timeouts
- image pull failures and CrashLoopBackOff in Kubernetes
- test timeouts, snapshot mismatches, flaky tests, and data races
- two starter source-detector playbooks for `inspect`

Premium adds the denser and more specialized playbooks, including:

- `cargo-build`, `dotnet-build`, `gradle-build`, `java-version-mismatch`,
  `maven-dependency-resolution`, `rubocop-failure`, `ruby-bundler`
- `helm-chart-failure`, `terraform-init`, `terraform-state-lock`
- `jest-worker-crash`, `pytest-fixture-error`, `rspec-failure`,
  `vitest-failure`
- `cors-error`, `database-connection-pool-exhausted`,
  `nodejs-unhandled-rejection`

This commercialization pass specifically moved 17 formerly bundled playbooks
into the sister premium repository. The premium repository already contained
additional provider-specific CI, deploy, runtime, and source-detector rules,
which is why its total count is higher than the starter pack by raw count.

The exact scoring and boundary review are documented in
[docs/playbook-packaging-review.md](docs/playbook-packaging-review.md) and
[docs/playbook-matrix.csv](docs/playbook-matrix.csv).

## Install Or Upgrade A Premium Pack

Recommended customer flow:

```bash
git clone <private-premium-pack-repo> ../faultline-premium
faultline packs install ../faultline-premium
faultline packs list
faultline list
```

Update flow:

```bash
cd ../faultline-premium && git pull
faultline packs install --force ../faultline-premium
```

One-off composition without installation:

```bash
faultline analyze build.log --playbook-pack ../faultline-premium
```

For local repository validation, this repo can use the symlinked sister pack
root when present:

```bash
faultline analyze build.log --playbook-pack ./playbooks/packs/premium-local
faultline packs install ./playbooks/packs/premium-local
```

If the symlink is not present locally, point Faultline at the sibling repo
directly instead:

```bash
faultline analyze build.log --playbook-pack ../faultline-premium
faultline packs install ../faultline-premium
```

Use `--playbooks <dir>` only for full catalog overrides. It replaces the starter
catalog instead of composing with it.

## Who It Is For

- Engineers who want a fast explanation for failed CI jobs without handing logs
  to a hosted service.
- Platform and developer-experience teams that want deterministic diagnostics in
  local runs, CI, and Docker.
- Automation or agent workflows that need stable JSON and reproducible ranking.

## Example Use Cases

- A failed `npm ci` job where the lockfile drifted from `package.json`.
- A deploy that pulled the image but entered CrashLoopBackOff on startup.
- A private dependency fetch that failed because the token or SSH key was not
  available to the runner.
- A source inspection pass that flags panic-prone request handlers before the
  next release.

## Output Modes And Integrations

Faultline auto-detects terminal output. Interactive terminals get styled output.
Redirected output, CI, `NO_COLOR`, and `FAULTLINE_PLAIN=1` fall back to readable
plain text. `--json` always bypasses terminal styling.

- `--json` is stable for automation.
- `--format markdown` is useful for issue templates, handoff notes, and agent
  workflows.
- `workflow` produces deterministic next-step plans.
- `--ci-annotations` emits GitHub Actions warnings for inline CI feedback.

## Why Deterministic Matters

CI failure tooling is most useful when it can be trusted under pressure.
Deterministic behavior gives Faultline a stable contract:

- the same log produces the same diagnosis every time
- evidence lines always come from the input that was analyzed
- ranking can be reviewed and improved with tests instead of prompt tuning
- JSON output stays safe for scripts and higher-level automation

That constraint keeps the product honest. Faultline is a diagnostic CLI, not a
chat wrapper around your logs.

## Roadmap And Extension Points

The main extension point is the playbook pack model.

- starter playbooks stay bundled under `playbooks/bundled/`
- extra packs install under `~/.faultline/packs/`
- pack composition works the same locally and in Docker
- the binary stays independent from premium-pack release cadence

See [ROADMAP.md](ROADMAP.md), [docs/architecture.md](docs/architecture.md), and
[docs/distribution.md](docs/distribution.md) for the broader plan.

## Repository Map

```text
cmd/                  CLI entrypoints and maintenance commands
internal/app/         Command-level application services
internal/cli/         Cobra command tree
internal/engine/      Analysis orchestration
internal/detectors/   Log and source detector implementations
internal/playbooks/   Catalog loading, validation, composition, review
internal/matcher/     Evidence extraction and scoring
internal/output/      JSON and command-facing output assembly
internal/renderer/    Human rendering for TTY, plain text, and markdown
internal/repo/        Local git context and correlation
internal/workflow/    Deterministic follow-up workflows
playbooks/bundled/    Starter playbook catalog
packs/premium-local/  Reference premium pack root
docs/                 Architecture, authoring, distribution, and ADRs
```

## Documentation

- [SYSTEM.md](SYSTEM.md): authoritative product and architecture constraints
- [docs/architecture.md](docs/architecture.md): runtime boundaries and pack resolution
- [docs/playbooks.md](docs/playbooks.md): playbook authoring guidance
- [docs/playbook-packaging-review.md](docs/playbook-packaging-review.md): exact starter versus premium split
- [docs/playbook-matrix.csv](docs/playbook-matrix.csv): scored per-playbook inventory
- [docs/distribution.md](docs/distribution.md): starter and premium delivery model
- [docs/adr/README.md](docs/adr/README.md): architecture decision records

## Development And Validation

```bash
make build
make test
make review
make premium-check PREMIUM_PACK_DIR=./packs/premium-local
make premium-review PREMIUM_PACK_DIR=./packs/premium-local
make release-check VERSION=v0.1.0 PREMIUM_PACK_DIR=./packs/premium-local
```

`make premium-link` creates the ignored local symlink at
`playbooks/packs/premium-local` for a locally checked out private premium
repository.

## Commercial Note

The starter catalog in this repository is intended to be complete enough to earn
trust quickly. Premium packs are installable add-ons delivered as separate pack
roots, typically from a private repository or release archive, using the same
deterministic composition model the CLI already supports.
