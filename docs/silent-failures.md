# Silent / Misleading Failures

Faultline detects a class of CI failure that most tools miss: **silent failures**, where CI appears to succeed but important work was silently skipped, suppressed, or never performed.

## What is a silent failure?

A silent failure occurs when a CI job reports a green status, but the log shows evidence that real work did not happen as expected.

Examples:

- A test command ran but discovered **zero tests** — the runner exited 0, but nothing was actually validated
- An artifact upload step ran but **no files were found** — no artifact was produced despite the "success"
- A command failure was **suppressed** with `|| true` or `set +e` — the real exit code was hidden
- A step used **`continue-on-error: true`** — an error was silently swallowed
- A cache restore **failed** but the job continued — repeated cache misses degrade CI performance
- A critical step was **skipped** by a condition — the build, test, or deploy step never ran

These are hard to catch by looking at CI status alone. Faultline surfaces them as first-class diagnoses.

## Why it matters

Silent failures cause:

- **False confidence** — reviewers and automation see green when work was incomplete
- **Regression blind spots** — tests never ran, so broken code ships
- **Invisible technical debt** — CI "works" but builds slower, uploads nothing, deploys nothing
- **Agent unreliability** — automated fix systems act on incomplete signals

## How to detect silent failures

Run `faultline analyze` on any CI log. Silent findings are included automatically alongside normal diagnoses.

```bash
faultline analyze build.log --json
```

Example JSON output when a silent failure is detected and no other playbook matched:

```json
{
  "faultline_status": "failure",
  "failure_class": "silent_failure",
  "failure_id": "zero-tests-executed",
  "findings": [
    {
      "id": "zero-tests-executed",
      "class": "silent_failure",
      "severity": "high",
      "confidence": "high",
      "explanation": "A test command appeared to run, but no tests were discovered or executed.",
      "evidence": [
        "npm test",
        "No tests found"
      ]
    }
  ]
}
```

When a normal playbook match is also present, the silent findings appear in the `findings` array alongside the primary `results`.

## --fail-on-silent

Pass `--fail-on-silent` to exit non-zero whenever a silent failure is detected, even when the CI run appears successful:

```bash
faultline analyze build.log --fail-on-silent
```

This is useful in:

- Post-build CI gates that check for misleading success signals
- Agent workflows that need to distinguish real success from silent failure
- Pre-merge checks that enforce test completeness

Without `--fail-on-silent`, silent findings are reported but Faultline's exit behavior is unchanged.

## Built-in detectors

Faultline includes six built-in silent-failure detectors. All are deterministic: no ML, no external calls, no guesswork.

| Detector | Signal | Severity |
|---|---|---|
| `ignored-exit-code` | `\|\| true`, `set +e`, `failed but continuing` | high |
| `continue-on-error` | `continue-on-error: true`, `allow_failure: true` | high |
| `zero-tests-executed` | Test command + `No tests found`, `0 tests`, etc. | high |
| `artifact-missing` | Upload step + `Skipping upload`, `no files found with the provided path`, etc. | high |
| `cache-miss-non-fatal` | Cache step + `Cache not found`, `Failed to restore cache`, etc. | medium |
| `skipped-critical-step` | `Skipping step due to condition` + critical domain keyword | high |

Each detector uses conservative AND-logic where possible: both a context signal (e.g. "an upload step ran") and a failure signal (e.g. "no files found") must be present.

## Playbooks

Each detector has a corresponding playbook with diagnosis, fix steps, and prevention notes:

- [`ignored-exit-code`](../playbooks/bundled/log/silent/ignored-exit-code.yaml)
- [`continue-on-error`](../playbooks/bundled/log/silent/continue-on-error.yaml)
- [`zero-tests-executed`](../playbooks/bundled/log/silent/zero-tests-executed.yaml)
- [`artifact-missing`](../playbooks/bundled/log/silent/artifact-missing.yaml)
- [`cache-miss-non-fatal`](../playbooks/bundled/log/silent/cache-miss-non-fatal.yaml)
- [`skipped-critical-step`](../playbooks/bundled/log/silent/skipped-critical-step.yaml)

## JSON schema

Silent findings extend the standard `faultline analyze --json` output without breaking existing consumers:

| Field | Description |
|---|---|
| `faultline_status` | `"failure"` when any finding (normal or silent) is present |
| `failure_class` | `"silent_failure"` when silent is the primary signal |
| `failure_id` | Detector ID of the primary silent finding |
| `findings[]` | Array of all silent findings (id, class, severity, confidence, explanation, evidence) |

Existing fields (`results`, `context`, `fingerprint`, etc.) are unchanged.

## Limitations

- **Conservative heuristics**: detectors prefer precision over recall. They will miss some silent failures.
- **No arbitrary code execution**: detectors are purely string-pattern based on log content.
- **No external calls**: all detection is local and offline.
- **No public extension system yet**: the detector interface is internal. Custom detectors are not yet supported via a public API. This will be addressed in a future iteration.
- **Log-level only**: silent detection runs on log text. It does not inspect CI configuration files or secrets.

## What's next

Planned improvements:

1. **More detector coverage** — detect silent linter/security-scan failures and empty deployment targets
2. **CI config cross-referencing** — correlate log signals with the workflow YAML to catch skipped steps more accurately
3. **Public detector interface** — allow custom org-level silent failure rules via a declarative config (not yet implemented)
