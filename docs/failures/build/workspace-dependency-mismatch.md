# Monorepo workspace dependency version mismatch

**Playbook ID:** `workspace-dependency-mismatch`
**Category:** build
**Severity:** high
**Tags:** `monorepo`, `lerna`, `nx`, `yarn`, `pnpm`, `workspace`, `dependency`, `version-mismatch`

## What this failure means

A package in the monorepo workspace depends on a version of another workspace package that doesn't match the version published in the workspace registry.

## Common log signals

```text
workspace.*version mismatch
unmet workspace dependencies
cannot find module.*matching ^
Unmet workspace dependencies detected
no packages matched
cannot find project
```

## Diagnosis

Modern monorepos use tools like Lerna, Nx, Yarn Workspaces, or pnpm Workspaces to manage multiple interdependent packages. Each workspace package is versioned independently. When a package specifies a dependency on another workspace package with a version constraint that doesn't match any published version, the workspace resolution fails.

This is distinct from a missing package — the package exists in the workspace, but the version constraint can't be satisfied.

Common causes:
- A package was upgraded to v2, but dependent packages still require v1
- A package's `package.json` specifies `@org/shared@^2.0.0` but the workspace contains `@org/shared@1.2.0`
- Cross-package version constraints are inconsistent after a partial update
- Lerna/Nx bootstrap tries to resolve transitive dependencies and fails
- Version specifiers are too restrictive (e.g., `@org/core@1.5.0` instead of `^1.5.0`)

## Fix steps

1. Identify the conflicting dependency from the error message. It will say something like:
   ```
   @myorg/frontend requires @myorg/shared@^2.0.0
   Available: @myorg/shared@1.2.0 (in workspace registry)
   ```

2. Check the current versions of all workspace packages:

   ```bash
   lerna list --all
   # or
   npm ls -a --depth=0 --workspace
   ```

3. Choose a resolution strategy:

   **Option A: Upgrade the dependency to match the workspace version:**
   ```bash
   cd packages/frontend
   npm install --save @myorg/shared@^1.2.0
   ```

   **Option B: Upgrade the workspace package to match the constraint:**
   ```bash
   cd packages/shared
   npm version minor  # Bump 1.2.0 -> 1.3.0 or similar
   ```

   **Option C: Use a parallel version if multiple major versions are needed:**
   ```bash
   # Rename shared to shared-v1 and create shared-v2
   mkdir -p packages/shared-v2
   # Move new code as appropriate
   ```

4. Verify all package.json files in the workspace reference compatible versions:

   ```bash
   # List all interdependencies
   for dir in packages/*/; do echo "=== $dir ==="; grep -A5 '"dependencies"' "$dir/package.json"; done
   ```

5. Re-run the workspace bootstrap:

   ```bash
   lerna bootstrap
   # or
   yarn install
   # or 
   pnpm install
   ```

## Validation

- Run `lerna list` and confirm no version conflicts.
- Run `lerna bootstrap` or `npm install` without errors.
- Verify the monorepo installs successfully and all packages are symlinked correctly.

## Likely files to inspect

- `packages/*/package.json`
- `lerna.json`
- `workspace.json`
- `pnpm-workspace.yaml`
- `yarn.lock`


## Run Faultline

```bash
faultline analyze build.log
faultline explain workspace-dependency-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Monorepo workspace dependency version mismatch
- Build: monorepo workspace dependency version mismatch
- Unmet workspace dependencies detected
- GitHub Actions monorepo workspace dependency version mismatch
- faultline explain workspace-dependency-mismatch
- Yarn monorepo workspace dependency version mismatch


---

*Generated from [playbooks/bundled/log/build/workspace-dependency-mismatch.yaml](../../../playbooks/bundled/log/build/workspace-dependency-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
