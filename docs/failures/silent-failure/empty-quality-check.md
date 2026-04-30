# Lint, scan, or coverage step ran but checked nothing

**Playbook ID:** `empty-quality-check`
**Category:** silent_failure
**Severity:** medium
**Tags:** `silent-failure`, `lint`, `scan`, `security`, `coverage`, `empty-check`

## What this failure means

A lint, security-scan, or coverage step ran, but the log shows it processed
zero files or produced no meaningful report. The step may have exited zero,
but the intended quality check never really happened.

## Common log signals

```text
No files to lint
0 files checked
0 files scanned
0 packages scanned
Nothing to scan
No source files found
0 files analyzed
Coverage report not generated
```

## Diagnosis

Quality gates often succeed when they have no input. A linter, scanner, or
coverage tool can report zero files checked if the target path is wrong, the
working directory is unexpected, or the repository checkout is incomplete.

Common causes:

- The step runs in the wrong directory and sees no source files.
- The file glob passed to the tool matches nothing.
- A previous checkout or artifact-download step did not materialize the code.
- Coverage generation was skipped, so the reporting step has no data.

The result is a false green signal: the quality check step ran, but it did no
actual work.

## Fix steps

1. Verify the tool's working directory and file globs.
2. Confirm the repository or artifact checkout completed before the quality step.
3. Print the source tree in CI to ensure the expected files are present.
4. Add a guard that fails when zero files are analyzed unexpectedly.

## Validation

Re-run the job and confirm the step reports a positive number of files,
packages, or coverage records analyzed.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `.golangci.yml`
- `.semgrep.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain empty-quality-check
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Lint, scan, or coverage step ran but checked nothing
- Silent Failure: lint, scan, or coverage step ran but checked nothing
- Coverage report not generated
- faultline explain empty-quality-check


---

*Generated from [playbooks/bundled/log/silent/empty-quality-check.yaml](../../../playbooks/bundled/log/silent/empty-quality-check.yaml). Do not edit directly — run `make docs-generate`.*
