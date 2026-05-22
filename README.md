# Faultline

[![CI](https://github.com/faultline-cli/faultline/actions/workflows/ci.yml/badge.svg)](https://github.com/faultline-cli/faultline/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/faultline-cli/faultline)](https://github.com/faultline-cli/faultline/releases)
[![Coverage](https://img.shields.io/badge/coverage-87.8%25-brightgreen)](docs/fixture-corpus.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Playbooks](https://img.shields.io/badge/playbooks-182-blue)](docs/failures/catalog/README.md)

Recurring CI failures turn build logs into time sinks: repeated breakages, red herrings, flaky pipelines, and hours lost rediscovering fixes the team already knows. Faultline is a deterministic CLI for diagnosing CI build failures. It matches the log against known failure patterns and returns the failure class, evidence lines, and fix path it can justify — with no AI or LLM call required. If no known pattern matches, it stays quiet. Same log in → same result out.

Faultline is built for teams that want a trustworthy local classifier before deeper investigation starts and a repeatable way to turn CI incidents into shared knowledge:

- Local-first analysis with no LLM, search, issue tracker, or hosted service dependency.
- Evidence copied from the input log so humans and agents can verify the diagnosis.
- Stable text, markdown, JSON, and workflow output for CI steps, tickets, agent handoff, and postmortems.
- Local single-repo recurrence memory for seeing which known failures keep coming back.
- 182 bundled playbooks, 215 accepted real fixtures, and an 89.4% large-corpus GitHub Actions match-rate evaluation.

## Try this in 30 seconds

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
mkdir -p examples
test -f examples/missing-executable.log || curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/examples/missing-executable.log -o examples/missing-executable.log
```

Example log:

```text
exec /__e/node20/bin/node: no such file or directory
```

Analyze it:

```bash
faultline analyze examples/missing-executable.log
```

Explain the matched class:

```bash
faultline explain missing-executable
```

Emit JSON for automation:

```bash
faultline analyze examples/missing-executable.log --json
```

Generate a deterministic handoff for a human, ticket, or agent:

```bash
faultline fix examples/missing-executable.log
faultline workflow examples/missing-executable.log --json --mode agent
```

JSON output includes stable fields like:

```json
{
  "matched": true,
  "faultline_status": "failure",
  "results": [
    {
      "rank": 1,
      "failure_id": "missing-executable",
      "confidence": 0.6,
      "evidence": [
        "exec /__e/node20/bin/node: no such file or directory"
      ]
    }
  ]
}
```

Now try your own log:

```bash
faultline analyze path/to/your-ci.log
faultline analyze path/to/your-ci.log --json
```

## Deterministic First Pass

Faultline is a known-pattern classifier for CI failure logs.

- It classifies failures it recognizes from explicit, versioned patterns.
- It returns matched evidence from the log so the result can be checked.
- It is silent when the failure is unknown or evidence is too weak.
- AI, search, issue trackers, and debugging tools are still useful as the second step, after the log has been classified or left unmatched.

This makes Faultline useful before investigation starts: route the obvious failures, standardize known fixes, and avoid spending attention on classes the team already understands.

## Core Commands

```bash
faultline analyze build.log
cat build.log | faultline analyze --json
faultline batch build-1.log build-2.log --json
faultline fix build.log
faultline workflow build.log --json --mode agent
faultline list
faultline explain missing-executable
```

- `analyze`: classify a failing log and show evidence, diagnosis, and next action.
- `batch`: classify several logs and group the results by failure pattern.
- `fix`: print the remediation steps for the top diagnosis.
- `workflow`: generate deterministic follow-through output for automation or handoff.
- `list`: browse known failure classes.
- `explain`: inspect one failure class before trusting or changing it.

Companion surfaces such as `inspect`, `guard`, `trace`, `replay`, `compare`,
`report`, `history`, and `packs` exist, but they are not
the first-run story. `report` and `history` read only the local forensic store
created by prior local runs; cross-repo recurrence and coordination belong to
the Team layer.

## What to Trust

Faultline output is designed to be inspectable.

- `failure_id` is the stable class name.
- `evidence` is copied from the input log.
- `confidence` reflects rule strength, not certainty about your whole system.
- `faultline explain <failure-id>` shows the diagnosis, fix guidance, and matching rules for that class.
- `--json` returns the same classification in a machine-readable artifact for CI, agents, tickets, or postmortems.

Unknown output is not a failure of the CLI contract. If the log does not match a known class, Faultline should say so instead of inventing a diagnosis.

## What It Catches

Faultline ships with 182 bundled playbooks for common CI failure classes.

| Category | Examples |
| --- | --- |
| Runtime and executables | missing binaries, PATH failures, runtime version mismatches, OOM kills |
| Dependencies | npm/yarn/pnpm lockfile drift, Maven/Gradle restore failures, missing modules, registry errors |
| Containers and infrastructure | Docker auth, image pull failures, entrypoint errors, volume and artifact path issues |
| Test runners | pytest, jest, Go test, cargo test, flaky tests, timeouts, zero-test runs |
| Access and network | permission denied, DNS failures, TLS errors, proxy issues, expired credentials |
| CI workflow | skipped steps, ignored exit codes, missing artifacts, cache misses, shallow checkout, submodules |
| Build tooling | formatting gates, TypeScript and compiler errors, shell compatibility, config schema errors |
| IaC and deploy | Terraform init/state/provider failures, Kubernetes rollout errors, health check failures |

Use `faultline list` to inspect the full catalog and `faultline explain <failure-id>` to see how a class is matched.

## CI Use

Use Faultline after a job fails and you have a log file to classify. The key is
capturing the failing command's output to a file first.

```yaml
- name: Run tests
  shell: bash
  run: |
    set -o pipefail
    make test 2>&1 | tee build.log
```

Then run Faultline in a failure-only step:

```yaml
- name: Diagnose failure
  if: failure()
  uses: faultline-cli/action@v1
  with:
    log: build.log
```

Manual install works anywhere the CLI can read a log:

```yaml
- name: Diagnose failure
  if: failure()
  run: |
    curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
    faultline analyze build.log --json > faultline-analysis.json
    faultline workflow build.log --json --mode agent > faultline-workflow.json
```

See the [GitHub Actions contract](docs/github-action-contract.md) and [GitLab CI contract](docs/gitlab-ci-contract.md) for wrapper details.

## Recurring Failures

`faultline analyze` records small matched-failure metadata in a local forensic
store by default. It does not store full raw logs by default. After repeated
local runs, or CI runs that persist the store between jobs, use:

```bash
faultline report
faultline history
faultline analyze build.log --history
```

- `report` groups stored local runs by failure class so repeated breakages are easier to see.
- `history` shows recurring signatures and quality summaries from the local store.
- `--history` adds explicit local recurrence context to the current analysis output.

This local memory is deliberately single-repo. Cross-repo aggregation, ownership-aware recurrence, and team coordination belong to the Team layer documented in [docs/release-boundary.md](docs/release-boundary.md).

## Team Operating Loop

Faultline Core answers local diagnosis questions: what failed, what evidence supports that classification, and what fix path is known.

For teams with flaky CI, uneven documentation, or too much tribal knowledge, use the output as a shared operating loop:

1. Classify the failure with `faultline analyze build.log`, or group several job logs with `faultline batch`.
2. Use `failure_id` as the stable label in tickets, incident notes, and postmortems.
3. Paste markdown or JSON output into the handoff so the evidence and fix path travel together.
4. Run `faultline workflow build.log --json --mode agent` when a human or agent needs likely files, reproduction steps, verification steps, and remediation context.
5. Use `faultline report` and `faultline history` to pick the repeated classes worth standardizing.
6. When a missed or weak diagnosis keeps recurring, submit a sanitized failure or add a playbook so the next engineer gets the documented path first.

The team questions stay simple:

- What keeps failing?
- Where are we losing time?
- What should we standardize?

## Metrics

Faultline optimizes for high precision over broad coverage. The current checked-in fixture baseline is scoped to the repository corpus, not every possible CI failure.

- Bundled playbooks: 182
- Accepted real fixtures: 215
- Fixture top-1 baseline pass rate: 100% (215/215)
- Fixture false positives: 0
- Weak matches: 0
- Published large-scale GitHub Actions evaluation: 89.4% of 30,094 failed logs matched at least one bundled playbook

The lower playbook count is intentional: this release favors a smaller, cleaner default catalog over keeping low-evidence rules in the main path. These metrics mean the known corpus is reproducible and guarded against regression. The 89.4% figure is a match-rate claim, not a claim that every match was manually judged as the correct diagnosis. These metrics do not mean every new log should match. Silence is intentional when the evidence is unknown, ambiguous, or below the classifier threshold.

Faultline is most useful when a team already sees recurring CI failures and wants standard classifications, repeatable fix paths, and machine-readable artifacts without adding runtime network calls to the analysis path.

Details: [docs/fixture-corpus.md](docs/fixture-corpus.md).

## Release Notes

v0.4.9 expands the source detector catalog from 7 to 12 playbooks. New source playbooks cover hardcoded secrets (`hardcoded-secret`), goroutine leaks (`goroutine-leak`), disabled TLS verification (`insecure-tls-skip-verify`), missing transaction rollbacks (`missing-transaction-rollback`), and HTTP clients without timeouts (`http-client-no-timeout`). Each adds positive and safe fixtures. The 5 pre-existing source playbooks are tightened with improved scoring, compound signal logic, and fix guidance. `make review` passes at 294 classified conflict patterns.

v0.4.8 promotes `faultline batch` into the core release path for local
multi-log diagnosis. Use it when one workflow produces several job logs and you
want the same deterministic classification grouped across the set. `report`,
`history`, `trace`, and the other advanced surfaces remain companion commands.

v0.4.7 is a hardening release for the existing local-first product. It does
not add a new first-run surface; it tightens the release gates and makes local
recurrence, metrics, and generated docs easier to trust.

- Added generated failure-doc validation and Bayes regression checks to the
  release gate so stale docs or ranking drift block release cuts.
- Kept normal `analyze` and `workflow` output stable unless local history or
  explicit metrics history is requested.
- Added hidden advanced metrics-history input for deterministic reliability
  metrics and advisory policy hints without adding new public JSON fields.
- Expanded CLI smoke coverage for local `report`, workflow metrics hints, and
  bundled-plus-installed pack provenance ordering.
- Backfilled v0.4.6 release notes and corrected fixture-corpus drift so the
  repository proof matches the current checked-in baseline.

v0.4.6 keeps the default story narrow: classify the failed log, show the evidence, and hand off the known fix path. The release favors fewer, stronger defaults over broad but weak inference.

- Removed 20 low-signal or overly narrow playbooks from the default bundle, including project-specific test-suite rules, weak inference rules, and absence-of-run workflow variants.
- Regenerated the failure catalog from the tightened playbook set and added checks so stale generated docs are caught instead of drifting after a playbook is removed.
- Promoted 4 additional real fixtures, bringing the checked-in real corpus to 215 accepted failures with 100% top-1 and top-3 baseline pass rates.
- Reduced overlap noise in the bundled catalog; `make review` now passes against 260 classified conflict patterns.
- Kept specialized, provider-specific, and maintainer-only work out of the first-run story unless it has deterministic tests, fixture evidence, and release-grade verification.

## Submit a Failure

The most useful contribution is a real CI failure that Faultline missed, misclassified, or explained weakly.

Open a [missed failure issue](https://github.com/faultline-cli/faultline/issues/new?template=missed_failure.md) with:

- the CI provider
- the smallest sanitized log excerpt that reproduces the result
- the diagnosis or fix path you expected
- the exact Faultline output
- optional reproduction steps or a public build link when safe to share

Do not include secrets, tokens, signed URLs, private hostnames, customer names, or unsanitized internal repository data.

For code changes:

```bash
make build
make test
make fixture-check
make review
make docs-check
make cli-smoke
```

Use [docs/playbooks.md](docs/playbooks.md) for playbook authoring and [docs/fixture-corpus.md](docs/fixture-corpus.md) for corpus expectations.

## More Examples

```bash
faultline analyze examples/missing-executable.log
faultline analyze examples/runtime-mismatch.log
faultline analyze examples/lockfile-drift.log
faultline analyze examples/docker-auth.log
```

See [examples/README.md](examples/README.md) for expected outputs and checked-in snapshots.

## Install Options

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
VERSION=v0.4.8
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

## FAQ

**Does Faultline send my logs anywhere?**
No. Analysis runs entirely on your machine or inside your CI runner. No log data, telemetry, or output is transmitted to any external service. The only network activity is the optional `--git` flag, which reads your local `.git` directory — not the internet.

**What CI providers does it support?**
Any provider that produces a text log file: GitHub Actions, GitLab CI, CircleCI, Jenkins, Buildkite, Drone, Bitbucket Pipelines, and self-hosted runners. Faultline reads the log file — it has no provider-specific API dependency.

**Does it work with AI coding agents?**
Yes. `--json` and `--mode agent` output are designed for agent handoff. The `workflow` command returns structured diagnosis, likely files, reproduction steps, and remediation context in a stable JSON shape that agents can consume directly.

**How does it differ from just grepping the log?**
Faultline applies scored evidence rules against the full log, not a single line match. It normalizes multi-line patterns, ranks competing hypotheses, filters low-confidence matches, and returns a stable `failure_id` that travels cleanly into tickets, postmortems, and automation. `faultline explain <failure-id>` shows the full reasoning behind the classification.

**What if my failure isn't recognized?**
Faultline stays quiet rather than guessing. If you see a recurring unmatched failure, open a [missed failure issue](https://github.com/faultline-cli/faultline/issues/new?template=missed_failure.md) with a sanitized log excerpt. The playbook authoring guide is in [docs/playbooks.md](docs/playbooks.md).

**Can I add custom failure patterns for my team?**
Yes, via packs. A pack is a directory of YAML playbooks that Faultline loads alongside the bundled set. See `faultline packs` and [docs/release-boundary.md](docs/release-boundary.md) for the layering model.

**How do I integrate it in a GitHub Actions workflow?**
Use the [`faultline-cli/action`](https://github.com/faultline-cli/action) wrapper on a failure-only step, or install the binary directly. Full contract in [docs/github-action-contract.md](docs/github-action-contract.md).

**What is the performance overhead?**
Faultline is a statically compiled Go binary with no runtime dependencies. Typical analysis of a 10 000-line CI log completes in well under one second.

## License

MIT. See [LICENSE](LICENSE).
