# Dependency removed or yanked from upstream registry

**Playbook ID:** `dependency-removed-upstream`
**Category:** build
**Severity:** high
**Tags:** `dependency`, `removed`, `yanked`, `deleted`, `npm`, `pypi`, `registry`

## What this failure means

A required dependency has been removed, yanked, or unpublished from the
upstream registry. Package managers cannot resolve the dependency and the
build fails. This is different from a version mismatch or lockfile error:
the exact version or the entire package no longer exists in the registry.

## Common log signals

```text
Package.*not found
no matching version found for
package has been deprecated
package has been deleted
package.*is no longer
unable to find package
package.*does not exist
error getting.*package
```

## Diagnosis

Common causes:
1. An author unpublished the package (npm allows this within 72 hours)
2. A package was yanked for a security vulnerability (PyPI, RubyGems)
3. The package was renamed or moved to a new scope (e.g., `pkg` → `@org/pkg`)
4. A registry purged spam or abandoned packages
5. A private package was deleted from an internal registry

Distinguish from a network issue:

```bash
# Test registry directly
# npm
npm view <package>@<version>

# pip
pip index versions <package>

# If the command returns "not found", the package is gone.
# If it times out, it's a network issue.
```

## Fix steps

1. Verify the package is truly gone (not a transient network failure):

   ```bash
   curl https://registry.npmjs.org/<package>/<version>   # should 404
   curl https://pypi.org/pypi/<package>/<version>/json   # should 404
   ```

2. Find a replacement or successor:
   - Search the registry for the package name + readme for migration notes
   - Check the package's GitHub repo for archival or redirect notices
   - Look for a fork maintained by a different author

3. Update the dependency manifest to use the replacement:

   ```bash
   npm uninstall old-package && npm install new-package@latest
   pip uninstall old-package && pip install new-package
   ```

4. If the package was only yanked (not deleted), and you still need the
   specific version, vendor the dependency:

   ```bash
   # npm workspaces can reference local packages
   npm pack old-package-1.2.3.tgz
   npm install ./old-package-1.2.3.tgz
   ```

5. If the package was an internal/private package, restore it to the
   internal registry or update the package source.

6. Update the lockfile after changing the dependency:

   ```bash
   npm ci          # regenerates node_modules from updated lockfile
   pip-compile     # regenerates requirements.txt from pyproject.toml
   ```

## Validation

- Run `npm install` or `pip install -r requirements.txt` and confirm it
  succeeds without 404 errors.
- Re-run the full CI pipeline.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain dependency-removed-upstream
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Dependency removed or yanked from upstream registry
- Build: dependency removed or yanked from upstream registry
- no matching version found for
- GitHub Actions dependency removed or yanked from upstream registry
- faultline explain dependency-removed-upstream
- npm dependency removed or yanked from upstream registry


---

*Generated from [playbooks/bundled/log/build/dependency-removed-upstream.yaml](../../../playbooks/bundled/log/build/dependency-removed-upstream.yaml). Do not edit directly — run `make docs-generate`.*
