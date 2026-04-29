# Change outside component ownership caused failure

**Playbook ID:** `topology-ownership-mismatch`
**Category:** build
**Severity:** medium
**Tags:** `topology`, `codeowners`, `ownership`, `monorepo`, `review-gate`

## What this failure means

The files that were recently changed are owned by a different team than the
files where the failure originated, indicating that an out-of-team edit
introduced a regression in code the editor does not own.

## Common log signals

```text
CODEOWNERS
code owners
required review
requires review from
not authorized to merge
protection rules
branch protection
```

## Diagnosis

CODEOWNERS analysis shows that the recently changed files are in a different
ownership zone than the failing component. This typically means one of:

- A developer edited code outside their area of ownership, introducing a
  subtle incompatibility.
- A tooling or infra change (e.g. shared build scripts, Docker base images,
  Go workspace, monorepo configuration) had an unintended side effect on a
  downstream component.
- A code-generation step or auto-formatter was run without checking the
  downstream impact.

In all cases, the owning team was not the author of the change, so the normal
review signal for that component was absent when the change landed.

## Fix steps

1. Identify the ownership mismatch:
   ```
   git log --name-only -5 | head -30
   cat CODEOWNERS
   ```
2. Determine whether the intent was to change the failing component's
   behaviour or whether it was an unintended side effect.
3. Notify the owning team and open a joint review or revert the change.
4. If the change is intentional, add the downstream owner as a required
   reviewer in CODEOWNERS and ensure they approve before merge.
5. Add a regression test in the failing component to pin the relevant
   behaviour.

## Validation

- The failing component's owner has reviewed and approved the change.
- CI passes for the full test suite of both zones.
- CODEOWNERS coverage is reviewed for any newly shared paths.

## Likely files to inspect

- `CODEOWNERS`


## Run Faultline

```bash
faultline analyze build.log
faultline explain topology-ownership-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Change outside component ownership caused failure
- Build: change outside component ownership caused failure
- not authorized to merge
- GitHub Actions change outside component ownership caused failure
- faultline explain topology-ownership-mismatch


---

*Generated from [playbooks/bundled/log/build/topology-ownership-mismatch.yaml](../../playbooks/bundled/log/build/topology-ownership-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
