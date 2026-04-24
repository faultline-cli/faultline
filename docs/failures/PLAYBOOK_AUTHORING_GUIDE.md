# Playbook Authoring Guide

## Quick Start: Adding a New Playbook

This guide helps you add a new deterministic failure playbook following Faultline's established patterns.

## Step 1: Validate the Failure Signature

Before authoring, answer these questions:

1. **Is this a distinct root cause?** (Not just wording variations of an existing playbook)
2. **Is there a clear, reproducible pattern?** (Exact error codes, command names, or specific paths)
3. **Can you extract 2-3 concrete log lines as evidence?** (Not generic "error" messages)
4. **Is there a well-defined fix?** (Not "check your config" without specifics)

If you can't answer YES to all four, the playbook isn't ready yet.

## Step 2: Choose the Category

Place your playbook in the appropriate directory under `playbooks/bundled/log/`:

| Category | Path | Use For |
|----------|------|---------|
| build | `log/build/` | Compilation, dependency install, build tool failures |
| auth | `log/auth/` | Authentication, secrets, permissions |
| network | `log/network/` | DNS, TLS, connectivity, timeouts |
| runtime | `log/runtime/` | Container/process runtime, OS-level failures |
| test | `log/test/` | Test runner failures, test-specific issues |
| deploy | `log/deploy/` | Deployment, health checks, orchestration |
| ci | `log/ci/` | CI/CD platform config, workflow issues |

## Step 3: Create the Playbook YAML

Use this template:

```yaml
id: your-playbook-id
title: Human-readable title (60 chars)
category: build  # or auth, network, etc.
severity: high   # or medium
base_score: 0.95 # 0.85-0.95 range (don't over-rank)
tags: [tag1, tag2, tag3]
stage_hints: [build]  # Or: [test, deploy]

summary: |-
  One-sentence summary for CLI output. Describe what went wrong.

diagnosis: |-
  ## Diagnosis

  Explain the root cause in 2-3 paragraphs:
  - What is the failure signature?
  - When does it occur?
  - What are the common causes?

  Keep it concrete and specific, not generic.

fix: |-
  ## Fix steps

  1. First, concrete action (include command if applicable)
  2. Second action (e.g., verify with command)
  3. Continue with actionable steps

  Use short code blocks for exact commands, not scripts.

validation: |-
  ## Validation

  - How to verify the root cause was the real problem
  - Command to confirm the fix worked
  - Sign that the issue is resolved

why_it_matters: |-
  ## Why it matters

  Sentence explaining impact and frequency.

  ## Prevention

  - Best practice 1
  - Best practice 2

match:
  any:
    - "exact error message 1"
    - "exact error message 2"
    - "specific.*pattern"
  none:
    - "exclude pattern 1"  # False positive guard
    - "exclude pattern 2"

workflow:
  likely_files:
    - config.yaml
    - .github/workflows/*.yml
  local_repro:
    - command 1
    - command 2
  verify:
    - verification command
```

## Step 4: Create a Realistic Fixture

Create a log file: `internal/engine/testdata/fixtures/your-playbook-id.log`

**Rules for good fixtures:**

- ✅ **Real log output** from actual CI runs (sanitized)
- ✅ **Minimal but complete** - include the error and surrounding context
- ✅ **Distinctive** - unique enough to distinguish from similar failures
- ✅ **No sensitive data** - remove secrets, tokens, personal info, internal paths
- ⚠️ **Avoid generic patterns** - don't use placeholder text like `<hostname>`

**Example good fixture** (from `npm-eresolve-conflict.log`):

```
npm ERR! code ERESOLVE
npm ERR! ERESOLVE could not resolve
npm ERR!
npm ERR! While resolving: my-app@1.0.0
npm ERR! Found: react@18.2.0
npm ERR! Could not resolve dependency:
npm ERR! peer react@"^16.0.0" required by some-component@2.1.0
```

✅ **Specific**: URLs, tool names, version numbers
✅ **Trimmed**: Just the error section, not a 500-line build log
✅ **Sanitized**: Generic app name, no internal IDs

## Step 5: Add Regex Patterns to Match

Write patterns in the `match.any` section:

- **Exact strings** (preferred): `"npm ERR! code ERESOLVE"`
- **Regex patterns** (as needed): `"error:.*timeout.*seconds"`
  - Use `.*` for variable text
  - Use `\.` to escape dots
  - Keep it simple and readable
  
**Pattern guidelines:**

