# Docker build context or Dockerfile path issue

**Playbook ID:** `docker-build-context`
**Category:** build
**Severity:** medium
**Tags:** `docker`, `build`, `context`, `dockerfile`

## What this failure means

Docker could not read the Dockerfile or a required file from the selected build context, so the image build stopped before executing the full build.

## Common log signals

```text
failed to solve with frontend dockerfile.v0
failed to read dockerfile
dockerfile: no such file or directory
failed to compute cache key
file not found in build context or excluded by .dockerignore
forbidden path outside the build context
unable to prepare context
not found: not found
```

## Diagnosis

Docker could not read the Dockerfile or a required file from the selected build context, so the image build stopped before executing the full build.

This usually means the command ran from the wrong directory, used the wrong `-f` path, or excluded required files via `.dockerignore`.

## Fix steps

1. Verify the exact `docker build` command and confirm the final argument points at the intended build context.
2. Check that the Dockerfile path exists relative to the working directory or the `-f` flag.
3. Inspect `.dockerignore` for patterns that exclude required files such as lockfiles or source directories.
4. Re-run the same command locally from a clean checkout to confirm the context contents.

## Validation

- Re-run the local reproduction command after the fix.
- Confirm Docker can read the Dockerfile and required build inputs from the selected context.

## Likely files to inspect

- `Dockerfile`
- `.dockerignore`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain docker-build-context
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker build context or Dockerfile path issue
- Build: docker build context or dockerfile path issue
- file not found in build context or excluded by .dockerignore
- GitHub Actions docker build context or dockerfile path issue
- faultline explain docker-build-context
- Docker docker build context or dockerfile path issue


---

*Generated from [playbooks/bundled/log/build/docker-build-context.yaml](../../playbooks/bundled/log/build/docker-build-context.yaml). Do not edit directly — run `make docs-generate`.*
