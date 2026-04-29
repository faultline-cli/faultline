# File path case mismatch

**Playbook ID:** `path-case-mismatch`
**Category:** build
**Severity:** medium
**Tags:** `case-sensitive`, `filesystem`, `import`, `path`

## What this failure means

A file or module import path uses a different capitalisation from the actual filename on disk. On case-sensitive filesystems (Linux CI runners), this causes an immediate build failure.

## Common log signals

```text
cannot find module
module not found
file not found:
case-insensitive
path is not a file
capitalization
wrong case
```

## Diagnosis

A file or module import path uses a different capitalisation from the actual filename on disk. On case-sensitive filesystems (Linux CI runners), this causes an immediate build failure.

## Fix steps

1. Identify the mismatched import: compare the exact filename on disk with
   the casing used in the import statement.

2. Find all case variants of the file regardless of the current OS:

   ```bash
   find . -iname "filename.ts" -not -path "*/node_modules/*"
   git ls-files | grep -i <filename>
   ```

3. Rename the file or fix the import so the casing is consistent across the
   codebase — make the fix on a case-sensitive file system (Linux CI) where
   the mismatch is visible, not macOS where it is silently accepted.

4. Search the entire codebase for all occurrences before committing:

   ```bash
   grep -rn "from.*[Ff]ilename" src/
   ```

5. For TypeScript projects, enable `"forceConsistentCasingInFileNames": true`
   in `tsconfig.json` to catch future mismatches at compile time rather than
   only on case-sensitive file systems.

## Validation

- `find . -iname "<filename>" -not -path "*/node_modules/*"` returns exactly
  one result with the intended casing.
- Re-run the build and confirm no case-sensitivity error.

## Likely files to inspect

- `src/`
- `tsconfig.json`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain path-case-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- File path case mismatch
- Build: file path case mismatch
- cannot find module
- GitHub Actions file path case mismatch
- faultline explain path-case-mismatch


---

*Generated from [playbooks/bundled/log/build/path-case-mismatch.yaml](../../playbooks/bundled/log/build/path-case-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
