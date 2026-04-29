# Dependency version drift

**Playbook ID:** `dependency-drift`
**Category:** build
**Severity:** medium
**Tags:** `dependencies`, `lockfile`, `version`

## What this failure means

Dependency constraints have drifted apart enough that the package manager
cannot compute a valid install plan.

## Common log signals

```text
conflicting requirements
dependency conflict
version conflict
incompatible versions
requires a different version
error: there is a conflict
failed to resolve dependencies
dependency resolution failed
```

## Diagnosis

Two or more packages now require incompatible versions of a shared
dependency, so resolution fails before installation can begin.

## Fix steps

1. Re-run dependency resolution from a clean cache or clean environment to
   confirm the conflict is reproducible.
2. Inspect the dependency tree with the native tool, such as `npm ls`,
   `pipdeptree`, `bundle viz`, or `go mod graph`.
3. Regenerate the lockfile with the team-standard tool and version, then
   commit the result instead of hand-editing generated files.
4. Pin, upgrade, or replace the conflicting dependency pair so top-level
   constraints agree on a compatible range.
5. If the conflict appears only in CI, align the runtime or package manager
   version with local development before changing dependency constraints.

## Validation

- Re-run dependency resolution from a clean environment.
- Confirm the lockfile and install step complete without version conflicts.

## Likely files to inspect

- `package.json`
- `package-lock.json`
- `requirements.txt`
- `go.mod`


## Run Faultline

```bash
faultline analyze build.log
faultline explain dependency-drift
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Dependency version drift
- Build: dependency version drift
- failed to resolve dependencies
- GitHub Actions dependency version drift
- faultline explain dependency-drift


---

*Generated from [playbooks/bundled/log/build/dependency-drift.yaml](../../playbooks/bundled/log/build/dependency-drift.yaml). Do not edit directly — run `make docs-generate`.*
