# Faultline

CI failures waste time because the log is noisy and the diagnosis usually is not.

Faultline is a deterministic diagnosis layer for CI systems.
It reads failing CI logs, matches them against explicit, versioned, checked-in playbooks, and returns evidence-backed diagnoses with stable output for humans and automation.
Same log in, same result out. No guesswork. No AI.

- 🔁 Determinism is the contract: same log and playbook set, same result
- 🔍 Evidence-backed diagnoses: matched lines, not generated summaries
- 📦 First-class failure artifacts: replayable, comparable, and handoff-ready
- 🏠 Local-first by default: your logs stay on your machine
- 🤖 Automation-friendly output: stable JSON and workflow artifacts

Product boundary:

- Core (free): "What failed?"
- Team (paid): "What keeps failing, who owns it, and what do we do about it?"

```text
# input: ci.log
exec /__e/node20/bin/node: no such file or directory
```

```text
$ faultline analyze ci.log

[1] missing-executable (confidence: 84%)
Evidence:
  - exec /__e/node20/bin/node: no such file or directory

Fix:
  - Install the missing runtime in the CI image
  - Pin the runner to an image that includes the expected binary
```

Run `faultline analyze ci.log --json` for the same diagnosis in stable JSON. Same log and playbook set in, same result out.

Built on 106 bundled playbooks and 139 accepted real fixtures with perfect top‑1 and top‑3 match rates.

## Try it now 🚀

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
faultline workflow ci.log --json --mode agent
```

Default path: diagnose the failing log, then turn the winning diagnosis into a deterministic handoff artifact. Use `faultline analyze ci.log --json` for stable machine-readable output.

Output modes:

- 👩‍💻 `faultline analyze ci.log` for a human-readable diagnosis with evidence and fix steps
- 🤖 `faultline analyze ci.log --json` for stable diagnosis JSON
- 🧭 `faultline workflow ci.log --json --mode agent` for a stable handoff artifact
- 🛠 `faultline workflow explain ci.log` and `faultline workflow apply ci.log --dry-run` for the typed remediation workflow layer

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

Determinism is the point, not a feature flag. The same matched evidence and playbook set produce the same diagnosis and artifacts.

`--bayes` is optional and assistive: it reranks already-matched candidates and explains the ranking, but it never creates new matches.

## Why trust it ✅

- 🔁 Same log and playbook set produce the same diagnosis and artifacts every time.
- 🔍 Every diagnosis is backed by matched log lines.
- 📚 Diagnoses and fix steps come from explicit, versioned, inspectable playbooks checked into the repo.
- 🛠 Teams can extend those playbooks without turning diagnosis into a black box.
- 🧪 The shipped catalog is backed by 101 bundled playbooks and 139 accepted real fixtures with perfect top‑1 and top‑3 match rates.
- 🏠 Faultline runs locally by default, so build logs stay on your machine unless you choose otherwise.

Paid value is intentionally tied to coordination across runs and repositories:
correlation, aggregation, policies, shared playbooks, reporting, and sync
belong to the Team layer rather than the free local diagnosis path.

Some companion commands are supported but not part of the first-run story, and provider-backed delta remains experimental. The current boundary is documented in [docs/release-boundary.md](docs/release-boundary.md).

## Core workflow 🧭

The default path is small:

- `faultline analyze <logfile>` diagnoses a failing log from a file or stdin.
- `faultline workflow <logfile>` turns the winning diagnosis into a deterministic follow-up plan.
- `faultline list` and `faultline explain <id>` let you inspect the bundled, versioned catalog.
- `faultline fix <logfile>` prints the remediation steps for the top diagnosis.

For a fast local run:

```bash
faultline analyze ci.log
faultline analyze ci.log --json
faultline workflow ci.log --json --mode agent
```

Typical flow:

1. Run `faultline analyze` on the failing log.
2. Check the matched evidence lines and the top remediation steps.
3. Run `faultline workflow` if you want a structured handoff for an agent or automation step.
4. Use `list`, `explain`, or `fix` when you want to inspect the catalog or narrow in on one diagnosis.

## Common failure classes 🛠️

- 🧰 Missing executables, PATH failures, and command invocation errors
- 🧬 Runtime version mismatches across Node, Python, Ruby, and Go
- 📦 Dependency install, resolver, and lockfile drift failures
- 🐳 Docker, registry authentication, and image configuration failures
- 🔐 Permission, filesystem, and working-directory errors
- 🌐 DNS, TLS, timeout, and other network failures

Faultline is intentionally narrow: better on failures it knows than vague on everything else.

## CI integration 🔁

The same commands run in CI unchanged:

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
      - name: Analyze failure with Faultline
        if: failure()
        run: |
          VERSION=v0.3.1 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
          faultline analyze build.log --json > faultline-analysis.json
          faultline workflow build.log --json --mode agent > faultline-workflow.json
```

