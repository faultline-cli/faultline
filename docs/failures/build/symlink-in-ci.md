# Symlink resolution failure in CI

**Playbook ID:** `symlink-in-ci`
**Category:** build
**Severity:** medium
**Tags:** `symlink`, `symlinks`, `filesystem`, `ci`, `linux`, `dangling`

## What this failure means

A symlink in the repository or the build workspace is broken or cannot be
resolved in the CI environment. Common causes include symlinks pointing to
absolute paths that only exist on a developer's machine, or symlinks that
survive a `git clone` but point to targets that are not checked out.

## Common log signals

```text
Too many levels of symbolic links
dangling symlink
broken symlink
symlink.*not found
readlink: Invalid argument
ln: failed to create symbolic link
symbolic link loop
ELOOP
```

## Diagnosis

Symlinks behave differently across operating systems and CI environments:
- Absolute symlinks (e.g., `/Users/dev/project/lib`) break on any other machine
- Symlinks inside a shallow clone may point to commits or paths not present
- Windows CI runners may not support symlinks without Developer Mode enabled
- Docker `COPY` with some configurations does not preserve symlinks

Find broken symlinks:

```bash
# List all symlinks
find . -type l | head -20

# Find broken symlinks (dangling)
find . -xtype l 2>/dev/null

# Show what each symlink points to
find . -type l -exec ls -la {} \;
```

## Fix steps

1. Replace absolute symlinks with relative ones:

   ```bash
   # Remove the absolute symlink
   rm path/to/broken-link

   # Re-create as relative
   ln -sr actual/target path/to/broken-link
   # or manually:
   ln -s ../actual/target path/to/broken-link
   ```

2. If the symlink target is in `.gitignore`, either:
   - Move the target into a tracked location, or
   - Replace the symlink with a real file or a build step

3. Verify symlinks are committed correctly in git:

   ```bash
   git ls-files -s | grep "^120000"   # shows symlinks in index
   ```

4. For Windows CI runners that do not support symlinks, replace them with
   file copies or use `git config core.symlinks true` at job start:

   ```yaml
   - name: Enable symlinks
     run: git config --global core.symlinks true
     if: runner.os == 'Windows'
   ```

5. In Docker, use `COPY --follow-link` (BuildKit) to dereference symlinks
   during the build context transfer, or copy the target directly.

## Validation

- Run `find . -xtype l` and confirm it produces no output.
- Re-run the failing CI step.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain symlink-in-ci
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Symlink resolution failure in CI
- Build: symlink resolution failure in ci
- ln: failed to create symbolic link
- GitHub Actions symlink resolution failure in ci
- faultline explain symlink-in-ci


---

*Generated from [playbooks/bundled/log/build/symlink-in-ci.yaml](../../../playbooks/bundled/log/build/symlink-in-ci.yaml). Do not edit directly — run `make docs-generate`.*
