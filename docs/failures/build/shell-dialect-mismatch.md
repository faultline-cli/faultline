# Bash syntax used under a POSIX sh shell

**Playbook ID:** `shell-dialect-mismatch`
**Category:** build
**Severity:** medium
**Tags:** `source`, `shell`, `ci`, `bash`, `sh`

## What this failure means

A CI workflow or script selects `sh` but uses Bash-only syntax such as `[[ ... ]]`, `source`, heredoc strings, functions, or `set -o pipefail`.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

The command is likely to run under POSIX `sh` while using syntax that requires Bash. This causes CI-only failures on runners where `/bin/sh` is not Bash.

## Fix steps

1. Change the workflow step to `shell: bash` or update the script shebang to Bash.
2. If Bash is not required, rewrite the command using POSIX-compatible syntax.
3. Keep the selected shell and script syntax aligned in CI and local runs.

## Validation

- Run `faultline inspect .` from the repository root and confirm this source finding is absent or intentionally mitigated.
- Run the script with the configured shell and confirm it exits successfully.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `scripts/*.sh`


## Run Faultline

```bash
faultline analyze build.log
faultline explain shell-dialect-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Bash syntax used under a POSIX sh shell
- Build: bash syntax used under a posix sh shell
- GitHub Actions bash syntax used under a posix sh shell
- faultline explain shell-dialect-mismatch


---

*Generated from [playbooks/bundled/source/shell-dialect-mismatch.yaml](../../../playbooks/bundled/source/shell-dialect-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
