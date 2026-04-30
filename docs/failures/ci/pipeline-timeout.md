# CI job or pipeline exceeded time limit

**Playbook ID:** `pipeline-timeout`
**Category:** ci
**Severity:** high
**Tags:** `timeout`, `ci`, `hung`, `pipeline`

## What this failure means

The CI job ran past its configured time limit and was killed by the CI system. This is a hard stop — no further steps ran.

## Common log signals

```text
Job's timeout exceeded
execution timeout
timeout was exceeded
execution took longer than
exceeded the maximum execution time
build timed out
pipeline timed out
timed out after
```

## Diagnosis

The CI job ran past its configured time limit and was killed by the CI system. This is a hard stop — no further steps ran.

## Fix steps

1. Check the last progressing log line before the timeout killed the job —
   this tells you whether something was still working or had silently hung.

2. Wrap suspect commands with an explicit timeout so the failure is fast and
   obvious rather than silent:

   ```bash
   timeout 120 <command>          # exits non-zero if it exceeds 120 s
   ```

3. Remove interactive prompts and TTY-only behaviour from setup commands:
   `--yes`, `--non-interactive`, `CI=true`, `DEBIAN_FRONTEND=noninteractive`.

4. For Go test hangs: send SIGQUIT to get a goroutine dump showing exactly
   which goroutine is blocked:

   ```bash
   kill -QUIT $(pgrep 'test.timeout')
   # or set GOTRACEBACK=all before the test run
   ```

   For Node/Jest hangs: use `--detectOpenHandles` to find resource leaks.

5. Cache dependency downloads and split long-running stages into smaller
   jobs before increasing the global timeout limit.

6. Increase the CI timeout only after identifying a bounded slow path that
   legitimately needs more time — raising it blindly hides regressions.

## Validation

- Re-run the failing step with `timeout <N>` wrapper and confirm it
  completes within the expected window.
- Confirm no goroutine dump, `--detectOpenHandles`, or hang indicator
  appears in the output.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `.circleci/config.yml`
- `Makefile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pipeline-timeout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI job or pipeline exceeded time limit
- Ci: ci job or pipeline exceeded time limit
- ERROR: Job failed (system failure): job execution failed on runner
- GitHub Actions ci job or pipeline exceeded time limit
- faultline explain pipeline-timeout


---

*Generated from [playbooks/bundled/log/ci/pipeline-timeout.yaml](../../../playbooks/bundled/log/ci/pipeline-timeout.yaml). Do not edit directly — run `make docs-generate`.*
