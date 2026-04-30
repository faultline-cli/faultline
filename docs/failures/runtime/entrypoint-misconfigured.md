# Docker entrypoint or CMD misconfiguration

**Playbook ID:** `entrypoint-misconfigured`
**Category:** runtime
**Severity:** high
**Tags:** `docker`, `entrypoint`, `cmd`, `container`, `startup`, `exec-form`, `shell-form`

## What this failure means

A Docker container exits immediately after start because its `ENTRYPOINT`
or `CMD` is misconfigured. Common causes include a missing or non-executable
binary, an `exec format error` from an architecture mismatch, or incorrect
syntax mixing shell form and exec form.

## Common log signals

```text
no such file or directory
exec format error
not an executable
entrypoint.*not found
container.*exited immediately
container.*exit code 127
container.*exit code 126
permission denied.*entrypoint
```

## Diagnosis

Exit codes immediately after container start indicate entrypoint problems:
- **Exit 126**: permission denied — binary exists but is not executable
- **Exit 127**: command not found — ENTRYPOINT path does not exist in image
- **Exec format error**: binary compiled for wrong architecture (e.g.,
  arm64 binary on amd64 runner)

Inspect the entrypoint:

```bash
# Show entrypoint and cmd in image
docker inspect <image> --format '{{json .Config.Entrypoint}} {{json .Config.Cmd}}'

# Verify the binary exists in the image
docker run --rm --entrypoint /bin/sh <image> -c "ls -la /app/server"

# Check the binary's architecture
docker run --rm --entrypoint /bin/sh <image> -c "file /app/server"
```

## Fix steps

1. **Binary not found (exit 127)**:
   Verify the `COPY` or build step placed the binary at the path `ENTRYPOINT`
   references:

   ```dockerfile
   # Ensure binary is at /app/server not /usr/local/bin/server
   COPY --from=builder /go/bin/server /app/server
   ENTRYPOINT ["/app/server"]   # must match above
   ```

2. **Not executable (exit 126)**:
   Ensure the binary has execute permissions:

   ```dockerfile
   COPY --from=builder /go/bin/server /app/server
   RUN chmod +x /app/server
   ```

3. **Exec format error** (architecture mismatch):
   Build the binary for the target platform:

   ```bash
   # Cross-compile for linux/amd64
   GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

   # Or use multi-platform build
   docker buildx build --platform linux/amd64,linux/arm64 -t myimage .
   ```

4. **Entry point syntax**: use exec form (JSON array) not shell form for
   signals and PID 1 correctness:

   ```dockerfile
   # Prefer exec form (receives SIGTERM directly)
   ENTRYPOINT ["/app/server"]
   CMD ["--port", "8080"]

   # Shell form wraps in /bin/sh -c (SIGTERM goes to shell, not app)
   ENTRYPOINT /app/server    # AVOID for servers
   ```

5. **Shell dependencies in ENTRYPOINT**: if using shell form or a shell
   script, ensure the shell is present in the image:

   ```dockerfile
   FROM scratch   # no shell available — use exec form only
   ENTRYPOINT ["/app/server"]
   ```

## Validation

- Run `docker run --rm <image> true` and confirm exit code 0.
- Run the container with a shell override to inspect contents:
  `docker run --rm --entrypoint sh <image> -c "ls -la /app"`
- Re-run the CI deploy step and confirm the container stays running.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain entrypoint-misconfigured
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker entrypoint or CMD misconfiguration
- Runtime: docker entrypoint or cmd misconfiguration
- starting container process caused
- faultline explain entrypoint-misconfigured
- Docker docker entrypoint or cmd misconfiguration


---

*Generated from [playbooks/bundled/log/runtime/entrypoint-misconfigured.yaml](../../../playbooks/bundled/log/runtime/entrypoint-misconfigured.yaml). Do not edit directly — run `make docs-generate`.*