That preserves the same deterministic CLI contract in CI that you use locally.

## Examples and snapshots 🧪

The repository includes runnable sample logs and checked-in outputs if you want a quick proof run without bringing your own CI log yet:

```bash
./bin/faultline analyze examples/missing-executable.log
cat examples/missing-executable.log | ./bin/faultline workflow
./bin/faultline analyze examples/runtime-mismatch.log
./bin/faultline analyze examples/docker-auth.log
```

The demo below shows the same default flow on the checked-in `missing-executable` example:

![Faultline missing executable demo](docs/readme-assets/missing-executable.gif)

That gives you three different failure classes to inspect right away:

- `missing-executable` for missing runtime or binary failures
- `runtime-mismatch` for toolchain/version drift failures
- `docker-auth` for container registry authentication failures

More sample inputs and expected outputs live in [examples/README.md](examples/README.md).

## More commands 🧩

These companion commands are supported and documented, but they stay out of the default onboarding path on purpose:

- `trace` shows rule-by-rule evaluation and rejection context.
- `replay` and `compare` re-render or diff saved analysis artifacts deterministically.
- `inspect` and `guard` cover repository-local prevention and high-confidence checks.
- `packs` installs and lists optional extra playbook packs.

Experimental provider-backed delta also exists behind explicit opt-in. The current shipping boundary is documented in [docs/release-boundary.md](docs/release-boundary.md).

## More install options 📦

If you want something other than the one-command installer, Faultline also supports source builds, Docker, and release archives.

<details>
<summary>Build from source, use Docker, or download a release archive</summary>

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

Release archives are also published on the GitHub Releases page:

```bash
VERSION=v0.3.1
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

</details>

## Learn more 📚

- [examples/README.md](examples/README.md) for runnable sample logs and checked-in output snapshots
- [ROADMAP.md](ROADMAP.md) for the current v0.4 delivery order and product direction
- [docs/github-action-contract.md](docs/github-action-contract.md) for the GitHub Actions wrapper contract
- [docs/gitlab-ci-contract.md](docs/gitlab-ci-contract.md) for the GitLab CI wrapper contract
- [docs/fixture-corpus.md](docs/fixture-corpus.md) for regression corpus details and coverage snapshots
- [docs/playbooks.md](docs/playbooks.md) for playbook authoring, team extensions, and pack composition
- [docs/release-boundary.md](docs/release-boundary.md) for the locked Core vs Team boundary

## Development 👩‍💻

```bash
make build
make test
make review
make demo-assets
```

`make demo-assets` regenerates the README demo assets from the VHS tapes under `docs/readme-assets/tapes/`.

## Feedback 💬

The most useful issue is a sanitized CI failure that Faultline should diagnose better. Include the smallest log excerpt that reproduces the problem, the expected diagnosis, and what Faultline returned instead.

Raw ingestion artifacts belong in `fixtures/staging/` only as a local review queue. Sanitize them before promotion into `fixtures/real/`.

## License 📄

Faultline is licensed under MIT. See `LICENSE`.
