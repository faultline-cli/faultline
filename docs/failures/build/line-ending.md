# Line ending mismatch (CRLF vs LF)

**Playbook ID:** `line-ending`
**Category:** build
**Severity:** medium
**Tags:** `line-ending`, `crlf`, `lf`, `git`, `windows`, `encoding`

## What this failure means

A script or source file contains Windows-style CRLF line endings where
Unix LF endings are expected. On Linux CI runners this causes shell scripts
to fail with cryptic "command not found" or "bad interpreter" errors because
the carriage return character becomes part of the command or path.

## Common log signals

```text
CRLF will be replaced by LF
LF will be replaced by CRLF
\r.*no such file or directory
bad line endings
unexpected carriage return
': command not found
\r': No such file
syntax error near unexpected token
```

## Diagnosis

Git can silently convert line endings on checkout. When `core.autocrlf` is
`true` on a Windows development machine, files are committed with CRLF and
converted back on checkout locally but left as CRLF in the committed object if
the remote does not enforce normalization.

On Linux CI runners the raw bytes are used, turning `#!/bin/bash\r` into a
missing interpreter path.

Common symptoms:
- Shell scripts exit with `'\r': command not found` or `bad interpreter`
- Python scripts fail with `SyntaxError: invalid syntax` at otherwise valid lines
- `git diff` shows no changes but files appear modified to the OS
- Linters report unexpected character errors on every line

## Fix steps

1. Identify the affected file:

   ```bash
   file path/to/script.sh          # reports "CRLF line terminators"
   cat -A path/to/script.sh | head # shows ^M at line ends
   ```

2. Convert the file to LF endings:

   ```bash
   # Using sed (in-place):
   sed -i 's/\r//' path/to/script.sh

   # Using dos2unix:
   dos2unix path/to/script.sh

   # Using tr:
   tr -d '\r' < input.sh > output.sh && mv output.sh input.sh
   ```

3. Add or update `.gitattributes` at the repo root to enforce normalization:

   ```gitattributes
   # Normalize all text files to LF in the repository
   * text=auto eol=lf

   # Explicitly mark shell scripts
   *.sh text eol=lf
   *.bash text eol=lf

   # Windows-specific files can keep CRLF
   *.bat text eol=crlf
   *.cmd text eol=crlf
   ```

4. Re-normalize the index after adding `.gitattributes`:

   ```bash
   git add --renormalize .
   git commit -m "chore: normalize line endings to LF"
   ```

5. On CI, verify the post-checkout state:

   ```bash
   grep -rlU $'\r' . --include='*.sh' | head -20
   ```

## Validation

- Re-run the failing script after conversion.
- Confirm `file script.sh` reports "ASCII text" or "UTF-8 text" without
  "CRLF line terminators".
- Run `grep -c $'\r' script.sh` and confirm it prints `0`.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain line-ending
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Line ending mismatch (CRLF vs LF)
- Build: line ending mismatch (crlf vs lf)
- syntax error near unexpected token
- GitHub Actions line ending mismatch (crlf vs lf)
- faultline explain line-ending


---

*Generated from [playbooks/bundled/log/build/line-ending.yaml](../../../playbooks/bundled/log/build/line-ending.yaml). Do not edit directly — run `make docs-generate`.*
