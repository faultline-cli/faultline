# Go compilation error

**Playbook ID:** `go-compile-error`
**Category:** build
**Severity:** high
**Tags:** `go`, `golang`, `compile`, `build`, `undefined`

## What this failure means

The Go compiler rejected the package because of a semantic error — undefined
identifier, type mismatch, unused import, or interface non-compliance. The
binary cannot be produced until the errors are resolved.

## Common log signals

```text
imported and not used
declared and not used
does not implement
undefined: 
cannot use 
too many arguments in call to
not enough arguments in call to
cannot assign to
```

## Diagnosis

The Go compiler emitted one or more hard errors and aborted the build. These
are always deterministic: the source must be corrected before `go build ./...`
can succeed.

Common error classes:
- **undefined:** — a symbol, function, or type is referenced but not declared
  in the current package or any imported package.
- **cannot use X as Y** — a value of one type is passed where a different type
  is required, and no implicit conversion exists.
- **imported and not used** — a package was imported but none of its exported
  names appear in the file.
- **declared and not used** — a local variable was declared but never read.
- **does not implement** — a concrete type is used where an interface is
  expected but is missing one or more methods.

## Fix steps

1. Run `go build ./...` locally to see the full diagnostic output:

   ```bash
   go build ./... 2>&1 | head -30
   ```

2. For `undefined: Foo`: import the package that defines `Foo` or move the
   declaration into scope.
3. For `cannot use x (variable of type A) as type B`: correct the type or add
   an explicit conversion where the semantics are safe.
4. For `imported and not used`: remove the unused import, or add a blank
   import `_ "pkg"` if the side-effect is intentional.
5. For `declared and not used`: use the variable or replace it with `_`.
6. For `does not implement`: add the missing method to the concrete type, or
   use the correct type that already satisfies the interface.

## Validation

- Run `go build ./...` locally and confirm zero errors.
- Run `go vet ./...` to catch additional semantic issues before CI.
- Confirm the CI build step exits zero.

## Likely files to inspect

- `**/*.go`
- `go.mod`


## Run Faultline

```bash
faultline analyze build.log
faultline explain go-compile-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Go compilation error
- Build: go compilation error
- multiple-value expression in single-value context
- GitHub Actions go compilation error
- faultline explain go-compile-error
- Go go compilation error


---

*Generated from [playbooks/bundled/log/build/go-compile-error.yaml](../../../playbooks/bundled/log/build/go-compile-error.yaml). Do not edit directly — run `make docs-generate`.*
