# Required executable or runtime binary missing

- ID: `missing-executable`
- Confidence: 55%
- Category: build
- Severity: high
- Score: 2.10
- Detector: log

## Summary

The job tried to launch a required tool or runtime binary, but that executable was missing from the image, runner, or expected path.

## Evidence

- exec /__e/node20/bin/node: no such file or directory

## Confidence Breakdown

- reported confidence: 55%
- detector score: 2.10

## Suggested Fix

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
