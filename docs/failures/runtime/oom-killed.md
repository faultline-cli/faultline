# Process killed by OOM killer

**Playbook ID:** `oom-killed`
**Category:** runtime
**Severity:** critical
**Tags:** `oom`, `memory`, `kill`, `exit-137`

## What this failure means

The process was terminated by the operating system's out-of-memory (OOM) killer. Exit code 137 (128 + SIGKILL) is the canonical indicator.

## Common log signals

```text
exit code 137
OOMKilled
out of memory
oom killer
signal: killed
container killed
memory limit exceeded
```

## Diagnosis

The process was terminated by the operating system's out-of-memory (OOM) killer. Exit code 137 (128 + SIGKILL) is the canonical indicator.

## Fix steps

1. Confirm the OOM kill. Exit code 137 means the process received SIGKILL:

   ```bash
   echo "Exit code: $?"            # run immediately after the failing step
   dmesg | grep -i 'oom\|kill' | tail -20   # on the host
   ```

   For Kubernetes: `kubectl describe pod <pod>` and look for `OOMKilled` in
   the Last State block.

2. Measure peak memory before increasing limits:

   ```bash
   docker stats --no-stream          # for running containers
   kubectl top pod                   # for Kubernetes pods
   /usr/bin/time -v <command>        # prints peak "Maximum resident set size"
   ```

3. Profile the process to find the largest allocators: `go tool pprof`,
   `node --inspect` + Chrome DevTools heap snapshot, `valgrind --tool=massif`,
   or `py-spy dump --memory`.
4. Reduce parallelism to lower peak memory use: fewer parallel test workers or
   build threads.
5. Check for memory leaks: loading large fixtures into memory, accumulating
   response bodies, or unbounded in-memory caches that grow over the test run.
6. Set explicit memory limits and requests in the deployment spec or CI runner
   tier so future OOM events fail fast with a clear reason.

## Validation

- Monitor memory during the re-run: `docker stats` or `kubectl top pod`.
- Confirm the job completes with exit code 0, not 137.

## Likely files to inspect

- `Dockerfile`
- `k8s/**/*.yaml`
- `deploy/**/*.yaml`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain oom-killed
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Process killed by OOM killer
- Runtime: process killed by oom killer
- memory limit exceeded
- faultline explain oom-killed


---

*Generated from [playbooks/bundled/log/runtime/oom-killed.yaml](../../../playbooks/bundled/log/runtime/oom-killed.yaml). Do not edit directly — run `make docs-generate`.*
