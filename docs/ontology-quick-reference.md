# CI Failure Ontology: Quick Reference

This is a condensed reference for the Faultline CI Failure Ontology. For detailed information, see:
- `docs/ontology.md` - Full design and principles
- `docs/ontology-examples.md` - Six detailed examples
- `docs/ontology-implementation.md` - Implementation roadmap

---

## Five-Level Hierarchy

```
Domain (broad area)
  ↓
Class (family within domain)
  ↓
Mode (specific root cause)
  ↓
Evidence Pattern (deterministic signals)
  ↓
Remediation Strategy (fix approach)
```

---

## Domains (11 total)

| Domain | Purpose | Examples |
|--------|---------|----------|
| `dependency` | Package resolution, lockfiles | npm, pnpm, yarn, python |
| `runtime` | Executables, permissions, resources | Node, Python, OOM, segfault |
| `container` | Image build/pull, daemon | Docker, buildkit, manifest |
| `auth` | Credentials, tokens, secrets | docker-login, git-ssh, AWS creds |
| `network` | DNS, connectivity, TLS | connection-refused, timeout, cert |
| `ci-config` | Workflow validation, setup | GitHub Actions syntax, env vars |
| `test-runner` | Test framework, coverage | pytest, jest, flaky tests |
| `database` | Connection, migration, data | postgres, mysql, isolation |
| `filesystem` | Paths, permissions, disk space | COPY, missing files, disk-full |
| `platform` | Provider-specific failures | K8s, Terraform, scheduling |
| `source` | Code quality, compilation | syntax errors, type errors |

---

## Classes (Examples)

**dependency:**
- `lockfile-drift` (npm, pnpm, yarn)
- `version-conflict` (peer deps, resolution)
- `missing-package` (uninstalled module)
- `cache-poisoning` (corrupted cache)

**runtime:**
- `missing-executable` (command not found)
- `interpreter-mismatch` (version conflict)
- `permission-denied` (insufficient privilege)
- `resource-exhaustion` (OOM, disk full)

**auth:**
- `registry-auth` (docker, npm registry)
- `credential-mismatch` (wrong secret)
- `token-expiry` (rotated credentials)

**network:**
- `tls-validation` (certificate errors)
- `connection-refused` (service not listening)
- `dns-resolution` (hostname mismatch)
- `rate-limiting` (too many requests)

**ci-config:**
- `env-not-persisted` (GitHub Actions env)
- `workflow-syntax` (YAML validation)
- `resource-allocation` (runner limits)

---

## Minimal Playbook Record Schema

```yaml
# Required
id: unique-playbook-id
title: Human-readable title
category: build  # For backwards compatibility

# NEW: Ontology Classification
domain: dependency
class: lockfile-drift
mode: npm-ci-requires-package-lock
aliases: []
severity: high
confidence_baseline: 0.95

# Required
match:
  any: [...]

# Required
summary: |
  One-line summary
diagnosis: |
  Detailed diagnosis with ### Heading
fix: |
  Step-by-step fix
validation: |
  How to validate

# Existing optional fields work unchanged
workflow: {...}
```

**Key point:** Ontology fields are **additive**. Existing playbooks work unchanged.

---

## Confidence Baseline Guidelines

| Range | Meaning | Requirements |
|-------|---------|--------------|
| 0.95-1.00 | Very High | 2+ required signals, clear false positive exclusions, multiple positive fixtures |
| 0.85-0.94 | High | 1-2 required signals, 1+ optional signal, 1+ negative fixture |
| 0.70-0.84 | Medium | Requires review; may need stronger signals or more fixtures |
| < 0.70 | Low | Consider redesign or splitting into multiple modes |

---

## Evidence Pattern Format

```yaml
evidence:
  required:
    - log.contains: "exact error message"
    - log.regex: "pattern.*for.*regex"
  
  optional:
    - log.contains: "supporting signal"
    - delta.signal: dependency.lockfile.changed
  
  exclusions:
    - log.contains: "false positive pattern"
  
  confidence: 0.92
  
  false_positive_risks:
    - "edge case that could cause confusion"
```

**Signal types:**
- `log.contains:<text>` - String match
- `log.regex:<pattern>` - Regex (Rust syntax)
- `log.absent:<text>` - NOT present
- `delta.signal:<id>` - Baseline comparison
- `context.stage:<stage>` - build/test/deploy/runtime
- `file.exists:<path>` - File in repo
- `named_signal:<alias>` - Team-defined signal

---

## Mode Naming Convention

**Format:** `<tool>-<specific-error>` or `<operation>-<blocker>`

✓ Good:
- `npm-ci-requires-package-lock`
- `python-venv-not-activated`
- `postgres-connection-refused-startup-lag`

✗ Bad:
- `npm-error` (too generic)
- `runtime-failure` (symptom-based)
- `ci-issue` (not specific)

---

## Remediation Strategies (Core Set)

| Strategy | Purpose | Usage |
|----------|---------|-------|
| `align-lockfile` | Regenerate & commit lockfile | dependency/lockfile-drift |
| `install-missing-tool` | Install missing binary | runtime/missing-executable |
| `activate-venv` | Create & activate venv | runtime/interpreter-mismatch |
| `wait-for-service-readiness` | Retry, health checks, timeout | database/service-not-ready |
| `use-service-hostname` | Update DNS/discovery | network/connection-refused |
| `configure-token-scope` | Rotate credential, add scopes | auth/credential-mismatch |
| `persist-ci-env-correctly` | Use platform env syntax | ci-config/env-not-persisted |
| `clear-corrupt-cache` | Delete and regenerate | dependency/cache-poisoning |

---

