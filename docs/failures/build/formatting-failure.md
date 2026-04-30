# Code formatting check failure

**Playbook ID:** `formatting-failure`
**Category:** build
**Severity:** low
**Tags:** `formatting`, `prettier`, `black`, `gofmt`, `rustfmt`, `lint`, `style`

## What this failure means

A CI formatting check failed because one or more source files do not match
the project's required code style as enforced by a formatter (Prettier,
Black, gofmt, rustfmt, etc.). The check runs in verification mode and
exits non-zero without modifying files.

## Common log signals

```text
Reformatting...
would reformat
files would be reformatted
not formatted
format check failed
run gofmt
run goimports
prettier.*check
```

## Diagnosis

Most projects run a formatting check in `--check` mode in CI, which exits
non-zero if any file would be changed. The underlying code is syntactically
valid but does not conform to the configured style.

This commonly happens when:
- A developer's editor does not have format-on-save configured
- The developer used a different formatter version than CI
- Auto-generated code was committed without running the formatter
- A merge or rebase introduced formatting inconsistencies

Identify which files are affected:

```bash
# Prettier
npx prettier --check .

# Black (Python)
black --check .

# Go
gofmt -l .     # lists files that differ

# Rust
cargo fmt -- --check
```

## Fix steps

1. Run the formatter in write mode locally to auto-fix all issues:

   ```bash
   # Prettier
   npx prettier --write .

   # Black
   black .

   # Go
   gofmt -w .
   goimports -w .

   # Rust
   cargo fmt

   # C/C++
   clang-format -i path/to/file.cpp
   ```

2. Stage and commit the formatting changes:

   ```bash
   git add -u
   git commit -m "style: apply formatter"
   ```

3. Ensure all team members have format-on-save configured in their editors,
   or install a pre-commit hook to format automatically:

   ```bash
   # Using pre-commit framework
   # .pre-commit-config.yaml
   repos:
     - repo: https://github.com/psf/black
       rev: 24.3.0
       hooks:
         - id: black
   ```

4. If the formatter version in CI differs from local, pin or align them:

   ```json
   // package.json
   "prettier": "3.2.5"
   ```

   ```toml
   # pyproject.toml
   [tool.black]
   target-version = ['py311']
   ```

## Validation

- Re-run the formatter in check mode and confirm it exits 0.
- Push the formatted commit and verify the CI formatting step passes.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain formatting-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Code formatting check failure
- Build: code formatting check failure
- declaration-block-no-duplicate-properties
- GitHub Actions code formatting check failure
- faultline explain formatting-failure


---

*Generated from [playbooks/bundled/log/build/formatting-failure.yaml](../../../playbooks/bundled/log/build/formatting-failure.yaml). Do not edit directly — run `make docs-generate`.*
