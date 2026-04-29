# Docker manifest not found or bad image tag

**Playbook ID:** `docker-manifest-not-found`
**Category:** build
**Severity:** medium
**Tags:** `docker`, `registry`, `image`, `tag`, `manifest`

## What this failure means

Docker could not pull an image because the tag does not exist in the registry, the manifest is missing, or the image reference is malformed.

## Common log signals

```text
manifest not found
tag does not exist
repository does not exist
image not found
no such image
invalid reference format
manifest unknown
manifest for
```

## Diagnosis

Docker tried to pull an image from a registry (Docker Hub, ECR, GCR, etc.) but the tag or digest does not exist, the manifest is unavailable, or the image reference contains a typo.

Common causes:
- The image tag was deleted or never pushed.
- The image name or tag contains a typo.
- The registry requires authentication that is not configured.
- The image is private and the CI job lacks credentials.
- The image uses a multi‑arch manifest that is incomplete for the current platform.

## Fix steps

1. **Verify the image reference**:
   ```bash
   echo "Image: $IMAGE"
   docker manifest inspect $IMAGE 2>&1 | head -20
   ```

2. **Check if the tag exists** in the registry:
   - For Docker Hub: visit `https://hub.docker.com/r/<org>/<image>/tags`
   - For ECR: use `aws ecr describe-images --repository-name <repo>`
   - For GCR: use `gcloud container images list-tags <image>`

3. **If the image is private**, ensure credentials are available:
   ```bash
   echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
   ```

4. **If using a multi‑arch image**, verify it supports your platform:
   ```bash
   docker buildx imagetools inspect $IMAGE
   ```

5. **Retry with an explicit tag** instead of `latest`:
   ```bash
   docker pull myimage:1.2.3
   ```

6. **If the image is built in a previous CI step**, ensure it was pushed successfully before the pull step.

7. **Check network connectivity** to the registry (firewall, proxy, DNS).

## Validation

- `docker pull <image>` succeeds without "manifest not found" or "tag does not exist".
- `docker run --rm <image> echo ok` runs without image pull errors.

## Likely files to inspect

- `Dockerfile`
- `docker-compose.yml`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `kubernetes/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain docker-manifest-not-found
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker manifest not found or bad image tag
- Build: docker manifest not found or bad image tag
- repository does not exist
- GitHub Actions docker manifest not found or bad image tag
- faultline explain docker-manifest-not-found
- Docker docker manifest not found or bad image tag


---

*Generated from [playbooks/bundled/log/build/docker-manifest-not-found.yaml](../../playbooks/bundled/log/build/docker-manifest-not-found.yaml). Do not edit directly — run `make docs-generate`.*
