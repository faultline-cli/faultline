# no space left on device in CI builds

**Target search query:** `no space left on device CI build ENOSPC`

## Error snippet

```text
write /tmp/build-output.tar: no space left on device
ENOSPC: no space left on device, write
```

## What this error means

The host or container filesystem ran out of disk space. CI cannot keep writing build output, caches, or artifacts once the filesystem is full.

## Fix steps

1. Check free space with `df -h` and identify large paths with `du -sh`.
2. Prune Docker images, containers, and build cache with `docker system prune -af --volumes`.
3. Remove large build artifacts and language caches.
4. If cleanup is not enough, increase runner disk size or split the workload into smaller jobs.

## How Faultline detects it

Faultline maps this failure to `disk-full`.

Primary signals:

- `no space left on device`
- `ENOSPC`
- `errno 28`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [context deadline exceeded and read timeout in CI](../network/context-deadline-exceeded-read-timeout.md)
- [Cannot connect to the Docker daemon at unix:///var/run/docker.sock](../docker/cannot-connect-to-the-docker-daemon.md)