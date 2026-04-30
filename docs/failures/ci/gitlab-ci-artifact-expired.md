# GitLab CI artifact missing or expired

**Playbook ID:** `gitlab-ci-artifact-expired`
**Category:** ci
**Severity:** medium
**Tags:** `gitlab`, `artifact`, `expired`, `dependency`, `ci`

## What this failure means

A downstream GitLab CI job could not retrieve an artifact produced by an earlier job because it expired, was never produced, or was referenced incorrectly.

## Common log signals

```text
artifacts have expired
expired artifact
Could not retrieve artifact
needs: artifact
no artifacts to download
artifact download failed
```

## Diagnosis

A downstream GitLab CI job could not retrieve an artifact produced by an earlier job because it expired, was never produced, or was referenced incorrectly.

## Fix steps

1. Re-run the full pipeline from the beginning instead of retrying only the downstream job.
2. Increase the artifact lifetime with `artifacts: expire_in:` on the producing job.
3. Check the failing job's `needs:` or `dependencies:` references and confirm they point to the correct producer job.
4. Verify the upstream job actually declares `artifacts.paths` for the files the downstream job expects.

## Validation

- Re-run the pipeline from the producing job onward.
- Confirm the downstream job can download the artifact and no expiry message remains in the log.

## Likely files to inspect

- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain gitlab-ci-artifact-expired
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitLab CI artifact missing or expired
- Ci: gitlab ci artifact missing or expired
- Could not retrieve artifact
- GitHub Actions gitlab ci artifact missing or expired
- faultline explain gitlab-ci-artifact-expired


---

*Generated from [playbooks/bundled/log/ci/gitlab-ci-artifact-expired.yaml](../../../playbooks/bundled/log/ci/gitlab-ci-artifact-expired.yaml). Do not edit directly — run `make docs-generate`.*
