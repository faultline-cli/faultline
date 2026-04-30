# Build output artifact not found at expected path

**Playbook ID:** `build-output-path-mismatch`
**Category:** build
**Severity:** medium
**Tags:** `build`, `output`, `artifact`, `path`, `dist`, `target`

## What this failure means

A CI step that consumes the output of a prior build step cannot find the
expected artifact at the configured path. The build tool produced output
at a different location than the consumer expects, or the build step failed
silently without producing output.

## Common log signals

```text
dist.*not found
build.*not found
target.*not found
output directory.*does not exist
No such file or directory.*dist/
No such file or directory.*build/
artifact.*not found
expected.*output.*path
```

## Diagnosis

Build output path mismatches occur when:
1. The build tool's output directory changed in an upgrade
2. The CI config references a hardcoded path that differs from what the build
   tool actually produces
3. A build configuration (`outDir`, `outputPath`, `target_dir`) was changed
   locally but not updated in CI scripts
4. The build step exited non-zero but CI continued due to missing error
   propagation

Investigate by listing expected paths after the build step:

```bash
# Add a debug step before the consuming step
ls -la dist/ || echo "dist/ not found"
find . -name "*.whl" -o -name "*.jar" -o -name "*.tar.gz" | head -20
```

## Fix steps

1. First confirm the build step actually ran and succeeded:

   ```bash
   # Check exit code of the build step
   echo "Build exit: $?"
   ```

2. Identify where the build tool actually wrote its output:

   ```bash
   find . -newer package.json -name "*.js" -not -path "*/node_modules/*" | head -20
   ```

3. Align one of the two sides:
   - Update the build tool config to write to the expected path:

     ```json
     // webpack.config.js
     output: { path: path.resolve(__dirname, 'dist') }
     ```

     ```xml
     <!-- Maven pom.xml -->
     <build><directory>${project.basedir}/target</directory></build>
     ```

   - Update the CI step to reference the actual output path:

     ```yaml
     - name: Upload artifact
       uses: actions/upload-artifact@v4
       with:
         path: target/release/mybinary   # was: dist/mybinary
     ```

4. If using a multi-stage Dockerfile, ensure the `COPY --from=builder` path
   matches where the builder stage places its output.

5. For Go: `go build -o ./bin/app ./cmd/app` — the `-o` flag controls the
   exact output path.

## Validation

- Add a `ls -la <expected-path>` step before the consuming step and confirm
  the artifact is present.
- Re-run the pipeline end to end.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain build-output-path-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Build output artifact not found at expected path
- Build: build output artifact not found at expected path
- No such file or directory.*build/
- GitHub Actions build output artifact not found at expected path
- faultline explain build-output-path-mismatch


---

*Generated from [playbooks/bundled/log/build/build-output-path-mismatch.yaml](../../../playbooks/bundled/log/build/build-output-path-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
