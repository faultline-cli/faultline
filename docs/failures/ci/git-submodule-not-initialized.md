# Git submodule not initialized or updated

**Playbook ID:** `git-submodule-not-initialized`
**Category:** ci
**Severity:** high
**Tags:** `git`, `submodule`, `init`, `checkout`, `ci`

## What this failure means

A Git submodule referenced by the repository has not been initialized or
updated. CI clones the superproject but skips submodule initialization by
default, leaving dependent source files absent and causing build or test
failures.

## Common log signals

```text
No url found for submodule
fatal: No url found for submodule path
'git submodule' is not a git command
does not contain .git
missing or broken submodule
exit code 128
```

## Diagnosis

Git does not automatically clone or update submodules unless explicitly
instructed. A plain `git clone` or `actions/checkout` without
`submodules: recursive` leaves submodule directories empty.

Symptoms:
- Build tool reports a source file or include path does not exist
- Compiler or linker cannot find a vendored library
- `git submodule status` shows `-` prefix (not initialized) or `+` prefix
  (out of sync with expected commit)
- CMake, Meson, or Make reports missing dependency paths

Check the state of all submodules:

```bash
git submodule status --recursive
```

An uninitialized submodule shows a leading `-`:

```
-a3f7d82b4 vendor/some-lib (v1.2.3)
```

## Fix steps

1. Initialize and fetch all submodules after cloning:

   ```bash
   git submodule update --init --recursive
   ```

2. For GitHub Actions, add `submodules: recursive` to the checkout step:

   ```yaml
   - uses: actions/checkout@v4
     with:
       submodules: recursive
   ```

3. For GitLab CI, set the job-level variable:

   ```yaml
   variables:
     GIT_SUBMODULE_STRATEGY: recursive
   ```

4. For CircleCI, use a checkout step with submodule initialization:

   ```yaml
   - checkout
   - run: git submodule update --init --recursive
   ```

5. If submodules reference private repositories, ensure the CI runner's SSH
   key or deploy token has read access to each submodule remote.

6. If a submodule is no longer needed, remove it cleanly:

   ```bash
   git submodule deinit -f path/to/submodule
   git rm path/to/submodule
   rm -rf .git/modules/path/to/submodule
   ```

## Validation

- Run `git submodule status --recursive` and confirm every entry starts with
  a space (clean) rather than `-` or `+`.
- Re-run the build and confirm the missing-path errors are resolved.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain git-submodule-not-initialized
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Git submodule not initialized or updated
- Ci: git submodule not initialized or updated
- fatal: No url found for submodule path
- GitHub Actions git submodule not initialized or updated
- faultline explain git-submodule-not-initialized


---

*Generated from [playbooks/bundled/log/ci/git-submodule-not-initialized.yaml](../../playbooks/bundled/log/ci/git-submodule-not-initialized.yaml). Do not edit directly — run `make docs-generate`.*
