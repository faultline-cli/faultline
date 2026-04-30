# GitHub Actions GITHUB_TOKEN permission denied

**Playbook ID:** `github-actions-permission`
**Category:** ci
**Severity:** high
**Tags:** `github-actions`, `token`, `permissions`, `ci`

## What this failure means

The `GITHUB_TOKEN` used by this workflow does not have the permissions required for the requested GitHub operation.

## Common log signals

```text
HttpError: Resource not accessible by integration
Error: HttpError: Resource not accessible by integration
GraphQL: Resource not accessible by integration
Permission denied to github-actions\[bot\]
requires the 'contents: write' permission
Missing permissions
insufficient permission
refusing to allow a GitHub App to create or update file
```

## Diagnosis

The `GITHUB_TOKEN` used by this workflow does not have the permissions required for the requested GitHub operation.

## Fix steps

1. Add an explicit `permissions` block to the workflow or failing job and grant only the scopes that step actually needs.
2. Check the exact action or API call that failed and map it to the required permission such as `contents: write`, `packages: write`, `pull-requests: write`, or `id-token: write`.
3. For organization-owned repositories, confirm the repository or org-level Actions settings do not force a more restrictive default token policy.
4. If a third-party action is making the request, review its README for the minimum required permission scopes.

## Validation

- Re-run the job and confirm the `Resource not accessible by integration` or permission error is gone.
- Verify the workflow's `permissions:` block is present in the committed YAML.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-permission
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions GITHUB_TOKEN permission denied
- Ci: github actions github_token permission denied
- Error: HttpError: Resource not accessible by integration
- GitHub Actions github actions github_token permission denied
- faultline explain github-actions-permission


---

*Generated from [playbooks/bundled/log/ci/github-actions-permission.yaml](../../../playbooks/bundled/log/ci/github-actions-permission.yaml). Do not edit directly — run `make docs-generate`.*
