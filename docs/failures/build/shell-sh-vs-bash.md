# Shell compatibility issue (sh vs bash)

**Playbook ID:** `shell-sh-vs-bash`
**Category:** build
**Severity:** medium
**Tags:** `shell`, `bash`, `sh`, `posix`, `dash`, `container`, `script`

## What this failure means

A shell script uses bash-specific syntax but is executed by `/bin/sh`, which
is commonly `dash` on Ubuntu/Debian or `ash` on Alpine. Constructs like
`[[ ]]`, `source`, `local` outside functions, arrays, and process
substitution are bash extensions and fail silently or loudly under POSIX sh.

## Common log signals

```text
syntax error: unexpected
source: not found
local: not in a function
[[: not found
declare: not found
function: not found
Illegal option
no such file or directory.*bash
```

## Diagnosis

On many Linux distributions, `/bin/sh` is not bash:
- Ubuntu/Debian: `/bin/sh` → `dash` (faster POSIX sh, no bash extensions)
- Alpine: `/bin/sh` → `busybox ash` (minimal POSIX sh)
- macOS: `/bin/sh` → `bash 3.2` (old bash, missing newer features)

A script with `#!/bin/bash` shebang runs correctly. The same script run by a
CI step without a shebang, or run with `sh script.sh`, may fail.

Check the actual shell:

```bash
ls -la /bin/sh   # reveals symlink target
sh --version 2>/dev/null || sh -c "echo $0"
```

Find bash-specific syntax in scripts:

```bash
# Common bash-only constructs
grep -rn '\[\[' --include="*.sh" .
grep -rn 'source ' --include="*.sh" .
grep -rn 'declare -' --include="*.sh" .
grep -rn '<<(' --include="*.sh" .   # process substitution
```

## Fix steps

1. Add a `#!/usr/bin/env bash` shebang to scripts that use bash extensions:

   ```bash
   #!/usr/bin/env bash
   set -euo pipefail

   # Now bash-specific syntax is safe
   [[ -n "${VAR}" ]] && echo "set"
   declare -a arr=(a b c)
   source helpers.sh
   ```

2. Install bash in Alpine containers that use bash scripts:

   ```dockerfile
   # Alpine does not include bash by default
   RUN apk add --no-cache bash
   ```

3. Alternatively, rewrite bash-specific constructs using POSIX sh equivalents:

   ```bash
   # bash: [[ ]]  →  POSIX: [ ]
   if [ -n "$VAR" ]; then ...

   # bash: source  →  POSIX: . (dot)
   . helpers.sh

   # bash: local   →  POSIX: only valid inside functions (both support it)
   # bash arrays   →  No POSIX equivalent; restructure or use bash

   # bash: <<()    →  POSIX alternative uses temporary files
   ```

4. For GitHub Actions, specify the shell explicitly:

   ```yaml
   - name: Run script
     shell: bash
     run: |
       [[ -n "$VAR" ]] && echo "set"
   ```

5. Use ShellCheck to detect bash-isms in scripts intended to be portable:

   ```bash
   shellcheck --shell=sh script.sh   # strict POSIX mode
   shellcheck script.sh              # uses shebang to determine shell
   ```

## Validation

- Run `shellcheck script.sh` and resolve any issues it reports.
- Test with the target shell: `sh script.sh` and `bash script.sh`.
- Re-run the CI step and confirm the syntax error is resolved.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain shell-sh-vs-bash
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Shell compatibility issue (sh vs bash)
- Build: shell compatibility issue (sh vs bash)
- no such file or directory.*bash
- GitHub Actions shell compatibility issue (sh vs bash)
- faultline explain shell-sh-vs-bash


---

*Generated from [playbooks/bundled/log/build/shell-sh-vs-bash.yaml](../../../playbooks/bundled/log/build/shell-sh-vs-bash.yaml). Do not edit directly — run `make docs-generate`.*
