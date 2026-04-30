# Docker base image update introduced a breaking change

**Playbook ID:** `base-image-breaking-change`
**Category:** build
**Severity:** high
**Tags:** `docker`, `base-image`, `breaking-change`, `tag`, `update`, `floating-tag`

## What this failure means

A Docker build that previously succeeded now fails because the base image
tag (e.g., `ubuntu:latest`, `node:20`) was updated and the new version
introduces a breaking change. The build or runtime behavior changed without
any modification to the project's Dockerfile or application code.

## Common log signals

```text
package.*not found
executable.*not found
command not found
no such file or directory
Unable to locate package
Package.*has no installation candidate
symbol lookup error
version.*GLIBC
```

## Diagnosis

Floating tags (`latest`, `20`, `3.12`) do not guarantee a stable image.
The upstream maintainer may:
- Upgrade the OS distribution (Ubuntu 22.04 → 24.04)
- Remove a tool or library that was previously included
- Change the default shell, user, or working directory
- Update a linked library that changes the ABI

Confirm the image changed:

```bash
# Compare the image digest
docker pull node:20
docker inspect node:20 --format '{{.Id}}'

# Or check when the image was pushed
docker inspect node:20 --format '{{.Created}}'
```

Check the upstream image changelog:
- Docker Hub: image > Tags tab for recent pushes
- GitHub Container Registry: packages page

## Fix steps

1. Pin the base image to an exact digest or patch-level tag:

   ```dockerfile
   # Instead of:
   FROM node:20

   # Use:
   FROM node:20.11.0-alpine3.19
   # Or with digest (strongest pin):
   FROM node:20.11.0-alpine3.19@sha256:abc123...
   ```

2. Now that the breakage is understood, fix the underlying issue:
   - If a package was removed from the base image, install it explicitly in
     the Dockerfile
   - If a binary path changed, update the `CMD` or `ENTRYPOINT`
   - If a glibc version changed, either use the old base image or recompile
     binaries

3. Update dependencies to be compatible with the new base image:

   ```dockerfile
   # Install required system packages explicitly, don't rely on base image
   RUN apt-get update && apt-get install -y --no-install-recommends \
       libssl3 \
       ca-certificates \
   && rm -rf /var/lib/apt/lists/*
   ```

4. Use Dependabot or Renovate to get automated, tested PRs when the base
   image updates:

   ```yaml
   # .github/dependabot.yml
   version: 2
   updates:
     - package-ecosystem: docker
       directory: /
       schedule:
         interval: weekly
   ```

## Validation

- Build the Docker image locally with the pinned base image tag.
- Run the container and confirm the application starts correctly.
- Re-run the CI pipeline.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain base-image-breaking-change
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker base image update introduced a breaking change
- Build: docker base image update introduced a breaking change
- Package.*has no installation candidate
- GitHub Actions docker base image update introduced a breaking change
- faultline explain base-image-breaking-change
- Docker docker base image update introduced a breaking change


---

*Generated from [playbooks/bundled/log/build/base-image-breaking-change.yaml](../../../playbooks/bundled/log/build/base-image-breaking-change.yaml). Do not edit directly — run `make docs-generate`.*
