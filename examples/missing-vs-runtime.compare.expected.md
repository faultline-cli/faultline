# Faultline Compare
- Previous: `stdin`
- Current: `stdin`
## Summary
- top diagnosis changed from missing-executable to runtime-mismatch
- 1 new top-diagnosis evidence line(s) appeared
- 1 prior top-diagnosis evidence line(s) disappeared
## Diagnosis
- Previous: `missing-executable` (Required executable or runtime binary missing) 57%
- Current: `runtime-mismatch` (Python, Ruby, or Go runtime version mismatch) 83%
- Diagnosis changed: yes
- Artifact status changed: no
## Evidence Changes
- Added: Go Version: go1.26.0
- Removed: exec /__e/node20/bin/node: no such file or directory
## Fix Step Changes
- Added: Identify the required version from the repository source of truth:
- Added: If the failure came from a dependency upgrade, confirm the new minimum
- Added: Pin the same version in the workflow or container image that runs the
- Added: Re-run the build locally with the pinned runtime so you can confirm the
- Removed: Confirm the binary exists in the active runner or container image:
- Removed: Identify which command or runtime path the failing step expects to execute.
- Removed: If the failure happens only inside a containerized CI step, make sure the container image includes the same runtime that the host-based workflow expects.
- Removed: If the path is hard-coded, update the workflow or wrapper script to use the actual installed location.
- Removed: Install the missing package in the job image or switch to an image that already contains the required runtime.
