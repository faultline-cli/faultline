# Unresolved merge conflict in source code

**Playbook ID:** `merge-conflict`
**Category:** build
**Severity:** high
**Tags:** `git`, `merge`, `conflict`, `syntax`, `build`

## What this failure means

A file contains unresolved merge conflict markers (`<<<<<<<`, `=======`,
`>>>>>>>`). The build, linter, or type-checker rejects the file because the
conflict was not resolved before the commit was pushed.

## Common log signals

```text
<<<<<<< HEAD
>>>>>>> 
merge conflict
CONFLICT (content)
Automatic merge failed
<<<<<<< 
conflict marker
unresolved conflict
```

## Diagnosis

A file contains unresolved merge conflict markers left over after a merge or
rebase. The compiler or linter treats the markers as a syntax error because
they are not valid in any programming language.

Common causes:

- A merge or rebase was interrupted and finished partially.
- An editor auto-resolved a conflict incorrectly, leaving stale markers.
- The conflict was committed accidentally without checking the diff.

## Fix steps

1. Find all files with unresolved conflict markers:

   ```bash
   git grep -n "^<<<<<<< " -- ':!*.md'
   # or:
   grep -rn "^<<<<<<< " --include="*.go" --include="*.ts" --include="*.py" .
   ```

2. Open each affected file and resolve the conflict by choosing one side,
   combining both, or rewriting the section entirely. Remove all three
   marker lines (`<<<<<<<`, `=======`, `>>>>>>>`).

3. After resolving, verify no markers remain:

   ```bash
   git diff HEAD | grep "^[+-].*<<<<<<< " && echo "STILL HAS CONFLICTS" || echo "clean"
   ```

4. Stage the resolved files and amend or add a new commit:

   ```bash
   git add <file>
   git commit --amend   # if the conflict was in the last commit
   # or: git commit -m "Resolve merge conflict in <file>"
   ```

5. To prevent this in future: configure a pre-commit hook that rejects
   conflict markers:

   ```bash
   git config core.hooksPath .githooks
   # In .githooks/pre-commit:
   grep -rn "^<<<<<<< " --include="*.go" --include="*.ts" . && exit 1 || true
   ```

## Validation

- `git grep "^<<<<<<< " -- ':!*.md'` returns no results.
- The build command that previously failed now exits 0.

## Likely files to inspect

- `**/*.go`
- `**/*.ts`
- `**/*.py`
- `**/*.js`
- `**/*.java`


## Run Faultline

```bash
faultline analyze build.log
faultline explain merge-conflict
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Unresolved merge conflict in source code
- Build: unresolved merge conflict in source code
- Automatic merge failed
- GitHub Actions unresolved merge conflict in source code
- faultline explain merge-conflict


---

*Generated from [playbooks/bundled/log/build/merge-conflict.yaml](../../../playbooks/bundled/log/build/merge-conflict.yaml). Do not edit directly — run `make docs-generate`.*
