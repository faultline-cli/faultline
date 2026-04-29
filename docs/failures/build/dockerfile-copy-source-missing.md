# Docker COPY failed source file missing from build context

**Playbook ID:** `dockerfile-copy-source-missing`
**Category:** build
**Severity:** high
**Tags:** `docker`, `build`, `context`, `copy`, `dockerfile`

## What this failure means

Docker COPY or ADD instruction failed because the source file does not exist in the build context. The file may be excluded by `.dockerignore` or the path is incorrect.

## Common log signals

```text
COPY failed
COPY failed: stat
failed: file not found in build context
ADD failed: file not found
no such file or directory
Step [0-9]+ : COPY
```

## Diagnosis

Docker builds use a build context that includes only files and directories specified or not excluded. Common causes:

- The source file path in `COPY` or `ADD` instruction does not exist relative to the build context root.
- The file is excluded by `.dockerignore` but the Dockerfile tries to copy it anyway.
- The build context path is wrong (e.g., using `/app` or an absolute path instead of a relative path).
- A build step generates the file but the `COPY` runs before that step completes.
- The repository structure does not match what the Dockerfile expects.

The error appears as `COPY failed: file not found` or `file not found in build context`.

## Fix steps

1. Verify the exact relative path exists in the build context:

   ```bash
   docker build --no-cache -t test . 2>&1 | head -20
   ls -la <source-file>
   ```

2. Check `.dockerignore` for patterns that may be excluding the file:

   ```bash
   cat .dockerignore
   ```

3. Update the `COPY` instruction in the Dockerfile to use the correct relative path:

   ```dockerfile
   # Before (wrong)
   COPY /app/src/main.go .

   # After (correct)
   COPY src/main.go .
   ```

4. If the file is generated during build, ensure it's created in an earlier stage and either copied from that stage (multi-stage) or not excluded from the context.

5. Re-run `docker build` with explicit context path:

   ```bash
   docker build -t myimage:latest .
   ```

## Validation

- `docker build` completes without `COPY failed` errors.
- List files in the running container to confirm they were copied:

   ```bash
   docker run --rm myimage:latest ls -la /app
   ```

## Likely files to inspect

- `Dockerfile`
- `.dockerignore`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain dockerfile-copy-source-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker COPY failed source file missing from build context
- Build: docker copy failed source file missing from build context
- failed: file not found in build context
- GitHub Actions docker copy failed source file missing from build context
- faultline explain dockerfile-copy-source-missing
- Docker docker copy failed source file missing from build context


---

*Generated from [playbooks/bundled/log/build/dockerfile-copy-source-missing.yaml](../../playbooks/bundled/log/build/dockerfile-copy-source-missing.yaml). Do not edit directly — run `make docs-generate`.*
