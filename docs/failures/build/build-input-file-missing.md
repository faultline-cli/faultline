# Required build input file missing

**Playbook ID:** `build-input-file-missing`
**Category:** build
**Severity:** medium
**Tags:** `build`, `file`, `path`, `artifact`, `input`

## What this failure means

A build or lint step referenced a file that was never generated, was copied to the wrong path, or is missing from the repository checkout.

## Common log signals

```text
cp: cannot stat
cannot stat
withBinaryFile: does not exist
does not exist (No such file or directory)
{error,{enoent,
file not found
```

## Diagnosis

The failing command started normally, but one of the files it expected to read or copy was not present at the referenced path.

Common causes:

- a generated file was not created before the step ran
- a path changed in the repository without updating the build script
- a cache or artifact restore step did not populate the expected file
- the file exists locally but is not committed or not included in the CI context

## Fix steps

1. Identify the exact missing path from the failing line.
2. Confirm whether the file should be committed, generated, restored from cache, or copied from another job.
3. Add a directory listing immediately before the failing step to verify the expected file layout:

   ```bash
   pwd
   ls -R <parent-directory>
   ```

4. If the file is generated, ensure the generation step runs earlier in the same job.
5. If the file comes from artifacts or caches, verify the producing job publishes it and the consuming job restores it to the same path.

## Validation

- Re-run the build or lint step after restoring or generating the missing file.
- Confirm the referenced path exists in the CI workspace before the failing command starts.

## Likely files to inspect

- `Makefile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `docker-compose*.yml`
- `scripts/**/*`


## Run Faultline

```bash
faultline analyze build.log
faultline explain build-input-file-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Required build input file missing
- Build: required build input file missing
- does not exist (No such file or directory)
- GitHub Actions required build input file missing
- faultline explain build-input-file-missing


---

*Generated from [playbooks/bundled/log/build/build-input-file-missing.yaml](../../playbooks/bundled/log/build/build-input-file-missing.yaml). Do not edit directly — run `make docs-generate`.*