```yaml
match:
  any:
    # ✅ Good: Exact, tool-specific
    - "npm ERR! code ERESOLVE"
    - "ERR_PNPM_FROZEN_LOCKFILE_CHANGED"
    
    # ✅ Good: Specific regex
    - "psql: error: could not connect"
    - "git@.*Permission denied"
    
    # ❌ Bad: Too generic (matches everything)
    - "error"
    - "Connection.*fail"
    
    # ❌ Bad: Impossible regex
    - "npm.*error.*code.*timeout"  # Doesn't match actual error strings
```

## Step 6: Add Exclusions to Prevent False Positives

The `none:` list **guards against nearby false positives**:

```yaml
match:
  any:
    - "Connection refused"
  none:
    - "authentication failed"  # Different failure, guards against confusion
    - "timeout"                # Similar but distinct cause
    - "ECONNREFUSED"           # More specific error code that should rank higher
```

## Step 7: Add Regression Test

Edit `internal/engine/playbooks_fixtures_test.go`:

```go
tests := []struct {
    name   string
    file   string
    wantID string
}{
    // ... existing tests ...
    {name: "your playbook name", file: "your-playbook-id.log", wantID: "your-playbook-id"},
}
```

## Step 8: Run Tests

```bash
# Test just the fixture
cd /home/jake/workspace/faultline
make test 2>&1 | grep -A5 "TestBundledPlaybookFixtures"

# Test the whole suite
make test

# Check for pattern overlaps
go run ./cmd/playbook-review
```

**Expected outcomes:**

- ✅ Your fixture matches your playbook as the top result
- ✅ Deterministic results (run with `--bayes` and without, same top 3)
- ✅ No new false positives on unrelated logs

## Step 9: Add Negative Test (Optional)

Create a negative fixture to ensure the playbook doesn't have false positives:

```bash
# internal/engine/testdata/fixtures/your-playbook-id-negative.log
# Log output that is SIMILAR but should NOT match your playbook

# Then add to test:
{name: "your playbook negative case", file: "your-playbook-id-negative.log", wantID: "different-playbook-id"},
```

## Step 10: Document

Update `docs/failures/NEW_PLAYBOOKS_EXPANSION.md` with:

- Playbook name and path
- One-sentence description
- Confidence level (HIGH / MEDIUM / LOW)
- Key evidence patterns
- Brief fix summary

## Checklist Before Committing

- [ ] Playbook YAML is valid (runs without parse errors)
- [ ] Fixture is realistic and sanitized
- [ ] Match patterns are precise, not generic
- [ ] Exclusion patterns prevent false positives
- [ ] Regression test passes
- [ ] Deterministic output confirmed (multiple runs same result)
- [ ] No regressions in existing playbook rankings
- [ ] Documentation updated

## Common Pitfalls

### 1. Too Many Generic Patterns

**❌ Wrong:**
```yaml
match:
  any:
    - "error"
    - "failed"
    - "ERROR"
```

**✅ Right:**
```yaml
match:
  any:
    - "npm ERR! code ERESOLVE"
    - "ERESOLVE could not resolve"
```

### 2. Missing Exclusions

**❌ Wrong:**
```yaml
match:
  any:
    - "connection refused"
  # Missing guard against "timeout" and "auth failed"
```

**✅ Right:**
```yaml
match:
  any:
    - "connection refused"
  none:
    - "timeout"
    - "authentication"
    - "ECONNREFUSED"
```

### 3. Vague Fix Steps

**❌ Wrong:**
```
## Fix steps
1. Check your configuration
2. Verify your setup
3. Test again
```

**✅ Right:**
```
## Fix steps
1. Verify the SSH private key exists and has 600 permissions:
   ```bash
   ls -la ~/.ssh/id_rsa
   chmod 600 ~/.ssh/id_rsa
   ```
2. Test SSH connection:
   ```bash
   ssh -T git@github.com
   ```
```

### 4. Overlapping Playbooks

If your fixture matches an existing playbook better than your new one, either:

- Add stronger exclusions to your playbook
- Refine the fixture to be more specific
- Accept that existing playbook is more appropriate
- Reach out for guidance on pattern separation

## Examples to Study

Good playbooks to use as reference:

- `playbooks/bundled/log/build/npm-eresolve-conflict.yaml` - precise error code
- `playbooks/bundled/log/network/dns-enotfound.yaml` - specific network error
- `playbooks/bundled/log/auth/ssh-permission-denied.yaml` - tight AUTH pattern
- `playbooks/bundled/log/build/pnpm-lockfile-missing.yaml` - tool-specific error

## Questions?

Refer to:

- `docs/playbooks.md` - general playbook authoring conventions
- `docs/failures/NEW_PLAYBOOKS_EXPANSION.md` - recent additions and rationale
- Existing playbooks in `playbooks/bundled/log/` - live examples to study
