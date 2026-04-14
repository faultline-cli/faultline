# Faultline Examples

These examples are small, runnable inputs derived from real CI failures.

Each `.log` file has a matching `.expected.md` file so you can compare the current output with a known-good result.

They are intended for first-run checks, docs validation, and quick regression sanity checks.

## Included examples

| Input | What it demonstrates | Expected output |
| --- | --- | --- |
| `examples/docker-auth.log` | Registry authentication or missing login during image pull | `examples/docker-auth.expected.md` |
| `examples/missing-executable.log` | Required runtime or executable missing from the job image | `examples/missing-executable.expected.md` |
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

For the full playbook behind a diagnosis:

```bash
./bin/faultline explain docker-auth
```
