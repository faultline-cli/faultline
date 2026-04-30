# CLI tool flag or command changed between versions

**Playbook ID:** `cli-flag-changed`
**Category:** build
**Severity:** medium
**Tags:** `cli`, `flags`, `breaking-change`, `tool-version`, `argument`

## What this failure means

A CI script or Makefile invokes a CLI tool with a flag or subcommand that no
longer exists in the installed version. A tool upgrade introduced a
breaking change to its command-line interface, or the local and CI
environments have different tool versions.

## Common log signals

```text
unknown flag
unknown option
unrecognized option
invalid flag
flag provided but not defined
no such option
option.*not recognized
unexpected argument
```

## Diagnosis

CLI interface changes commonly occur when:
1. A tool is upgraded in CI (via `latest` tag or unpinned install) and the
   new version deprecates or removes a flag
2. A developer upgrades locally and updates scripts without noticing CI uses
   an older version
3. A flag is renamed (e.g., `--no-color` → `--color=never`)

Identify the version mismatch:

```bash
# Check what version CI is using
tool --version

# Compare against what the script expects
grep -r "tool --flag" . --include="*.sh" --include="Makefile" -l
```

Look for recent changelog entries describing the change:

```bash
# For npm packages
npx tool --help 2>&1 | grep -A2 "removed\|deprecated"
```

## Fix steps

1. Update the script to use the new flag syntax:

   ```bash
   # Before (deprecated flag)
   tool --old-flag value

   # After
   tool --new-flag value
   ```

2. If the change is accidental (a CI upgrade broke things), pin the tool
   version:

   ```yaml
   # GitHub Actions
   - uses: actions/setup-node@v4
     with:
       node-version: '20.11.0'   # pin, not '20' or 'lts/*'

   # Generic
   - run: npm install -g tool@2.1.0  # pin major+minor+patch
   ```

3. For tools installed via `apt`, pin the version:

   ```bash
   apt-get install -y tool=2.1.0-1ubuntu1
   ```

4. Align local and CI tool versions using a `.tool-versions` file (asdf) or
   `mise.toml`:

   ```toml
   [tools]
   mytool = "2.1.0"
   ```

5. Add a `--version` check at the start of CI to surface version mismatches
   early.

## Validation

- Run the updated command locally with the CI tool version.
- Push and confirm the CI step passes.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain cli-flag-changed
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CLI tool flag or command changed between versions
- Build: cli tool flag or command changed between versions
- flag provided but not defined
- GitHub Actions cli tool flag or command changed between versions
- faultline explain cli-flag-changed


---

*Generated from [playbooks/bundled/log/build/cli-flag-changed.yaml](../../../playbooks/bundled/log/build/cli-flag-changed.yaml). Do not edit directly — run `make docs-generate`.*
