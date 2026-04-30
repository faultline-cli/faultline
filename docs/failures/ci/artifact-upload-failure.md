# CI artifact upload or download failed

**Playbook ID:** `artifact-upload-failure`
**Category:** ci
**Severity:** medium
**Tags:** `artifact`, `upload`, `download`, `ci`, `pipeline`, `storage`

## What this failure means

A CI job failed to upload a build artifact, or a downstream job failed to download an artifact produced by an earlier job. Pipeline stages that depend on the artifact cannot proceed.

## Common log signals

```text
artifact upload failed
failed to upload artifact
Error uploading
artifact not found
No artifact found
Unable to find artifact
No files were found with the provided path
No artifacts will be uploaded
```

## Diagnosis

A CI job failed to upload a build artifact, or a downstream job failed to download an artifact produced by an earlier job. Pipeline stages that depend on the artifact cannot proceed.

## Fix steps

1. Check whether the producing job actually ran and exited zero before the upload step.
2. Verify the artifact path glob matches the actual output location (e.g., `dist/**` vs `build/dist/**`).
3. For GitHub Actions: if the log says `No files were found with the provided path`, resolve the path from the repository root or upload from an absolute path that exists in the runner's filesystem.
4. For GitHub Actions container jobs: confirm the upload step can see the build output path from the host runner context, not just from inside the container.
5. For GitLab CI: verify `artifacts.paths` points at files produced by the job and check whether the runner or coordinator is rejecting the upload.
6. For CircleCI: confirm `store_artifacts` path matches the actual build output directory.
7. For Jenkins: check Archivable artifacts path in the post section of the Jenkinsfile.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for CI artifact upload or download failed is gone.

## Likely files to inspect

- `.github/workflows/`
- `.gitlab-ci.yml`
- `.circleci/config.yml`
- `Jenkinsfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain artifact-upload-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CI artifact upload or download failed
- Ci: ci artifact upload or download failed
- Uploading artifacts to coordinator... forbidden
- GitHub Actions ci artifact upload or download failed
- faultline explain artifact-upload-failure


---

*Generated from [playbooks/bundled/log/ci/artifact-upload-failure.yaml](../../playbooks/bundled/log/ci/artifact-upload-failure.yaml). Do not edit directly — run `make docs-generate`.*
