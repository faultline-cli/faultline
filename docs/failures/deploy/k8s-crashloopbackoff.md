# Kubernetes pod in CrashLoopBackOff

**Playbook ID:** `k8s-crashloopbackoff`
**Category:** deploy
**Severity:** high
**Tags:** `kubernetes`, `k8s`, `crashloopbackoff`, `container`, `restart`, `pod`

## What this failure means

A Kubernetes pod is stuck in CrashLoopBackOff: the container starts, exits with a non-zero code, and Kubernetes restarts it with exponential back-off. The root cause is almost always inside the container itself.

## Common log signals

```text
CrashLoopBackOff
Back-off restarting failed container
Error: CrashLoopBackOff
Restarting failed container
Reason: CrashLoopBackOff
```

## Diagnosis

A Kubernetes pod is stuck in CrashLoopBackOff: the container starts, exits with a non-zero code, and Kubernetes restarts it with exponential back-off. The root cause is almost always inside the container itself.

## Fix steps

1. Read the container's last log output before it exited: `kubectl logs <pod> --previous`.
2. Describe the pod for exit code and reason: `kubectl describe pod <pod> -n <namespace>` — look at `Last State` under the container section.
3. Exit code 1 = the application crashed with an error; exit code 137 = OOM killed; exit code 126/127 = entrypoint not found or not executable.
4. For missing configuration: check that all required `ConfigMaps`, `Secrets`, and volume mounts exist and have the expected keys.
5. For application startup errors: reproduce with `docker run --env-file .env <image>` to isolate the failure from Kubernetes.
6. For OOM: increase the container's `resources.limits.memory` or reduce the application's memory footprint.
7. Force-trigger a log dump by temporarily patching the command to sleep: `kubectl patch deployment <name> --patch '{"spec":{"template":{"spec":{"containers":[{"name":"<n>","command":["sleep","3600"]}]}}}}'` then exec in to inspect.

## Validation

- kubectl get pod <pod> -n <namespace> -w
- kubectl logs <pod> -n <namespace>

## Likely files to inspect

- `k8s/**`
- `kubernetes/**`
- `helm/**`
- `Dockerfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain k8s-crashloopbackoff
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Kubernetes pod in CrashLoopBackOff
- Deploy: kubernetes pod in crashloopbackoff
- Back-off restarting failed container
- faultline explain k8s-crashloopbackoff
- Kubernetes kubernetes pod in crashloopbackoff


---

*Generated from [playbooks/bundled/log/deploy/k8s-crashloopbackoff.yaml](../../playbooks/bundled/log/deploy/k8s-crashloopbackoff.yaml). Do not edit directly — run `make docs-generate`.*
