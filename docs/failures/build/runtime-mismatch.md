# Python, Ruby, or Go runtime version mismatch

**Playbook ID:** `runtime-mismatch`
**Category:** build
**Severity:** high
**Tags:** `node`, `python`, `ruby`, `go`, `java`, `version`

## What this failure means

The installed Python, Ruby, or Go runtime does not satisfy the version
constraint declared by the project. The build is using a language runtime
outside the version range expected by the repository or dependency set.

## Common log signals

```text
requires python
requires ruby version
unsupported python
go.mod requires go
go version .* required
python.*requires.*version
incompatible python version
requires Python >=
```

## Diagnosis

The installed Python, Ruby, or Go runtime does not satisfy the version
constraint declared by the project. This playbook is the generic fallback
for those runtimes after more specific Node.js and Java rules have been
considered.

## Fix steps

1. Identify the required version from the repository source of truth:
   `.python-version`, `.ruby-version`, `.tool-versions`, `go.mod`, or the
   CI image tag.
2. Pin the same version in the workflow or container image that runs the
   failing job.
3. Re-run the build locally with the pinned runtime so you can confirm the
   version mismatch is resolved before re-running CI.
4. If the failure came from a dependency upgrade, confirm the new minimum
   runtime requirement before relaxing project constraints or changing the
   CI image.

## Validation

- Print the runtime version in CI and confirm it matches the version file or
  toolchain declaration.
- Re-run the failing workflow step and confirm the original version error is
  gone.

## Likely files to inspect

- `.python-version`
- `.ruby-version`
- `.tool-versions`
- `go.mod`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain runtime-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Python, Ruby, or Go runtime version mismatch
- Build: python, ruby, or go runtime version mismatch
- incompatible python version
- GitHub Actions python, ruby, or go runtime version mismatch
- faultline explain runtime-mismatch
- Node.js python, ruby, or go runtime version mismatch


---

*Generated from [playbooks/bundled/log/build/runtime-mismatch.yaml](../../playbooks/bundled/log/build/runtime-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
