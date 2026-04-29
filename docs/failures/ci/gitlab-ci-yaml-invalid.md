# GitLab CI pipeline configuration invalid

**Playbook ID:** `gitlab-ci-yaml-invalid`
**Category:** ci
**Severity:** high
**Tags:** `gitlab`, `gitlab-ci`, `yaml`, `lint`, `pipeline`, `ci`

## What this failure means

The `.gitlab-ci.yml` configuration failed GitLab's built-in YAML and schema validation. No pipeline was started.

## Common log signals

```text
ERROR: Config file is not valid
ERROR: Job is invalid
config validation error
jobs:.*config contains unknown keys
Included file
error validating
did you mean
needs.*not defined
```

## Diagnosis

The `.gitlab-ci.yml` configuration failed GitLab's built-in YAML and schema validation. No pipeline was started.

## Fix steps

1. Run the GitLab CI Lint API to get a detailed error: visit `<gitlab-url>/-/ci/lint` and paste your config.
2. Use the GitLab CLI: `glab ci lint .gitlab-ci.yml` for local validation.
3. Validate the YAML structure first: `yamllint .gitlab-ci.yml`.
4. Check the exact line in the error — `needs:` jobs must exist in the same or a prior stage.
5. Ensure every stage referenced by a job is declared in the top-level `stages:` list.
6. For `include:` directives, verify the referenced file path, project, or URL is accessible to the runner.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for GitLab CI pipeline configuration invalid is gone.

## Likely files to inspect

- `.gitlab-ci.yml`
- `.gitlab/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain gitlab-ci-yaml-invalid
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitLab CI pipeline configuration invalid
- Ci: gitlab ci pipeline configuration invalid
- jobs:.*config contains unknown keys
- GitHub Actions gitlab ci pipeline configuration invalid
- faultline explain gitlab-ci-yaml-invalid
- GitLab CI gitlab ci pipeline configuration invalid


---

*Generated from [playbooks/bundled/log/ci/gitlab-ci-yaml-invalid.yaml](../../playbooks/bundled/log/ci/gitlab-ci-yaml-invalid.yaml). Do not edit directly — run `make docs-generate`.*
