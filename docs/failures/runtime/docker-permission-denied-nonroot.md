# Docker permission denied running as non-root user

**Playbook ID:** `docker-permission-denied-nonroot`
**Category:** runtime
**Severity:** high
**Tags:** `docker`, `runtime`, `permission`, `nonroot`, `user`

## What this failure means

A Docker container process running as a non-root user encountered a permission denied error while trying to access files or directories owned by root or another user.

## Common log signals

```text
permission denied
cannot open
cannot read
Access denied
Operation not permitted
running as user
uid 1000
chown
```

## Diagnosis

Docker containers often run as non-root for security. Permission denied errors occur when:

- The container runs as a non-root user (e.g., via `USER` directive in Dockerfile) but tries to access files owned by root.
- File permissions in the image are too restrictive (e.g., `chmod 600` instead of `644`).
- Mount volumes from the host have the wrong ownership or permissions.
- The application writes to directories it does not own (e.g., `/app` owned by root).
- The container tries to write to a read-only filesystem.

The error typically appears as `permission denied`, `Cannot write to...`, or `Operation not permitted`.

## Fix steps

1. Identify which files or directories the error is accessing:

   ```bash
   docker logs <container-id> | grep "Permission denied"
   ```

2. Check the file ownership and permissions in the Dockerfile:

   ```dockerfile
   RUN useradd -m -u 1000 appuser
   WORKDIR /app
   COPY --chown=appuser:appuser . .
   USER appuser
   ```

3. Ensure writable directories are owned by the non-root user:

   ```dockerfile
   RUN mkdir -p /app/logs /app/cache && \
       chown -R appuser:appuser /app/logs /app/cache && \
       chmod 755 /app/logs /app/cache
   USER appuser
   ```

4. If mounting volumes, ensure consistent permissions:

   ```bash
   # On host before mounting
   sudo chown 1000:1000 /host/path
   sudo chmod 755 /host/path

   # In docker run or docker-compose
   docker run -v /host/path:/app/data <image>
   ```

5. Verify the container starts without permission errors:

   ```bash
   docker run --rm <image>
   ```

## Validation

- `docker logs <container-id>` shows no "Permission denied" errors.
- The container runs without exiting or crashing.
- Application writes to expected log or data directories successfully.

## Likely files to inspect

- `Dockerfile`
- `docker-compose.yml`
- `.github/workflows/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain docker-permission-denied-nonroot
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker permission denied running as non-root user
- Runtime: docker permission denied running as non-root user
- Operation not permitted
- faultline explain docker-permission-denied-nonroot
- Docker docker permission denied running as non-root user


---

*Generated from [playbooks/bundled/log/runtime/docker-permission-denied-nonroot.yaml](../../../playbooks/bundled/log/runtime/docker-permission-denied-nonroot.yaml). Do not edit directly — run `make docs-generate`.*
