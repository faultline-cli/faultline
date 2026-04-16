# Faultline Examples

These examples are small, runnable inputs derived from real CI failures.

Each `.log` file has a matching `.expected.md` file so you can compare the current output with a known-good result. The missing-executable example also includes checked-in `workflow` snapshots for local and agent handoff flows.

They are intended for first-run checks, docs validation, and quick regression sanity checks. They are deliberately small; the broader regression corpus lives under `fixtures/real/` and the bundled catalog covers 67 diagnoses.

## Included examples

| Input | What it demonstrates | Expected output |
| --- | --- | --- |
| `examples/docker-auth.log` | Registry authentication or missing login during image pull | `examples/docker-auth.expected.md` |
| `examples/missing-executable.log` | Required runtime or executable missing from the job image | `examples/missing-executable.expected.md`, `examples/missing-executable.workflow.local.txt`, `examples/missing-executable.workflow.agent.json` |
| `examples/runtime-mismatch.log` | Language runtime version mismatch between the job and the project | `examples/runtime-mismatch.expected.md` |

## Run them

```bash
make build
./bin/faultline analyze examples/docker-auth.log --format markdown
./bin/faultline analyze examples/missing-executable.log --format markdown
./bin/faultline analyze examples/runtime-mismatch.log --format markdown
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

Those commands correspond to these checked-in snapshots:

- `examples/missing-executable.workflow.local.txt`
- `examples/missing-executable.workflow.agent.json`

For the full playbook behind a diagnosis:

```bash
./bin/faultline explain docker-auth
```

## Coverage snapshot

The bundled catalog currently includes diagnoses such as:

- `docker-auth`, `git-auth`, `ssh-key-auth`, `aws-credentials`
- `missing-executable`, `runtime-mismatch`, `dependency-drift`, `npm-ci-lockfile`, `poetry-lockfile-drift`
- `dns-resolution`, `ssl-cert-error`, `connection-refused`, `network-timeout`, `rate-limited`
- `permission-denied`, `disk-full`, `oom-killed`, `port-in-use`, `docker-daemon-unavailable`
- `flaky-test`, `test-timeout`, `go-data-race`, `coverage-gate-failure`, `snapshot-mismatch`
- `image-pull-backoff`, `k8s-crashloopbackoff`, `health-check-failure`, `artifact-upload-failure`, `pipeline-timeout`

Use `./bin/faultline list` to inspect the full bundled catalog.
