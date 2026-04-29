# Required executable or runtime binary missing

**Playbook ID:** `missing-executable`
**Category:** build
**Severity:** high
**Tags:** `toolchain`, `executable`, `binary`, `path`, `runtime`

## What this failure means

The job tried to launch a required tool or runtime binary, but that executable was missing from the image, runner, or expected path.

## Common log signals

```text
executable file not found in $PATH
exec /
exec: "
command not found
is not recognized as an internal or external command
exit status 127
OCI runtime exec failed
No Rakefile found
```

## Diagnosis

The failing step reached the point where it tried to execute a binary, but the runner could not find that binary in `$PATH` or at the configured absolute path.

Common causes:

- the base image does not include the required tool
- the workflow assumes a runtime path that does not exist inside a container job
- an install step failed or never ran before the command executed
- a wrapper script points at a moved or renamed binary

## Fix steps

1. Identify which command or runtime path the failing step expects to execute.
2. Confirm the binary exists in the active runner or container image:

   ```bash
   command -v <tool>
   ls -l <expected-path>
   ```

3. Install the missing package in the job image or switch to an image that already contains the required runtime.
4. If the path is hard-coded, update the workflow or wrapper script to use the actual installed location.
5. If the failure happens only inside a containerized CI step, make sure the container image includes the same runtime that the host-based workflow expects.

## Validation

- `command -v <tool>` returns a valid path inside the same environment that runs the failing step.
- Re-run the job and confirm the step starts the executable instead of failing before launch.

## Likely files to inspect

- `Dockerfile`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `docker-compose*.yml`
- `scripts/**/*`


## Run Faultline

```bash
faultline analyze build.log
faultline explain missing-executable
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Required executable or runtime binary missing
- Build: required executable or runtime binary missing
- is not recognized as an internal or external command
- GitHub Actions required executable or runtime binary missing
- faultline explain missing-executable


---

*Generated from [playbooks/bundled/log/build/missing-executable.yaml](../../playbooks/bundled/log/build/missing-executable.yaml). Do not edit directly — run `make docs-generate`.*
