# Docker BuildKit session lost

**Playbook ID:** `buildkit-session-lost`
**Category:** build
**Severity:** high
**Tags:** `docker`, `buildkit`, `build`, `session`, `timeout`

## What this failure means

Docker BuildKit lost its session while loading the build definition or context, so the image build stopped before it could execute any build steps.

## Common log signals

```text
no active session for
load build definition from Dockerfile
load .dockerignore
no active session
```

## Diagnosis

Docker BuildKit lost its session while loading the build definition or context, so the image build stopped before it could execute any build steps.

This usually means the build frontend could not keep its session open long enough to read the Dockerfile or context files.

## Fix steps

1. Re-run the build with plain progress output so the failing phase is easier to see:

   ```bash
   docker build --progress=plain .
   ```

2. Confirm the Dockerfile path and build context are correct.
3. Check whether the CI runner is terminating idle sessions or losing its Docker connection mid-build.
4. If you are using BuildKit or `buildx`, verify the builder instance is healthy:

   ```bash
   docker buildx ls
   docker buildx inspect
   ```

## Validation

- `docker build --progress=plain .` reaches the first build step without a session error.
- Re-run the failing job and confirm the build definition and context load successfully.

## Likely files to inspect

- `Dockerfile`
- `.dockerignore`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `docker-compose*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain buildkit-session-lost
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker BuildKit session lost
- Build: docker buildkit session lost
- load build definition from Dockerfile
- GitHub Actions docker buildkit session lost
- faultline explain buildkit-session-lost
- Docker docker buildkit session lost


---

*Generated from [playbooks/bundled/log/build/buildkit-session-lost.yaml](../../../playbooks/bundled/log/build/buildkit-session-lost.yaml). Do not edit directly — run `make docs-generate`.*
