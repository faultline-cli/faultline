# Docker volume mount failure or inaccessible mount

**Playbook ID:** `volume-mount-issue`
**Category:** runtime
**Severity:** medium
**Tags:** `docker`, `volume`, `mount`, `bind-mount`, `permissions`, `path`

## What this failure means

A Docker container failed to start or operate correctly because a volume
or bind mount is inaccessible. The host path does not exist, the container
path is invalid, the mount is read-only when the container requires write
access, or the UID/GID inside the container does not match the mounted path's
ownership.

## Common log signals

```text
mount.*failed
cannot mount
error response from daemon.*mount
volume.*not found
No such file or directory.*volume
bind mount
invalid mount config
invalid volume specification
```

## Diagnosis

Volume mount failures have several distinct categories:
1. **Path does not exist**: the host path in a bind mount was not created
2. **Wrong absolute path**: relative paths in `-v` are not supported on all
   systems
3. **Permissions mismatch**: directory owned by root but container runs as
   non-root
4. **Read-only volume**: container writes to a volume mounted `ro`
5. **Docker Desktop on macOS/Windows**: paths outside `/Users` (macOS) or
   `C:\Users` (Windows) are not shared by default

Inspect the mount configuration:

```bash
docker inspect <container> --format '{{json .Mounts}}' | python3 -m json.tool
```

## Fix steps

1. Verify the host path exists before running the container:

   ```bash
   ls -la /path/to/host/dir
   mkdir -p /path/to/host/dir   # create if missing
   ```

2. Always use absolute paths in volume mounts:

   ```bash
   # WRONG: relative path
   docker run -v ./data:/app/data myimage

   # CORRECT: absolute path
   docker run -v "$(pwd)/data":/app/data myimage
   ```

3. Fix ownership mismatches by setting the UID/GID:

   ```bash
   # Option A: run container as the host user
   docker run --user "$(id -u):$(id -g)" -v "$(pwd)/data":/app/data myimage

   # Option B: set correct ownership on the host directory
   sudo chown -R 1000:1000 ./data
   ```

   ```dockerfile
   # Option C: in Dockerfile, create the directory as the correct user
   RUN mkdir -p /app/data && chown app:app /app/data
   ```

4. If the volume is mounted read-only but needs write access, change the
   mount mode:

   ```yaml
   # docker-compose.yml
   volumes:
     - ./data:/app/data:rw    # change from :ro to :rw
   ```

5. For Docker Desktop on macOS, add the directory to the allowed file
   sharing paths: Docker Desktop > Settings > Resources > File sharing.

6. For tmpfs-mounted paths in CI (e.g., `/run`, `/tmp`), ensure the
   container is not trying to bind-mount those paths from outside.

## Validation

- Run `docker run --rm -v <mount_spec> busybox ls <container_path>` to
  verify the mount is accessible.
- Re-run the failing CI job and confirm the container starts.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain volume-mount-issue
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Docker volume mount failure or inaccessible mount
- Runtime: docker volume mount failure or inaccessible mount
- error response from daemon.*mount
- faultline explain volume-mount-issue
- Docker docker volume mount failure or inaccessible mount


---

*Generated from [playbooks/bundled/log/runtime/volume-mount-issue.yaml](../../playbooks/bundled/log/runtime/volume-mount-issue.yaml). Do not edit directly — run `make docs-generate`.*
