# Faultline Examples

These examples are small, runnable inputs derived from real CI failures.

Each `.log` file has a matching `.expected.md` file so you can compare the current output with a known-good result. The missing-executable example also includes checked-in snapshots for `workflow`, `replay`, and `trace`, plus a deterministic `compare` snapshot against the runtime-mismatch sample.

They are intended for first-run checks, docs validation, and quick regression sanity checks. They are deliberately small; the broader regression corpus lives under `fixtures/real/` and the bundled catalog currently ships 100 playbooks.

## Included examples

| Input | What it demonstrates | Expected output |
| --- | --- | --- |
| `examples/docker-auth.log` | Registry authentication or missing login during image pull | `examples/docker-auth.expected.md` |
| `examples/missing-executable.log` | Required runtime or executable missing from the job image | `examples/missing-executable.expected.md`, `examples/missing-executable.workflow.local.txt`, `examples/missing-executable.workflow.agent.json`, `examples/missing-executable.replay.expected.md`, `examples/missing-executable.trace.expected.md` |
| `examples/runtime-mismatch.log` | Language runtime version mismatch between the job and the project | `examples/runtime-mismatch.expected.md` |
| `examples/missing-vs-runtime.compare.expected.md` | Deterministic compare report between missing-executable and runtime-mismatch analyses | `examples/missing-vs-runtime.compare.expected.md` |

## High-impact demo path

Run the core product flow first: diagnose, then generate the deterministic workflow handoff.

```bash
make build
./bin/faultline analyze examples/missing-executable.log --format markdown --mode quick
cat examples/missing-executable.log | ./bin/faultline workflow --no-history
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --no-history
```

## Run all examples

```bash
make build
./bin/faultline analyze examples/docker-auth.log --format markdown
./bin/faultline analyze examples/missing-executable.log --format markdown
./bin/faultline analyze examples/runtime-mismatch.log --format markdown
make cli-smoke
```

For a tighter remediation view:

```bash
./bin/faultline fix examples/docker-auth.log --format markdown
```

For the deterministic follow-up workflow:

```bash
cat examples/missing-executable.log | ./bin/faultline workflow --no-history
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --no-history
```

For the optional deterministic evidence-fusion layer:

```bash
./bin/faultline analyze examples/missing-executable.log --json --bayes
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --bayes --no-history
```

Those commands correspond to these checked-in snapshots:

- `examples/missing-executable.workflow.local.txt`
- `examples/missing-executable.workflow.agent.json`
- `examples/missing-executable.replay.expected.md`
- `examples/missing-executable.trace.expected.md`
- `examples/missing-vs-runtime.compare.expected.md`

To refresh the checked-in example outputs after a renderer, workflow, or playbook change:

```bash
./bin/faultline analyze examples/docker-auth.log --format markdown --no-history > examples/docker-auth.expected.md
./bin/faultline analyze examples/missing-executable.log --format markdown --no-history > examples/missing-executable.expected.md
./bin/faultline analyze examples/runtime-mismatch.log --format markdown --no-history > examples/runtime-mismatch.expected.md
cat examples/missing-executable.log | ./bin/faultline workflow --no-history > examples/missing-executable.workflow.local.txt
cat examples/missing-executable.log | ./bin/faultline workflow --json --mode agent --no-history > examples/missing-executable.workflow.agent.json
cat examples/missing-executable.log | ./bin/faultline analyze --json --no-history > /tmp/missing.analysis.json
cat examples/runtime-mismatch.log | ./bin/faultline analyze --json --no-history > /tmp/runtime.analysis.json
./bin/faultline replay --format markdown --mode detailed /tmp/missing.analysis.json > examples/missing-executable.replay.expected.md
cat examples/missing-executable.log | ./bin/faultline trace --format markdown --playbook missing-executable --no-history > examples/missing-executable.trace.expected.md
./bin/faultline compare --format markdown /tmp/missing.analysis.json /tmp/runtime.analysis.json > examples/missing-vs-runtime.compare.expected.md
```

For the full playbook behind a diagnosis:

```bash
./bin/faultline explain docker-auth
```

For a quiet high-confidence local prevention check in a repository:

```bash
./bin/faultline guard .
```

## Coverage snapshot

The bundled catalog currently includes 100 playbooks across 98 log diagnoses and 2 source-detector rules. Representative diagnoses include:

- `docker-auth`, `git-auth`, `ssh-key-auth`, `aws-credentials`
- `missing-executable`, `runtime-mismatch`, `dependency-drift`, `npm-ci-lockfile`, `poetry-lockfile-drift`
- `dns-resolution`, `ssl-cert-error`, `connection-refused`, `network-timeout`, `rate-limited`
- `permission-denied`, `disk-full`, `oom-killed`, `port-in-use`, `docker-daemon-unavailable`
- `flaky-test`, `test-timeout`, `go-data-race`, `coverage-gate-failure`, `snapshot-mismatch`
- `image-pull-backoff`, `k8s-crashloopbackoff`, `health-check-failure`, `artifact-upload-failure`, `pipeline-timeout`

Use `./bin/faultline list` to inspect the full bundled catalog.
