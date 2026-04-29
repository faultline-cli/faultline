# Multi-stage Docker build artifact not copied correctly

**Playbook ID:** `multistage-build-missing-artifact`
**Category:** build
**Severity:** high
**Tags:** `docker`, `multistage`, `COPY`, `artifact`, `build-stage`, `missing`

## What this failure means

A multi-stage Docker build failed because a `COPY --from=<builder>` instruction
references a file or directory that was not produced by the builder stage.
The artifact is either at a different path than expected, was not compiled
due to an earlier error, or the builder stage name is incorrect.

## Common log signals

```text
COPY --from
failed to copy files
error copying files
path does not exist in build context
unable to find file
binary.*not found
artefact.*missing
no artifacts in
```

## Diagnosis

In a multi-stage Dockerfile, the final stage copies built artifacts from an
earlier stage. The `COPY --from` instruction fails silently or loudly when
the source path does not exist in the named builder stage.

Common causes:
1. The build tool outputs to a path that differs from the `COPY --from` source
2. The builder stage's build command failed but the Docker build continued
3. A stage name typo (`--from=buider` instead of `--from=builder`)
4. The build tool is language-specific and changed its output directory

Debug by running the builder stage interactively:

```bash
# Build only the builder stage
docker build --target builder -t debug-builder .

# Inspect what was produced
docker run --rm debug-builder find /app -type f | head -20
docker run --rm debug-builder ls -la /go/bin/
```

## Fix steps

1. Run the builder stage standalone and list its contents to find where the
   artifact was actually written:

   ```bash
   docker build --target builder -t debug-builder .
   docker run --rm debug-builder find / -name "myapp" -not -path "*/proc/*"
   ```

2. Update the `COPY --from` source path to match the actual artifact location:

   ```dockerfile
   # Before (wrong path)
   COPY --from=builder /app/server /app/server

   # After (correct path found via debug)
   COPY --from=builder /go/bin/server /app/server
   ```

3. If the builder stage compile step may fail silently, add an explicit
   validation:

   ```dockerfile
   RUN go build -o /go/bin/server ./cmd/server && \
       test -f /go/bin/server || (echo "Build output missing" && exit 1)
   ```

4. Verify stage names are consistent between definition and reference:

   ```dockerfile
   # STAGE DEFINITION
   FROM golang:1.22 AS builder    # <-- name: "builder"

   # COPY REFERENCE (must match exactly)
   COPY --from=builder /go/bin/server .   # correct
   COPY --from=build /go/bin/server .     # WRONG: "build" != "builder"
   ```

5. For intermediate stages, use named stages explicitly:

   ```dockerfile
   FROM golang:1.22 AS build-deps
   FROM build-deps AS builder
   FROM scratch AS final
   COPY --from=builder /go/bin/server /server
   ```

## Validation

- Run `docker build .` and confirm it exits 0.
- Run `docker run --rm <image> <binary-cmd>` and confirm the artifact
  executes correctly.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain multistage-build-missing-artifact
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Multi-stage Docker build artifact not copied correctly
- Build: multi-stage docker build artifact not copied correctly
- path does not exist in build context
- GitHub Actions multi-stage docker build artifact not copied correctly
- faultline explain multistage-build-missing-artifact
- Docker multi-stage docker build artifact not copied correctly


---

*Generated from [playbooks/bundled/log/build/multistage-build-missing-artifact.yaml](../../playbooks/bundled/log/build/multistage-build-missing-artifact.yaml). Do not edit directly — run `make docs-generate`.*
