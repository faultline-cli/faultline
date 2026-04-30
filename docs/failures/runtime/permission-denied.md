# Permission denied

**Playbook ID:** `permission-denied`
**Category:** runtime
**Severity:** high
**Tags:** `permissions`, `file`, `chmod`, `sudo`

## What this failure means

The process tried to read, write, execute, or connect to a resource it does not have permission to access.

## Common log signals

```text
permission denied
operation not permitted
EACCES
EPERM
access denied
insufficient permissions
cannot create directory
ERROR 1045
```

## Diagnosis

The process tried to read, write, execute, or connect to a resource it does not have permission to access.

## Fix steps

1. Run `ls -la` on the target file or directory to inspect the current permissions.
2. Add `chmod +x` for scripts that need to be executable.
3. Do not rely on `sudo` in CI; redesign the step to work without elevated privileges.
4. For file system paths, ensure the CI user owns or has the correct group access.

## Validation

- Re-run the failing workflow step.
- Confirm the original failure signature for Permission denied is gone.

## Likely files to inspect

- `Dockerfile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `scripts/*`


## Run Faultline

```bash
faultline analyze build.log
faultline explain permission-denied
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Permission denied
- Runtime: permission denied
- dial unix /var/run/docker.sock
- faultline explain permission-denied


---

*Generated from [playbooks/bundled/log/runtime/permission-denied.yaml](../../../playbooks/bundled/log/runtime/permission-denied.yaml). Do not edit directly — run `make docs-generate`.*
