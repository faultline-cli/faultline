# Container image pull failure

**Playbook ID:** `image-pull-backoff`
**Category:** deploy
**Severity:** high
**Tags:** `kubernetes`, `docker`, `image`, `registry`, `deploy`

## What this failure means

The deployment failed because the runtime could not pull the requested container image from the registry.

## Common log signals

```text
imagepullbackoff
errimagepull
failed to pull image
pull access denied for
back-off pulling image
rpc error: code = unknown desc = error response from daemon
```

## Diagnosis

The deployment failed because the runtime could not pull the requested container image from the registry.

## Fix steps

1. Verify the image name and tag in the deployment manifest exactly match a published image.
2. Confirm the cluster or runner has registry credentials for the target repository.
3. Inspect the imagePullSecrets or runtime credential configuration used by the deployment.
4. Check the latest successful build or release job to confirm the image was pushed before deploy started.
5. Inspect the pod events and confirm the failure is an image reference or credential issue rather than a broader scheduling problem.

## Validation

- kubectl describe pod <pod>
- kubectl get events --sort-by=.lastTimestamp

## Likely files to inspect

- `Dockerfile`
- `docker-compose*.yml`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `k8s/**/*.yaml`
- `deploy/**/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain image-pull-backoff
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Container image pull failure
- Deploy: container image pull failure
- rpc error: code = unknown desc = error response from daemon
- faultline explain image-pull-backoff
- Kubernetes container image pull failure


---

*Generated from [playbooks/bundled/log/deploy/image-pull-backoff.yaml](../../playbooks/bundled/log/deploy/image-pull-backoff.yaml). Do not edit directly — run `make docs-generate`.*
