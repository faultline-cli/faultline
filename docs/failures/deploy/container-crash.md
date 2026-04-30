# Container exited unexpectedly

**Playbook ID:** `container-crash`
**Category:** deploy
**Severity:** high
**Tags:** `container`, `crash`, `oomkilled`, `exit-code`, `kubernetes`, `docker`

## What this failure means

The container started but then exited with a non-zero status code. If it keeps restarting, Kubernetes enters CrashLoopBackOff and the pod never becomes ready.

## Common log signals

```text
container exited with
exit code 1
exit code 139
container is restarting
```

## Diagnosis

The container started but then exited with a non-zero status code. If it
keeps restarting, Kubernetes enters CrashLoopBackOff and the pod never
becomes ready.

CrashLoopBackOff is the platform symptom. The root cause is usually in the
last container logs, the exit code, or the startup configuration passed to
the image.

## Fix steps

1. Retrieve the logs from the previous crashed container: `kubectl logs <pod> --previous`.
2. Check the exit code with `kubectl describe pod <pod>` — look for the Last State block.
3. If exit 137, increase the memory limit in the container spec or profile memory usage.
4. If exit 139, inspect the logs for segfaults or crashes in native extensions.
5. If exit 1, search the log output for `FATAL`, `panic`, or unhandled exception traces.
6. Reproduce the failure locally with `docker run` using the same image and environment variables.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Container exited unexpectedly is gone.

## Likely files to inspect

- `Dockerfile`
- `k8s/**/*.yaml`
- `deploy/**/*.yaml`
- `docker-compose*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain container-crash
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Container exited unexpectedly
- Deploy: container exited unexpectedly
- container is restarting
- faultline explain container-crash
- Kubernetes container exited unexpectedly


---

*Generated from [playbooks/bundled/log/deploy/container-crash.yaml](../../../playbooks/bundled/log/deploy/container-crash.yaml). Do not edit directly — run `make docs-generate`.*
