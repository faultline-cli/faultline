# Resource limit exceeded

**Playbook ID:** `resource-limits`
**Category:** runtime
**Severity:** high
**Tags:** `ulimit`, `cpu`, `file-descriptors`, `threads`

## What this failure means

The process exceeded an OS-enforced resource limit such as the maximum number of threads or processes.

## Common log signals

```text
ulimit
resource temporarily unavailable
cannot create thread
system limit
```

## Diagnosis

The process exceeded an OS-enforced resource limit such as the maximum number of threads or processes.

## Fix steps

1. Determine which limit was hit first (`ulimit -a`, open files, pids, threads) so you change the right control instead of raising all limits blindly.
2. Find and fix the leak or bursty workload causing the limit breach, such as file descriptors left open, unbounded goroutines, or too many worker processes.
3. Reduce test or worker parallelism to lower peak file descriptor, process, and thread usage.
4. If the workload is legitimate, raise the specific runner or container limit that failed, for example `ulimit -n 65536` for open files.
5. Capture diagnostics during the failing step (`lsof`, process counts, runtime metrics) so you can see whether usage is steadily leaking or spiking.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Resource limit exceeded is gone.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain resource-limits
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Resource limit exceeded
- Runtime: resource limit exceeded
- resource temporarily unavailable
- faultline explain resource-limits


---

*Generated from [playbooks/bundled/log/runtime/resource-limits.yaml](../../../playbooks/bundled/log/runtime/resource-limits.yaml). Do not edit directly — run `make docs-generate`.*
