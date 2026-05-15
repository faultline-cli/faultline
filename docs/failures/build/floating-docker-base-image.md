# Docker base image uses a floating latest tag

**Playbook ID:** `floating-docker-base-image`
**Category:** build
**Severity:** medium
**Tags:** `source`, `docker`, `build`, `reproducibility`, `image`

## What this failure means

A Dockerfile uses a `:latest` base image tag, so rebuilds can pull different bytes over time without a source change.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

The Docker base image is mutable. CI builds that rely on `:latest` can become non-reproducible when the upstream image moves.

## Fix steps

1. Replace `:latest` with a specific version tag that matches the supported runtime.
2. For release images, pin the base image by digest or use an automated dependency update workflow.
3. Rebuild the image and commit the intended base version in the Dockerfile.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Build the Docker image from a clean cache and confirm it uses the pinned base image.

## Likely files to inspect

- `Dockerfile`
- `docker/Dockerfile`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain floating-docker-base-image
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker base image uses a floating latest tag
- Build: docker base image uses a floating latest tag
- GitHub Actions docker base image uses a floating latest tag
- faultline explain floating-docker-base-image
- Docker docker base image uses a floating latest tag


---

*Generated from [playbooks/bundled/source/floating-docker-base-image.yaml](../../../playbooks/bundled/source/floating-docker-base-image.yaml). Do not edit directly — run `make docs-generate`.*
