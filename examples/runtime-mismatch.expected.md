# Python, Ruby, or Go runtime version mismatch

- ID: `runtime-mismatch`
- Confidence: 89%
- Category: build
- Severity: high
- Score: 3.75
- Detector: log
- Stage: build

## Summary

The installed Python, Ruby, or Go runtime does not satisfy the version
constraint declared by the project. The build is using a language runtime
outside the version range expected by the repository or dependency set.
