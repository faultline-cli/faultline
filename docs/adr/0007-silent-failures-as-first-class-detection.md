# ADR 0007: Silent Failures As First-Class Detection

- Status: Accepted
- Date: 2026-04-25

## Context

Faultline historically focused on detecting explicit CI failures — logs with non-zero exit codes, error messages, and clear error signals. However, a critical class of CI failures exists where the job reports success but important work was silently skipped, suppressed, or never performed.

Examples include:

- Tests ran but discovered zero tests (`npm test` exited 0 with no tests found)
- Artifact uploads succeeded but no files matched the glob
- Error handling was suppressed with `|| true` or `set +e`
- A critical step was skipped due to a workflow condition
- Cache or dependency operations failed but were allowed to continue

These are visible only by careful log inspection and cause:

- False confidence in automation and reviews
- Regression blind spots when tests never run
- Invisible technical debt from silent degradation
- Agent unreliability when acting on incomplete signals

The repository history in commit `b7e9a9b` introduced the silent-failure detection class and built-in detectors, documented in [docs/silent-failures.md](../silent-failures.md).

## Decision

Faultline adds `silent_failure` as a first-class `failure_class` alongside normal playbook matches. Silent detection runs automatically on all analysis and surfaces findings in both text and JSON output.

The silent detector model includes:

- A closed set of deterministic, built-in detectors in `internal/silentdetector/`
- Eight conservative detectors using AND-logic where possible (context signal + failure signal)
- Automatic ranking alongside playbook matches by severity and confidence
- Optional `--fail-on-silent` flag to treat detected silent failures as non-zero exit
- JSON schema extension with `failure_class: "silent_failure"` when detected
- Public playbooks for each silence pattern under `playbooks/bundled/log/silent/`

No public detector plugin API or declarative DSL is exposed; detectors are internal and conservative.

## Consequences

- Users get automatic detection of misleading CI success without special flags or configuration
- Operators can use `--fail-on-silent` in post-build validation gates
- Agent workflows can distinguish real success from silent failure via JSON `failure_class`
- Silent findings are ranked alongside normal playbook matches, with deterministic ordering
- Future extensions (repo expectations, CI config cross-referencing, public DSLs) can build on this foundation
- The closed detector set means diagnostic breadth is intentionally bounded; coverage gaps must be addressed through targeted new detectors, not user-supplied rules

## References

- [docs/silent-failures.md](../silent-failures.md)
- [internal/silentdetector/silentdetector.go](../../internal/silentdetector/silentdetector.go)
- [playbooks/bundled/log/silent/](../../playbooks/bundled/log/silent/)
- Git history: `b7e9a9b` add silent_failure class with 6 built-in detectors, `b07893f` finalize silent detector follow-up fixes, `d709932` fix false-positive AND-check overlaps in three detectors
