# Required executable or runtime binary missing

- ID: `missing-executable`
- Confidence: 69%
- Category: build
- Severity: high
- Score: 6.15
- Detector: log

## Summary

The job tried to launch a required tool or runtime binary, but that executable was missing from the image, runner, or expected path.

## Evidence

- exec /__e/node20/bin/node: no such file or directory

## Confidence Breakdown

- reported confidence: 69%
- detector baseline: 2.27
- final reranked score: 6.15
- conservative prior: +0.30
- +0.72 detector confidence supports the candidate
- +0.35 token overlap between evidence and playbook patterns supports the candidate
- +0.07 broader explicit signal coverage supports the candidate

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
