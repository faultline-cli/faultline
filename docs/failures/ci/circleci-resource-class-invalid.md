# CircleCI machine image or resource class not valid

**Playbook ID:** `circleci-resource-class-invalid`
**Category:** ci
**Severity:** high
**Tags:** `circleci`, `resource-class`, `machine-image`, `executor`, `ci`

## What this failure means

A CircleCI job failed to start because the specified machine image or resource class is no longer valid, was removed, or was never available for the organization.

## Common log signals

```text
not a valid resource class
resource class.*is not a valid
machine image.*is not a valid
image.*not a valid resource class
Job was rejected
job failed to schedule
unsupported machine image
image was renamed
```

## Diagnosis

CircleCI rejected the job at scheduling time because:

- The machine image (e.g., `ubuntu-2404:2026.02.2`) was an internal or preview image that was removed from the index.
- The resource class is not available for the specified executor or org plan.
- The image name was renamed or deprecated between releases.

## Fix steps

1. Check the current list of available machine images on CircleCI Developer - Machine Images.

2. Update the machine image to a stable, released version:
   ```yaml
   machine:
     image: ubuntu-2404:2026.02.20
   ```

3. For resource class errors, verify the class is available for your plan:
   ```yaml
   resource_class: medium
   ```

4. If using a preview or internal image, migrate to the stable alias:
   - `ubuntu-2404:2026.02.2` → `ubuntu-2404:2026.02.20`
   - Or use `docker-docker29` alias instead of direct image name.

5. Contact CircleCI support if the error persists for stable images.

## Validation

- The pipeline triggers and the job is scheduled on the expected executor.
- The job reaches the first step successfully.
- No "not a valid resource class" or "image is not valid" error in the job setup.

## Likely files to inspect

- `.circleci/config.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain circleci-resource-class-invalid
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CircleCI machine image or resource class not valid
- Ci: circleci machine image or resource class not valid
- image.*not a valid resource class
- GitHub Actions circleci machine image or resource class not valid
- faultline explain circleci-resource-class-invalid


---

*Generated from [playbooks/bundled/log/ci/circleci-resource-class-invalid.yaml](../../playbooks/bundled/log/ci/circleci-resource-class-invalid.yaml). Do not edit directly — run `make docs-generate`.*
