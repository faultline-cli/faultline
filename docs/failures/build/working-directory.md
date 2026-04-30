# Wrong working directory

**Playbook ID:** `working-directory`
**Category:** build
**Severity:** medium
**Tags:** `directory`, `path`, `config`

## What this failure means

A command ran in an unexpected working directory. Relative paths expected a specific directory layout that did not exist at the time of execution.

## Common log signals

```text
no such file or directory
cannot find the path specified
chdir
not a git repository
error: no such file
working directory not found
failed to change directory
```

## Diagnosis

A command ran in an unexpected working directory. Relative paths expected a specific directory layout that did not exist at the time of execution.

## Fix steps

1. Check the `working-directory` configuration in your CI workflow file.
2. Use absolute paths or `$(pwd)` references where possible.
3. Add `ls -la` before the failing step to confirm the directory contents.
4. Verify the repository structure matches what the workflow expects.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Wrong working directory is gone.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain working-directory
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Wrong working directory
- Build: wrong working directory
- cannot find the path specified
- GitHub Actions wrong working directory
- faultline explain working-directory


---

*Generated from [playbooks/bundled/log/build/working-directory.yaml](../../../playbooks/bundled/log/build/working-directory.yaml). Do not edit directly — run `make docs-generate`.*
