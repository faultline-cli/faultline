# Terraform state lock conflict

**Playbook ID:** `terraform-state-lock`
**Category:** deploy
**Severity:** high
**Tags:** `terraform`, `state`, `lock`, `infra`

## What this failure means

Terraform could not acquire the state lock because another Terraform process is currently holding it. Only one Terraform operation can modify a given state at a time.

## Common log signals

```text
error locking state
state lock
lock id
another terraform process
terraform state is locked
error acquiring the state lock
```

## Diagnosis

Terraform could not acquire the state lock because another Terraform process is currently holding it. Only one Terraform operation can modify a given state at a time.

State locking prevents concurrent modifications to the same infrastructure. The lock is usually held by a parallel CI job, a manual run, or a crashed previous job that did not release the lock.

## Fix steps

1. Identify which job or operator currently holds the lock before taking action. The backend usually records the lock ID and owner context.
2. Wait for any active Terraform `plan` or `apply` touching the same state to complete, or cancel the duplicate run in CI.
3. Use `terraform force-unlock <LOCK_ID>` only after confirming the lock is stale and no active Terraform process is still working against that state.
4. Check your CI configuration for parallel jobs, retries, or manual approval paths that target the same Terraform backend or workspace.

## Validation

- Re-run the Terraform command after the active lock holder exits or the stale lock is cleared.
- Confirm Terraform acquires the lock cleanly and proceeds without a state-lock error.

## Likely files to inspect

- `*.tf`
- `terraform/**`
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain terraform-state-lock
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Terraform state lock conflict
- Deploy: terraform state lock conflict
- error acquiring the state lock
- faultline explain terraform-state-lock
- Terraform terraform state lock conflict


---

*Generated from [playbooks/bundled/log/deploy/terraform-state-lock.yaml](../../playbooks/bundled/log/deploy/terraform-state-lock.yaml). Do not edit directly — run `make docs-generate`.*
