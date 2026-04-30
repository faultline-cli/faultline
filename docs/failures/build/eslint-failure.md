# ESLint or linter check failure

**Playbook ID:** `eslint-failure`
**Category:** build
**Severity:** medium
**Tags:** `eslint`, `lint`, `javascript`, `typescript`, `code-quality`

## What this failure means

A linting or formatting check step (`eslint`, `tslint`, `prettier`) found violations in the source code and exited with a non-zero code, failing the CI step.

## Common log signals

```text
ESLint
eslint
Lint errors found
lint failed
error  Parsing error
problems (0 errors
tslint
prettier --check
```

## Diagnosis

A linting or formatting check step (`eslint`, `tslint`, `prettier`) found violations in the source code and exited with a non-zero code, failing the CI step.

## Fix steps

1. Run `npx eslint . --fix` locally to auto-fix fixable violations (formatting, import ordering, trailing commas).
2. Run `npx prettier --write .` for formatting violations that ESLint delegates to Prettier.
3. For rule violations that cannot be auto-fixed, address each `error` line manually — the output includes the rule name and a URL to the rule documentation.
4. To see the resolved configuration applied to a specific file: `npx eslint --print-config path/to/file.ts` — useful when a shared config override is unexpected.
5. If a new rule was enabled by a shared config update (e.g., `eslint-config-company` bumped): check the config package's changelog and decide whether to fix violations or temporarily raise the rule to `warn`.
6. For `@typescript-eslint` errors: confirm `parserOptions.project` in `.eslintrc` points to the correct `tsconfig.json` — a missing project reference causes false positives.
7. For `no-unused-vars` on type imports: use `import type` syntax or add `@typescript-eslint/no-unused-vars` with `varsIgnorePattern`.

## Validation

- npx eslint .
- npx eslint . --fix --dry-run

## Likely files to inspect

- `.eslintrc.js`
- `.eslintrc.json`
- `.eslintrc.yml`
- `.prettierrc`
- `package.json`


## Run Faultline

```bash
faultline analyze build.log
faultline explain eslint-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- ESLint or linter check failure
- Build: eslint or linter check failure
- Formatting check failed
- GitHub Actions eslint or linter check failure
- faultline explain eslint-failure


---

*Generated from [playbooks/bundled/log/build/eslint-failure.yaml](../../../playbooks/bundled/log/build/eslint-failure.yaml). Do not edit directly — run `make docs-generate`.*