## When to Author a New Playbook

### Step 1: Choose Domain

Where does the failure originate?

- `dependency` = resolver, lockfile, version
- `runtime` = executor, environment, binary
- `auth` = credential validation
- `network` = transport, DNS, TLS
- **NOT:** where it's detected, but where it originates

### Step 2: Choose Class

Does an existing class fit?

```
dependency: lockfile-drift ✓ (fits npm-ci-lockfile)
runtime: missing-executable ✓ (fits missing-node)
test-runner: ??? (need new class if truly novel)
```

Reuse unless genuinely distinct.

### Step 3: Choose Mode

What is the specific error signature?

```
mode: npm-ci-requires-package-lock
      ↑ tool/operation
      ↑ specific-blocker
```

Each mode = one playbook. One root cause, one fix strategy.

### Step 4: Define Evidence

```yaml
# Minimum: 2 required signals
required:
  - log.contains: "npm ci can only install packages when"
  - log.regex: "package\\.json and package-lock\\.json.*(?:not in sync|out of sync)"

# Add exclusions for known false positives
exclusions:
  - log.contains: "ENOENT"
  - log.contains: "ERR! code E"
```

### Step 5: Write Remediation

Clear, numbered steps:

```yaml
steps:
  - "Run npm install locally"
  - "Review the diff"
  - "Commit and push"
  - "Re-run CI"
```

### Step 6: Add Fixtures

- **1+ positive:** Log that SHOULD match
- **1+ negative:** Similar log that should NOT match (confusable edge case)

---

## Coverage Reporting

### Generate Report

```bash
faultline coverage                                    # All domains
faultline coverage --domain=dependency                # One domain
faultline coverage --gaps                             # Show gaps
faultline coverage --format=json > coverage.json      # Machine-readable
```

### Interpret Output

```
dependency: 22 playbooks (deep)
  - lockfile-drift: 5 playbooks ✓ (full coverage)
  - version-conflict: 4 playbooks ✓ (full coverage)
  - cache-poisoning: 0 playbooks ✗ (gap!)

runtime: 18 playbooks (deep)
  - missing-executable: 4 playbooks ✓
  - interpreter-mismatch: 3 playbooks ✓
  - permission-denied: 2 playbooks ✓

Confidence Distribution:
  0.95-1.00: 38 playbooks (high ✓)
  0.85-0.94: 42 playbooks (good ✓)
  0.70-0.84: 18 playbooks (medium, needs review)
  < 0.70: 5 playbooks (low, consider redesign)
```

---

## Editor Checklist: New Playbook Review

Before committing a playbook:

- [ ] **Domain correct?** (operational subsystem, not symptom)
- [ ] **Class justified?** (reused or new with reasoning)
- [ ] **Mode specific?** (describes one root cause, not family)
- [ ] **2+ required signals?** (deterministic, no weak matchers)
- [ ] **Confidence honest?** (0.95+ only if truly tight)
- [ ] **Exclusions documented?** (known false positives listed)
- [ ] **1+ positive fixture?** (log that should match)
- [ ] **1+ negative fixture?** (confusable edge case)
- [ ] **Related modes noted?** (nearby discriminators)
- [ ] **Tests pass?** (`make test`)
- [ ] **Review pass?** (`make review`)

---

## Migration FAQ

**Q: Will ontology fields break my existing playbooks?**

A: No. Ontology fields are optional and additive. Playbooks without them continue to work.

**Q: What if I'm not ready to adopt ontology?**

A: You don't have to. Old playbooks remain valid. Coverage reporting just won't include your playbook.

**Q: Can I use ontology in external packs?**

A: Yes. External packs can add ontology metadata. Faultline will recognize it.

**Q: What if I disagree with a domain assignment?**

A: File an issue. Ontology is designed to evolve based on feedback.

**Q: How often do I run coverage reports?**

A: Suggested: weekly in CI, quarterly review with team. No required cadence.

---

## Key Principles

1. **Determinism first.** Same playbook + same log = same result always.
2. **One mode per playbook.** Each playbook describes one specific root cause.
3. **Evidence must be tight.** Prefer 2+ required signals over weak optional matches.
4. **Remediation must be actionable.** Steps should lead directly to verification.
5. **Fixtures are proof.** Positive AND negative fixtures validate the ontology.
6. **Reuse strategies.** Multiple modes can share `remediation.strategy`.
7. **Coverage drives roadmap.** Use reports to identify priority gaps.

---

## Documents

| Document | Purpose |
|----------|---------|
| `docs/ontology.md` | Complete hierarchy, schema, philosophy |
| `docs/ontology-examples.md` | 6 detailed complete examples |
| `docs/ontology-implementation.md` | Phase-by-phase implementation roadmap |
| `docs/ontology-quick-reference.md` | This document (condensed reference) |
| `docs/playbooks.md` | Existing playbook authoring guide (still valid) |

---

## Getting Started

### For Contributors

1. Read `docs/ontology.md` section "Contributor Guidance"
2. Choose domain, class, mode for your playbook
3. Follow the checklist above
4. Add 1+ positive + 1+ negative fixtures
5. Run `make test` and `make review`

### For Maintainers

1. Read `docs/ontology-implementation.md` Phase 2
2. Tag 10-15 playbooks per batch
3. Run tests after each batch
4. Generate coverage report monthly
5. Identify gaps and plan expansions

### For Enterprise Users

1. Run `faultline coverage --format=json`
2. Analyze domains, classes, confidence distribution
3. Identify gaps relevant to your CI systems
4. Request playbooks or contribute patches

---

**Version:** 1.0  
**Status:** Ready for adoption  
**Last Updated:** 2026-04-25
