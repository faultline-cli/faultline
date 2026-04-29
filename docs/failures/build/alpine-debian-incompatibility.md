# Alpine vs Debian/Ubuntu package or binary incompatibility

**Playbook ID:** `alpine-debian-incompatibility`
**Category:** build
**Severity:** medium
**Tags:** `alpine`, `debian`, `ubuntu`, `musl`, `glibc`, `docker`, `linux`, `apk`, `apt`

## What this failure means

A CI build or container fails because it mixes Alpine Linux (musl libc) and
Debian/Ubuntu (glibc) artifacts. Pre-compiled binaries or native extensions
linked against glibc crash on Alpine's musl libc, and package names may
differ between `apk` (Alpine) and `apt` (Debian/Ubuntu).

## Common log signals

```text
musl
libc.musl
glibc
error loading shared libraries
not a valid ELF file
apk add.*not found
no such package
apk: command not found
```

## Diagnosis

Alpine Linux uses musl libc instead of glibc. Most pre-compiled binaries
distributed on the internet are glibc-linked and will not run on Alpine.
Additionally, Alpine packages are managed with `apk`, not `apt`, and
many package names differ.

Common failure patterns:
1. A glibc-linked binary runs inside an Alpine container
2. Node.js native addons compiled on glibc cannot load on musl
3. A Dockerfile switches base image from `ubuntu` to `alpine` without
   updating `apt-get install` commands to `apk add`

Check the binary's libc requirement:

```bash
# Inside the container
ldd /app/server
# If output shows "/lib/x86_64-linux-gnu/libc.so.6" → glibc binary
# Alpine's musl: "/lib/ld-musl-x86_64.so.1"

file /app/server
# Shows "dynamically linked" and the interpreter path
```

## Fix steps

1. **Use a glibc base image** if the application requires glibc-linked
   libraries:

   ```dockerfile
   # Switch from Alpine
   FROM alpine:3.19         # musl
   # To Debian slim for a smaller glibc image
   FROM debian:bookworm-slim  # glibc
   # Or Ubuntu minimal
   FROM ubuntu:24.04
   ```

2. **Compile for musl** if staying on Alpine:

   ```bash
   # Go: produces statically-linked binary (no libc dependency)
   CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server
   ```

   ```dockerfile
   # Multi-stage: compile on glibc builder, copy to Alpine
   FROM golang:1.22 AS builder
   RUN CGO_ENABLED=0 go build -o /server ./cmd/server

   FROM alpine:3.19
   COPY --from=builder /server /server   # works: no glibc dependency
   ```

3. **Fix package names** when switching between Alpine and Debian:

   | Debian          | Alpine              |
   |-----------------|---------------------|
   | `apt-get`       | `apk`               |
   | `libssl-dev`    | `openssl-dev`       |
   | `libpq-dev`     | `postgresql-dev`    |
   | `build-essential` | `build-base`      |
   | `ca-certificates` | `ca-certificates` |

   ```dockerfile
   # Debian/Ubuntu
   RUN apt-get update && apt-get install -y --no-install-recommends \
       libssl-dev ca-certificates && rm -rf /var/lib/apt/lists/*

   # Alpine
   RUN apk add --no-cache openssl-dev ca-certificates
   ```

4. For Node.js native addons on Alpine, add the build tools and rebuild:

   ```dockerfile
   FROM node:20-alpine
   RUN apk add --no-cache python3 make g++
   RUN npm install --build-from-source
   ```

## Validation

- Run `ldd /app/server` inside the container and confirm it resolves to
  the correct libc (`ld-musl-x86_64.so.1` on Alpine).
- Run the container and confirm the application starts.
- Re-run the CI pipeline.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain alpine-debian-incompatibility
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Alpine vs Debian/Ubuntu package or binary incompatibility
- Build: alpine vs debian/ubuntu package or binary incompatibility
- error loading shared libraries
- GitHub Actions alpine vs debian/ubuntu package or binary incompatibility
- faultline explain alpine-debian-incompatibility
- Docker alpine vs debian/ubuntu package or binary incompatibility


---

*Generated from [playbooks/bundled/log/build/alpine-debian-incompatibility.yaml](../../playbooks/bundled/log/build/alpine-debian-incompatibility.yaml). Do not edit directly — run `make docs-generate`.*
