# TypeScript compile or type-check failure

**Playbook ID:** `typescript-compile`
**Category:** build
**Severity:** high
**Tags:** `typescript`, `tsc`, `javascript`, `typecheck`

## What this failure means

The TypeScript compiler rejected the code because a type, symbol, or module import does not match the project's declared types.

## Common log signals

```text
error ts
tsc:
type 'string' is not assignable to type
property does not exist on type
cannot find name
compilation complete. watching for file changes
```

## Diagnosis

The TypeScript compiler rejected the code because a type, symbol, or module import does not match the project's declared types.

## Fix steps

1. Run `tsc --noEmit` locally to reproduce the failing diagnostic with the same config.
2. Inspect the first TypeScript error in the log and update the code or type definitions to match.
3. If a generated type file changed, regenerate it before committing.
4. Confirm path aliases in `tsconfig.json` match the repository layout used in CI.

## Validation

- tsc --noEmit

## Likely files to inspect

- `tsconfig.json`
- `package.json`
- `src/**/*.ts`
- `src/**/*.tsx`


## Run Faultline

```bash
faultline analyze build.log
faultline explain typescript-compile
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- TypeScript compile or type-check failure
- Build: typescript compile or type-check failure
- compilation complete. watching for file changes
- GitHub Actions typescript compile or type-check failure
- faultline explain typescript-compile


---

*Generated from [playbooks/bundled/log/build/typescript-compile.yaml](../../../playbooks/bundled/log/build/typescript-compile.yaml). Do not edit directly — run `make docs-generate`.*
