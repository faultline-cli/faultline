# Faultline Examples

These examples are small, runnable inputs derived from real fixture patterns.

Run them after `make build`:

```bash
./bin/faultline analyze examples/docker-auth.log --format markdown
./bin/faultline analyze examples/missing-executable.log --format markdown
./bin/faultline analyze examples/runtime-mismatch.log --format markdown
```

Expected outputs are checked in beside each log as `*.expected.md` for quick comparison.
