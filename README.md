# Faultline

[![CI](https://github.com/faultline-cli/faultline/actions/workflows/ci.yml/badge.svg)](https://github.com/faultline-cli/faultline/actions/workflows/ci.yml) [![142 playbooks](https://img.shields.io/badge/playbooks-142-blue)](docs/playbooks.md) [![top-1 accuracy](https://img.shields.io/badge/top--1_accuracy-100%25-brightgreen)](docs/fixture-corpus.md) [![174 real fixtures](https://img.shields.io/badge/real_fixtures-174-informational)](docs/fixture-corpus.md) [![corpus coverage](https://img.shields.io/badge/corpus_coverage-7.3%25-lightgrey)](eval-work/coverage.md)

Stop spelunking CI logs. Point Faultline at the failure and get the diagnosis.

Faultline is a deterministic diagnosis engine for CI failures. It matches your failing build log against an explicit, versioned catalog of 142 playbooks and returns evidence-backed diagnoses — the exact matched lines, the root cause, and the fix. No AI. No guesswork. Same log in, same result out.

**Your build just failed. Here's what the next 30 seconds looks like:**

```text
# build.log contains:
exec /__e/node20/bin/node: no such file or directory
```

```text
$ faultline analyze build.log

[1] missing-executable (confidence: 84%)
Evidence:
  - exec /__e/node20/bin/node: no such file or directory

Fix:
  - Install the missing runtime in the CI image
  - Pin the runner to an image that includes the expected binary
```

No digging through 2,000 lines of output. No asking an LLM to guess.
The diagnosis is backed by matched evidence, sourced from an inspectable playbook, and stable enough to pipe into automation.

**v0.4.0+** — 142 bundled playbooks · 174 real fixtures · top-1: 1.000 · top-3: 1.000 · unmatched: 0.000 · false-positive: 0.000

## ⚡ Install

One command. Works locally and in CI.

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze build.log
```

## ⚙ How it works

1. **Analyze** — match the failing log against 142 bundled playbooks, extract evidence lines, score and rank candidates
2. **Diagnose** — return the top match with confidence, the exact evidence, and concrete fix steps
3. **Handoff** — optionally emit a stable JSON artifact for your automation, agent, or post-mortem tool

```bash
faultline analyze build.log                    # human-readable: evidence, root cause, fix
faultline analyze build.log --json             # same diagnosis as stable machine-readable JSON
faultline workflow build.log --json --mode agent  # typed remediation artifact for automation
faultline list                                 # browse the full versioned catalog
faultline explain <failure-id>                 # deep-dive on a single failure pattern
faultline fix build.log                        # print remediation steps, nothing else
```

Determinism is the contract, not a feature flag. The same log and playbook set produce the same output every time — which means you can diff it, store it, replay it, and build on top of it.

## 🔍 What it catches

142 playbooks covering the failures that actually break builds in production CI:

| Category | Examples |
|---|---|
| ⚙ Runtime & executables | Missing binaries, PATH failures, node/python/ruby/go version mismatches |
| 📦 Dependencies | npm/yarn/pnpm lockfile drift, Maven/Gradle resolution, dotnet restore |
| 🏗 Infrastructure | Docker auth, registry errors, kubectl auth, OIDC token failures |
| 🧪 Test runners | pytest fixture errors, jest worker crashes, testcontainer startup failures |
| 🔒 Access & network | Permission denied, DNS failures, TLS errors, timeouts |
| 🌐 IaC | terraform init, state lock, provider auth |

Faultline is intentionally narrow: precise on failures it knows, silent on failures it doesn't. No hallucinated diagnoses.

## ↪ Drop it into CI

Add a single step to your failure path. The CLI contract is identical in CI and locally.

```yaml
- name: Diagnose failure
  if: failure()
  run: |
    VERSION=v0.4.0 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
    faultline analyze build.log --json > faultline-analysis.json
    faultline workflow build.log --json --mode agent > faultline-workflow.json
```

The JSON artifacts are stable across runs — safe to store, diff, and feed into downstream automation.
See the [GitHub Actions contract](docs/github-action-contract.md) and [GitLab CI contract](docs/gitlab-ci-contract.md) for full wrapper details.

## → Automation handoff

`faultline workflow` turns the winning diagnosis into a typed, structured artifact — ready to hand off to a remediation agent, feed into a ticket, or attach to a post-mortem.

```json
{
  "schema_version": "workflow.v1",
  "mode": "agent",
  "failure_id": "missing-executable",
  "evidence": [
    "exec /__e/node20/bin/node: no such file or directory"
  ],
  "files": [
    "Dockerfile",
    ".github/workflows/ci.yml"
  ]
}
```

## ◆ What's new in v0.4.0

**Biggest catalog expansion yet** and the sharpest accuracy numbers to date.

- **123 bundled playbooks**, up 60% from 77 in v0.3.0. New coverage: `gradle-build`, `maven-dependency-resolution`, `dotnet-restore`, `terraform-init`, `terraform-state-lock`, `kubectl-auth`, `oidc-token-failure`, `testcontainer-startup`, `pytest-fixture-error`, `jest-worker-crash`
- **Fewer false triggers.** Weak-match rate dropped from 11.5% → 5.7% via targeted exclusion fixes. Five playbooks that were firing on unrelated log patterns no longer do.
- **174 accepted real fixtures** with zero unmatched and zero false positives — validated against real CI logs from GitHub, GitLab, Reddit, Discourse, and Stack Exchange

Full release notes: [docs/releases/v0.4.0.md](docs/releases/v0.4.0.md)

## ◈ Free vs Team

**Core (free):** everything you need to diagnose failures fast, locally, with your logs staying on your machine.
`analyze` · `workflow` · `list` · `explain` · `fix`

**Team (paid):** built for orgs that want to track failure patterns over time.
Cross-run correlation, failure aggregation, policies, shared playbook repos, and reporting across teams.

Companion surfaces (`inspect`, `guard`, `trace`, `replay`, `compare`, `packs`) are supported but non-default. See [docs/release-boundary.md](docs/release-boundary.md).

## ▶ Try the examples

The repo ships with real failure logs and checked-in expected outputs. No CI log needed to kick the tires.

```bash
./bin/faultline analyze examples/missing-executable.log
./bin/faultline analyze examples/runtime-mismatch.log
./bin/faultline analyze examples/docker-auth.log
```

![Faultline demo](docs/readme-assets/missing-executable.gif)

More samples and expected outputs: [examples/README.md](examples/README.md)

## More install options

<details>
<summary>Build from source, Docker, or release archive</summary>

Requires Go 1.25+ for source builds.

```bash
git clone https://github.com/faultline-cli/faultline
cd faultline
go build -o faultline ./cmd
./faultline analyze examples/missing-executable.log
```

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/missing-executable.log
```

```bash
VERSION=v0.4.0
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

</details>

## 📚 Learn more

- [examples/README.md](examples/README.md) — runnable sample logs and expected output snapshots
- [docs/playbooks.md](docs/playbooks.md) — authoring playbooks, team extensions, and packs
- [docs/fixture-corpus.md](docs/fixture-corpus.md) — regression corpus and accuracy methodology
- [ROADMAP.md](ROADMAP.md) — what's coming next
- [docs/release-boundary.md](docs/release-boundary.md) — Core vs Team boundary detail

## 🛠 Development

```bash
make build
make test
make review
make cli-smoke
```

## 💬 Feedback

The most useful issue is a sanitized CI failure that Faultline should diagnose better. Include the smallest log excerpt that reproduces the problem, the expected diagnosis, and what Faultline returned instead.

## License

MIT. See `LICENSE`.
