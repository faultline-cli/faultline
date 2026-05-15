# Faultline

Recurring CI failures turn build logs into time sinks: repeated breakages, red herrings, flaky pipelines, and hours lost proving what did not cause the failure. Faultline is a deterministic CLI for the first pass over a failed CI log. It matches the log against known failure patterns and returns the failure class, evidence lines, and fix path it can justify. If no known pattern matches, it stays quiet. Same log in -> same result out.

Faultline is built for teams that want a trustworthy local classifier before deeper investigation starts:

- Local-first analysis with no LLM, search, issue tracker, or hosted service dependency.
- Evidence copied from the input log so humans and agents can verify the diagnosis.
- Stable text and JSON output for CI steps, tickets, agent handoff, and postmortems.
- 173 bundled playbooks, 215 accepted real fixtures, and a published 89.4% large-corpus GitHub Actions match evaluation.

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

## What to Trust

Faultline output is designed to be inspectable.

- `failure_id` is the stable class name.
- `evidence` is copied from the input log.
- `confidence` reflects rule strength, not certainty about your whole system.
- `faultline explain <failure-id>` shows the diagnosis, fix guidance, and matching rules for that class.
- `--json` returns the same classification in a machine-readable artifact for CI, agents, tickets, or postmortems.

Unknown output is not a failure of the CLI contract. If the log does not match a known class, Faultline should say so instead of inventing a diagnosis.

## v0.4.7 Release

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

## v0.4.6 Release

v0.4.6 keeps the default story narrow: classify the failed log, show the evidence, and hand off the known fix path. The release favors fewer, stronger defaults over broad but weak inference.

- Removed 20 low-signal or overly narrow playbooks from the default bundle, including project-specific test-suite rules, weak inference rules, and absence-of-run workflow variants.
- Regenerated the failure catalog from the tightened playbook set and added checks so stale generated docs are caught instead of drifting after a playbook is removed.
- Promoted 4 additional real fixtures, bringing the checked-in real corpus to 215 accepted failures with 100% top-1 and top-3 baseline pass rates.
- Reduced overlap noise in the bundled catalog; `make review` now passes against 260 classified conflict patterns.
- Kept specialized, provider-specific, and maintainer-only work out of the first-run story unless it has deterministic tests, fixture evidence, and release-grade verification.

## Core Commands

```bash
faultline analyze build.log
cat build.log | faultline analyze --json
faultline fix build.log
faultline workflow build.log --json --mode agent
faultline list
faultline explain missing-executable
```

- `analyze`: classify a failing log and show evidence, diagnosis, and next action.
- `fix`: print the remediation steps for the top diagnosis.
- `workflow`: generate deterministic follow-through output for automation or handoff.
- `list`: browse known failure classes.
- `explain`: inspect one failure class before trusting or changing it.

Companion surfaces such as `inspect`, `guard`, `trace`, `replay`, `compare`,
`report`, `history`, `batch`, `coverage`, and `packs` exist, but they are not
the first-run story. `report` and `history` read only the local forensic store
created by prior local runs; cross-repo recurrence and coordination belong to
the Team layer.

## What It Catches

Faultline ships with 173 bundled playbooks for common CI failure classes.

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

Use Faultline after a job fails and you have a log file to classify.

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

## Metrics

Faultline optimizes for high precision over broad coverage. The current checked-in fixture baseline is scoped to the repository corpus, not every possible CI failure.

- Bundled playbooks: 173
- Accepted real fixtures: 215
- Fixture top-1 baseline pass rate: 100% (215/215)
- Fixture false positives: 0
- Weak matches: 1
- Published large-scale GitHub Actions evaluation: 89.4% of 30,094 failed logs matched at least one bundled playbook

The lower playbook count is intentional: this release favors a smaller, cleaner default catalog over keeping low-evidence rules in the main path. These metrics mean the known corpus is reproducible and guarded against regression. They do not mean every new log should match. Silence is intentional when the evidence is unknown, ambiguous, or below the classifier threshold.

Faultline is most useful when a team already sees recurring CI failures and wants standard classifications, repeatable fix paths, and machine-readable artifacts without adding runtime network calls to the analysis path.

Details: [docs/fixture-corpus.md](docs/fixture-corpus.md).

## Team Questions

Faultline Core answers one-log questions locally: what failed, what evidence supports that classification, and what fix path is known.

Team work should be framed around operational questions:

- What keeps failing?
- Where are we losing time?
- What should we standardize?

The release boundary is documented in [docs/release-boundary.md](docs/release-boundary.md).

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
VERSION=v0.4.7
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

## License

MIT. See [LICENSE](LICENSE).
