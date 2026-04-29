# Lint, format, or static-analysis gate failure

**Playbook ID:** `quality-gate-failure`
**Category:** build
**Severity:** medium
**Tags:** `quality`, `lint`, `format`, `static-analysis`, `checks`

## What this failure means

A non-compiler quality gate failed during CI. The job stopped because lint,
formatting, or static-analysis checks reported violations even though the
code may still compile.

## Common log signals

```text
lint failed
format check failed
static analysis failed
quality gate failed
style violations found
code quality checks failed
formatting violations
checks did not pass
```

## Diagnosis

A non-compiler quality gate failed during CI. This playbook is the generic
fallback for quality tooling when Faultline cannot match a stronger
tool-specific rule.

## Fix steps

1. Re-run the exact quality-check command locally to see which files and rules
   are failing:

   ```bash
   make lint        # or: eslint ., golangci-lint run, ruff check .
   make check       # or a direct invocation of the tool
   ```

2. Fix the reported violations before considering any relaxation of the gate.
   Suppression comments (`// nolint`, `# noqa`) should be used sparingly and
   must include a justification.

3. If the failure followed a shared config update (`.eslintrc`, `.golangci.yml`,
   `pyproject.toml`), review that change and confirm the stricter rule was
   intentional before fixing callers.

4. If multiple tools run under one quality step, isolate each tool locally to
   identify whether the failure is lint, format, type-check, or another
   static-analysis check:

   ```bash
   eslint .                  # JavaScript linting only
   prettier --check .        # formatting only
   tsc --noEmit              # type-check only
   ```

## Validation

- `make lint` and `make check` both exit 0.
- Re-run CI and confirm the quality gate passes.

## Likely files to inspect

- `package.json`
- `pyproject.toml`
- `Makefile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain quality-gate-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Lint, format, or static-analysis gate failure
- Build: lint, format, or static-analysis gate failure
- code quality checks failed
- GitHub Actions lint, format, or static-analysis gate failure
- faultline explain quality-gate-failure


---

*Generated from [playbooks/bundled/log/build/quality-gate-failure.yaml](../../playbooks/bundled/log/build/quality-gate-failure.yaml). Do not edit directly — run `make docs-generate`.*
