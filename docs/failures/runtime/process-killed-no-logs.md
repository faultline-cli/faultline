# Process killed without diagnostic output

**Playbook ID:** `process-killed-no-logs`
**Category:** runtime
**Severity:** high
**Tags:** `sigkill`, `killed`, `exit-137`, `exit-143`, `signal`, `oom`, `timeout`

## What this failure means

A CI process was killed externally, producing little or no diagnostic output.
Exit code 137 (SIGKILL) is the most common indicator. The kill is usually
triggered by the OOM killer, a CI job timeout, or a watchdog process
terminating a runaway job.

## Common log signals

```text
Killed
signal: killed
exit status 137
exit code: 137
Exit Code: 137
exit code 143
signal 9
SIGKILL
```

## Diagnosis

A bare `Killed` with exit 137 means the process was sent SIGKILL from
outside, not from application code. The absence of a stack trace or panic is
a key diagnostic clue.

Exit code reference:
- **137** = 128 + 9 (SIGKILL): OOM killer, container memory limit, or
  explicit `kill -9`
- **143** = 128 + 15 (SIGTERM): CI job timeout, graceful shutdown request
- **130** = 128 + 2 (SIGINT): Manual cancellation or Ctrl-C in script

Determine the cause:

1. Check runner memory usage at the time of failure in CI platform dashboards.
2. Check whether the job time limit was reached:

   ```bash
   # GitHub Actions: look for "The operation was canceled" in the job log
   # GitLab CI: look for "Job ... timed out" in the runner log
   ```

3. Check for OOM events in the runner or container:

   ```bash
   # Linux kernel OOM log
   dmesg | grep -i "oom\|killed process"
   journalctl -k | grep -i "oom\|Out of memory"
   ```

## Fix steps

**If OOM-killed:**
1. Profile memory usage and reduce peak consumption (see `oom-killed` playbook
   for detailed steps).
2. Increase the runner or container memory limit.
3. Reduce parallelism: use `--workers=1` or equivalent to lower peak memory.

**If killed by job timeout:**
1. Identify and fix slow tests or steps that consistently approach the limit.
2. Split the job into smaller parallel jobs that each complete within budget.
3. Increase the timeout as a temporary measure while fixing the root cause:

   ```yaml
   # GitHub Actions
   jobs:
     build:
       timeout-minutes: 60   # increase from default 6-hour or explicit limit

   # GitLab CI
   job_name:
     timeout: 1h
   ```

**If killed by watchdog or concurrency limit:**
1. Ensure the process respects SIGTERM by installing a signal handler and
   flushing output before exiting.
2. Check whether CI concurrency limits cancel in-progress jobs when a newer
   run starts (GitHub Actions `concurrency` with `cancel-in-progress: true`).

**General:**
- Add resource monitoring to the CI job so memory and CPU trends are visible
  before the kill occurs.
- Ensure test output is flushed before the process exits so partial results
  are not lost.

## Validation

- Re-run the failing job and confirm it completes without being killed.
- Check that the exit code is 0 rather than 137 or 143.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain process-killed-no-logs
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Process killed without diagnostic output
- Runtime: process killed without diagnostic output
- process exited with signal
- faultline explain process-killed-no-logs


---

*Generated from [playbooks/bundled/log/runtime/process-killed-no-logs.yaml](../../../playbooks/bundled/log/runtime/process-killed-no-logs.yaml). Do not edit directly — run `make docs-generate`.*
