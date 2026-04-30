# Syntax error in source code

**Playbook ID:** `syntax-error`
**Category:** build
**Severity:** critical
**Tags:** `syntax`, `parse`, `compile`

## What this failure means

The compiler or interpreter hit invalid source syntax and stopped before the
build could continue.

## Common log signals

```text
SyntaxError:
unexpected token
parse error
syntax error near
unterminated string literal
expected ';' before
error: expected expression
invalid syntax
```

## Diagnosis

The failing source file contains a parse error such as an unclosed delimiter,
invalid token, missing separator, or unresolved merge marker.

This is a hard failure: the code must be corrected before the build can pass.

## Fix steps

1. Look at the file and line number from the error message.
2. Run the compiler or linter locally to reproduce the exact parse failure.
3. Check for unclosed brackets, missing separators, or merge conflict
   markers such as `<<<<<<<`, `=======`, and `>>>>>>>`.
4. If the file was generated or transformed, rebuild the generated output
   from source rather than patching the artifact by hand.

## Validation

- Re-run the parser, compiler, or linter locally.
- Confirm the build step completes without the syntax error.

## Likely files to inspect

- `src/`
- `internal/`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain syntax-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Syntax error in source code
- Build: syntax error in source code
- Failed to load interface for
- GitHub Actions syntax error in source code
- faultline explain syntax-error


---

*Generated from [playbooks/bundled/log/build/syntax-error.yaml](../../../playbooks/bundled/log/build/syntax-error.yaml). Do not edit directly — run `make docs-generate`.*
