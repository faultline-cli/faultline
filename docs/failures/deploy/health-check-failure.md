# Service health check failure

**Playbook ID:** `health-check-failure`
**Category:** deploy
**Severity:** high
**Tags:** `health-check`, `kubernetes`, `docker`, `deployment`

## What this failure means

The new service instance never became healthy enough for traffic.

## Common log signals

```text
health check failed
liveness probe failed
readiness probe failed
readiness probe always fails
unhealthy
service is not ready
waiting for deployment to be ready
/health/ready
```

## Diagnosis

The deploy reached the runtime environment, but the process did not satisfy the configured readiness or liveness checks before the timeout window closed.

If liveness stays green while readiness returns `503`, the process usually started but is still blocked on startup work such as migrations, downstream API checks, or dependency initialization.

Common causes:

- missing configuration or secrets
- startup migrations failing
- the app listening on the wrong port
- a crash loop during bootstrap

## Fix steps

1. Inspect startup logs for the failing container:

   ```bash
   kubectl logs <pod> --previous
   ```

2. Inspect the pod description for the probe configuration and recent events:

   ```bash
   kubectl describe pod <pod>
   ```

3. Verify the health check path and port match what the application actually serves.
4. Confirm all required environment variables and secrets are present in the deployment.
5. Check that dependent services such as databases and caches are reachable from the new pod.
6. If liveness is `200` while readiness is `503`, inspect startup tasks and dependency checks that gate the ready endpoint.

## Validation

- Hit the readiness endpoint directly from a pod or staging environment.
- Confirm the deployment becomes ready without restarts.
- Re-run the rollout and watch the probe status until it stabilizes.

## Likely files to inspect

- `k8s/**/*.yaml`
- `deploy/**/*.yaml`
- `docker-compose*.yml`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain health-check-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Service health check failure
- Deploy: service health check failure
- waiting for deployment to be ready
- faultline explain health-check-failure
- Kubernetes service health check failure


---

*Generated from [playbooks/bundled/log/deploy/health-check-failure.yaml](../../../playbooks/bundled/log/deploy/health-check-failure.yaml). Do not edit directly — run `make docs-generate`.*
