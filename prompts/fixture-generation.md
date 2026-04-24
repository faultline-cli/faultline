# Fixture Generation

Use this procedure to produce high-fidelity CI log fixtures for a specific playbook.

Called from: `agents/skills/fixture-generation`

## Inputs Required Before Starting

- target playbook ID and its `match.any`, `match.none`, `base_score`
- nearest confusable neighbor ID and its `match.any` phrases
- at least one real-world log sample or `faultline explain <id>` output as a reference anchor

## Required Variants

Produce all three variants for every playbook being defended:

| Variant | Purpose | Match expectation |
|---------|---------|-------------------|
| **canonical** | Direct failure, error and context clearly visible | top-1 ≥ base_score |
| **noisy** | Same failure buried in multi-step, multi-tool output | top-1 ≥ base_score |
| **near-miss** | Similar failure that must NOT trigger this playbook | NOT top-1 |

## Log Structure

Every canonical and noisy fixture must follow this structure:

1. **Setup phase** — checkout step, environment setup, tool version output (at least one step completes successfully)
2. **Execution phase** — the command or step that causes the failure
3. **Failure point** — the actual error with full context: error code, message, stack trace, or command output as the tool would produce it
4. **Post-failure noise** — at least two lines of CI runner summary, cleanup output, or unrelated step output after the error

Near-miss fixtures must use the same tool and command as the canonical, but with a different root cause that the `match.none` exclusion catches.

## Realism Rules

Every log must:
- include timestamps or step markers matching the CI system (e.g. `##[error]`, `::error::`, `$ `)
- include at least one warning or unrelated output line not relevant to the failure
- reproduce error text verbatim as the real tool emits it (npm, gradle, pytest, docker, etc.)
- name the working directory (`/home/runner/work/...`, `/builds/...`) where the tool would show it

**Minimal one-liner logs are rejected.** A fixture must have enough surrounding output that the matching engine must discriminate signal from noise.

Bad:
```
ERROR: module not found
```

Good:
```
Run npm ci
npm warn old lockfile
npm ERR! code EUSAGE
npm ERR! `npm ci` can only install packages when your package.json and
npm ERR! package-lock.json or npm-shrinkwrap.json are in sync.
npm ERR! Please update your lock file with `npm install` before continuing.
npm ERR! Missing: eslint@^8.0.0 from lock file
##[error]Process completed with exit code 1.
```

## Output Format

For each variant, return:

**Fixture path:** `fixtures/minimal/<playbook-id>[-variant].yaml`

**Fixture content:**
```yaml
id: <playbook-id>[-variant]
log: |
  <full CI log>
expected_playbook: <playbook-id>   # omit for near-miss
```

**Match signal:** which `match.any` phrase anchors the match, and which `match.none` phrase the near-miss relies on.

## Constraints

- Do not simplify or strip noise from canonical and noisy fixtures
- Do not fabricate error formats that cannot be verified against real tool documentation
- Do not use the same log lines in canonical and noisy; vary the surrounding context
- Near-miss must be genuinely similar, not trivially different
