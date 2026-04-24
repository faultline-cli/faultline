# CI Failure Ontology for Faultline

## Executive Summary

This document defines the **CI Failure Ontology**, a canonical taxonomy that maps CI log evidence into deterministic failure classes, diagnoses, remediation guidance, and coverage metrics.

The ontology transforms Faultline from:

> "a collection of playbooks"

into:

> "a structured CI failure knowledge system"

It enables:

- **deterministic playbook classification** across a consistent taxonomy
- **coverage gap analysis** by domain, class, and failure mode
- **fixture organization** with clear ownership and depth expectations
- **test planning** with confidence distribution and negative test coverage
- **contributor guidance** for consistent ontology adoption
- **enterprise reporting** on CI failure patterns, trends, and remediation effectiveness

---

## Design Principles

### Determinism First

- The ontology is **read-only metadata**, not a runtime classifier
- Same playbook + same log always produces the same result
- No probabilistic assignment; each playbook explicitly declares its ontology membership
- Classification serves humans and agents, not the matching engine

### Backwards Compatible

- Existing playbooks continue to work unchanged
- Ontology metadata is **additive and optional** for existing playbooks
- Migration happens incrementally; no flag-day rewrites
- Old playbooks without ontology tags are valid; they simply won't report coverage metrics

### Lightweight and Readable

- YAML/JSON compatible
- One ontology record per playbook (no separate lookup tables)
- Clear naming conventions for human understanding
- Suitable for tooling, analytics, and future enterprise reporting

### Extensible

- Domains, classes, and remediation strategies can grow
- New failure modes discovered in practice can be added deterministically
- Teams can extend with org-specific metadata overlays

---

## Five-Level Hierarchy

### Level 1: Failure Domain

The **broad area of failure**. Typically matches the operational subsystem where the problem manifests.

| Domain | Examples | Typical Stages |
|--------|----------|-----------------|
| `dependency` | lockfile mismatches, version conflicts, missing packages | build |
| `runtime` | permission errors, OOM, resource exhaustion, segfaults | build, test, deploy, runtime |
| `container` | image pulls, image builds, daemon issues | build, deploy |
| `auth` | credential mismatches, token expiry, scope errors | build, deploy |
| `network` | DNS failures, connection timeouts, firewalls | build, test, deploy, runtime |
| `ci-config` | workflow syntax, step validation, resource limits | meta |
| `test-runner` | test framework issues, coverage gates, isolation | test |
| `database` | connection failures, migration errors, isolation | test, deploy |
| `filesystem` | missing files, permission denied, disk full | build, test, deploy, runtime |
| `platform` | scheduler errors, provider-specific failures | deploy |
| `source` | code quality, type errors, compilation | build |

### Level 2: Failure Class

The **specific family within a domain**. Describes the detection mechanism or root cause family.

Examples:

| Class | Domain | Example Modes |
|-------|--------|----------------|
| `lockfile-drift` | dependency | `npm-ci-lockfile-outdated`, `pnpm-frozen-lockfile-outdated` |
| `missing-executable` | runtime | `node-missing`, `python-interpreter-missing`, `go-compiler-missing` |
| `interpreter-mismatch` | runtime | `python-version-mismatch`, `node-version-mismatch` |
| `service-not-ready` | database | `postgres-not-listening`, `mysql-connection-refused` |
| `registry-auth` | auth | `docker-registry-403`, `npm-registry-401` |
| `tls-validation` | network | `certificate-unknown-authority`, `hostname-mismatch` |
| `wrong-working-directory` | filesystem | `npm-enoent-relative-path`, `pytest-conftest-not-found` |
| `cache-poisoning` | dependency | `npm-cache-corrupted`, `maven-local-repo-broken` |
| `env-not-persisted` | ci-config | `github-actions-env-not-carried-to-next-step` |
| `resource-overallocation` | ci-config | `circleci-oom-resource-class` |

### Level 3: Failure Mode

The **concrete root cause**. Specific error signature and fix strategy that distinguishes this mode from nearby classes.

Examples:

| Mode | Class | Root Cause |
|------|-------|-----------|
| `npm-ci-requires-package-lock` | lockfile-drift | npm ci fails when package-lock.json missing or out of sync |
| `pnpm-frozen-lockfile-outdated` | lockfile-drift | pnpm --frozen-lockfile fails when .pnpm-lock.yaml is stale |
| `python-module-installed-to-wrong-interpreter` | interpreter-mismatch | Package installed to Python 2 but code runs Python 3 |
| `docker-copy-source-missing` | filesystem | Dockerfile COPY fails because source file doesn't exist in build context |
| `github-actions-env-not-persisted` | env-not-persisted | GitHub Actions step exports env var but next step doesn't see it |
| `postgres-service-not-ready` | service-not-ready | CI tries to connect before postgres has finished startup |
| `x509-unknown-authority` | tls-validation | Client doesn't trust root CA in certificate chain |

### Level 4: Evidence Pattern

The **deterministic signals used to match** the failure mode. Each evidence pattern defines how to identify the failure from log text.

**Fields:**

- `required`: List of signals that MUST be present for a confident match
- `optional`: List of signals that increase confidence when present
- `exclusions`: List of signals that rule out this mode (conflict or false positive)
- `confidence`: How certain we are of this detection (0.0-1.0)
- `false_positive_risks`: Documented confusable scenarios

Example:

```yaml
evidence:
  required:
    - log.contains: "npm ci can only install packages when"
    - log.contains: "package.json and package-lock.json"
  optional:
    - log.contains: "in sync"
    - delta.signal: dependency.lockfile.changed
  exclusions:
    - log.contains: "npm ERR! cipm"  # Different error path
    - log.contains: "manifest unknown"  # Registry issue, not lockfile
  confidence: 0.95
  false_positive_risks:
    - npm ci can emit this message during recovery from partial install
    - lockfile regeneration mid-install can produce ambiguous messages
```

### Level 5: Remediation Strategy

The **fix family**. Groups related remediation approaches and provides step guidance.

| Strategy | Domain | Examples |
|----------|--------|----------|
| `align-lockfile` | dependency | regenerate lockfile, commit, verify sync |
| `install-missing-tool` | runtime | install binary, add to PATH, verify availability |
| `activate-venv` | runtime | source venv, set PYTHONPATH, verify interpreter |
| `wait-for-service-readiness` | database | add retry logic, check health endpoint, increase timeout |
| `use-service-hostname` | network | update DNS name, use service discovery, verify connectivity |
| `configure-token-scope` | auth | rotate credential, add required scopes, verify in CI secret |
| `persist-ci-env-correctly` | ci-config | use ::set-output or ::set-env, export to GITHUB_ENV, verify next step |
| `clear-corrupt-cache` | dependency | delete cache directory, regenerate, re-run |

---

## Ontology Record Schema

### Required Fields

Each failure mode record should include:

```yaml
id: <unique-id>
domain: <domain>
class: <class>
mode: <mode>
aliases: [<string>, ...]
severity: <critical|high|medium|low>
confidence_baseline: <float 0.0-1.0>
```

### Evidence Fields

```yaml
evidence:
  required: [<signal>, ...]
  optional: [<signal>, ...]
  exclusions: [<signal>, ...]
  confidence: <float>
  false_positive_risks: [<string>, ...]
```

### Signal Format

Signals are deterministic matchers:

- `log.contains:<text>` - Raw string match
- `log.regex:<pattern>` - Regex match (Rust syntax)
- `log.absent:<text>` - String does NOT appear
- `delta.signal:<id>` - Baseline comparison signal
- `delta.absent:<id>` - Baseline signal NOT present
- `context.stage:<stage>` - Build stage (build, test, deploy, runtime)
- `file.exists:<path>` - File exists in repo
- `named_signal:<alias>` - Team-defined or shared signal

Examples:

```yaml
evidence:
  required:
    - log.contains: "npm ci can only install packages"
    - log.regex: "package.json and package-lock.json.*not in sync|out of sync"
  optional:
    - log.contains: "run `npm install` to update the lock"
    - delta.signal: dependency.npm.lockfile.changed
  exclusions:
    - log.contains: "ENOENT: no such file or directory"
    - file.exists: package-lock.json
```

### Remediation Fields

```yaml
remediation:
  strategy: <strategy>
  summary: <one-line description>
  steps:
    - <numbered step>
    - <numbered step>
  validation:
    - <verification step>
    - <verification step>
  docs_link: <optional URL>
```

### Fixture Fields

```yaml
fixtures:
  positive:
    - id: <fixture-id>
      source: <source or description>
      confidence: <expected confidence>
  negative:
    - id: <fixture-id>
      description: <why this should NOT match>
      confuses_with: <other mode id>
```

### Relations Fields

```yaml
related_modes:
  - id: <nearby mode>
    reason: "easily confused because both involve lockfiles"
  - id: <alternative>
    reason: "same symptom but different toolchain"
```

---

## Complete Ontology Schema (YAML)

```yaml
# Failure Mode Definition
id: npm-ci-lockfile-mismatch
title: "npm ci lockfile mismatch"
category: build  # For backwards compatibility

# Ontology Classification
domain: dependency
class: lockfile-drift
mode: npm-ci-requires-package-lock
aliases:
  - "npm-ci-lockfile-outdated"
  - "npm lockfile not in sync"

# Urgency and Confidence
severity: medium
confidence_baseline: 0.95

# Evidence Matching
evidence:
  required:
    - log.contains: "npm ci can only install packages when"
    - log.regex: "package\\.json and package-lock\\.json.*not in sync|out of sync"
  optional:
    - log.contains: "in sync"
    - log.contains: "run `npm install`"
    - delta.signal: dependency.npm.lockfile.changed
    - delta.signal: delta.dependency.changed
  exclusions:
    - log.contains: "ENOENT: no such file or directory"
    - log.contains: "ERR! code ENOTFOUND"
  confidence: 0.95
  false_positive_risks:
    - "Lockfile regeneration mid-install can produce ambiguous intermediate messages"
    - "Partial npm ci success followed by cleanup can echo the sync error"

# Root Cause Context
root_cause: |
  npm ci enforces strict lockfile fidelity: package.json and package-lock.json 
  must be in sync. If they drift (from direct edits, workspace changes, or 
  generation errors), npm ci refuses to install.

# Remediation Path
remediation:
  strategy: align-lockfile
  summary: "Regenerate and commit package-lock.json"
  steps:
    - "Run `npm install` locally to regenerate package-lock.json"
    - "Verify the file changed and diffs are sensible"
    - "Commit the updated package-lock.json"
    - "Ensure package-lock.json is not listed in .gitignore"
    - "If using workspaces, regenerate from the workspace root"
  validation:
    - "Run `npm ci` locally and confirm it succeeds"
    - "Re-run the CI job"
    - "Confirm package-lock.json checksum is stable after install"

# Fixtures
fixtures:
  positive:
    - id: npm-ci-lockfile-out-of-sync-1
      source: "npm 10.x on Node 18, package-lock.json 2 revisions old"
      confidence: 0.95
    - id: npm-ci-lockfile-workspace-mismatch
      source: "npm workspaces, lockfile regenerated with npm 9.x but CI uses npm 10.x"
      confidence: 0.90
  negative:
    - id: npm-ci-enoent-file-missing
      description: "ENOENT error from missing package.json, not lockfile sync"
      confuses_with: npm-enoent-package-json
    - id: npm-registry-auth-timeout
      description: "Registry connectivity issue masquerading as sync error"
      confuses_with: registry-auth

# Related Failure Modes
related_modes:
  - id: pnpm-lockfile-missing
    reason: "Same root cause (lockfile mismatch) but different package manager"
  - id: yarn-lockfile-mismatch
    reason: "Same pattern for Yarn, different remediation paths"
  - id: dependency-drift
    reason: "Broader class; this mode is one specific manifestation"

# Coverage and Documentation
coverage:
  domains: [dependency]
  classes: [lockfile-drift]
  modes: [npm-ci-requires-package-lock]
  depth: deep  # Deeply covered with multiple fixtures
doc_link: "https://docs.npmjs.com/cli/v10/commands/npm-ci"
```

---

## Practical Integration Example

### Playbook with Full Ontology

A bundled playbook evolved to include full ontology:

```yaml
# From: playbooks/bundled/log/build/npm-ci-lockfile.yaml
id: npm-ci-lockfile
title: npm ci lockfile mismatch
category: build

# NEW: Ontology Classification
domain: dependency
class: lockfile-drift
mode: npm-ci-requires-package-lock
aliases: []
severity: medium
confidence_baseline: 0.95

# Existing match logic (unchanged)
match:
  any:
    - npm ci can only install packages when your package.json and package-lock.json
    - npm error `npm ci` can only install packages when your
    - package.json and package-lock.json are in sync
    - missing package-lock.json
    - run `npm install` to generate a lockfile

# Existing diagnosis (unchanged)
summary: |
  `npm ci` found a missing or out-of-sync `package-lock.json`.
diagnosis: |
  ## Diagnosis
  `npm ci` installs strictly from the lockfile. If `package.json` and `package-lock.json` disagree, CI fails.
fix: |
  ## Fix steps
  1. Regenerate the lockfile locally:
     ```bash
     npm install
     ```
  2. Commit the updated `package-lock.json`.
  3. Make sure `package-lock.json` is not ignored.
validation: |
  ## Validation
  - Run `npm ci` locally.
  - Re-run the CI job.

# Existing workflow metadata (unchanged)
workflow:
  likely_files:
    - package.json
    - package-lock.json
```

**Migration note:** This playbook works exactly as before. The ontology fields are **purely additive** for cataloging and reporting.

---

## Coverage Reporting Model

### What Can Be Reported

With ontology metadata, Faultline can generate:

#### 1. Domain Coverage
```
Domain Coverage Summary
=======================
dependency:        22 playbooks (deep)
runtime:           18 playbooks (deep)
auth:              12 playbooks (medium)
network:            9 playbooks (medium)
ci-config:          7 playbooks (shallow)
test-runner:       15 playbooks (deep)
database:           5 playbooks (medium)
container:          8 playbooks (medium)
filesystem:         6 playbooks (shallow)
platform:           3 playbooks (shallow)
source:             3 playbooks (shallow)
```

#### 2. Class Coverage
```
Class Coverage by Domain
=========================
dependency:
  - lockfile-drift:            5 playbooks ✓
  - version-conflict:          3 playbooks ✓
  - missing-package:           2 playbooks ○ (shallow, high false-positive risk)
  - cache-corruption:          1 playbook  ✗ (needs negative fixtures)
  - peer-dependency-mismatch:  1 playbook  ○

runtime:
  - missing-executable:        4 playbooks ✓
  - interpreter-mismatch:      3 playbooks ✓
  - permission-denied:         2 playbooks ✓
  - oom-killed:                2 playbooks ✓
  - disk-full:                 1 playbook  ✓
  ...
```

#### 3. Fixture Depth
```
Fixture Depth by Mode
=====================
npm-ci-lockfile:         5 positive + 3 negative ✓✓✓
pnpm-lockfile-missing:   2 positive + 1 negative ✓✓
yarn-lockfile-mismatch:  1 positive + 0 negative ✓
python-venv-mismatch:    4 positive + 5 negative ✓✓✓
docker-auth:             6 positive + 4 negative ✓✓✓
```

#### 4. Confidence Distribution
```
Confidence Distribution
=======================
0.95-1.00: 038 playbooks (high confidence)
0.85-0.94: 042 playbooks (good confidence)
0.70-0.84: 018 playbooks (medium confidence, review needed)
0.50-0.69: 005 playbooks (low confidence, needs work)
```

#### 5. Remediation Coverage
```
Remediation Strategy Distribution
==================================
align-lockfile:                 5 playbooks
install-missing-tool:          8 playbooks
activate-venv:                 2 playbooks
wait-for-service-readiness:    3 playbooks
configure-token-scope:         4 playbooks
clear-corrupt-cache:           2 playbooks
use-service-hostname:          1 playbook
...
```

#### 6. Gap Analysis
```
Coverage Gaps
=============
HIGH PRIORITY:
  - cache-poisoning (dependency): 0 playbooks ✗✗✗
  - job-cancellation (ci-config): 0 playbooks ✗
  - container-oom (container): 0 playbooks ✗

MEDIUM PRIORITY:
  - rate-limiting (network): 1 playbook (shallow) ○
  - secret-rotation (auth): 0 playbooks ✗
  - test-parallelism-limit (test-runner): 1 playbook ○

LOW CONFIDENCE:
  - jest-worker-crash (test-runner): confidence 0.62 ⚠
  - flaky-test (test-runner): confidence 0.58 ⚠
```

### Reporting Implementation

Create a new `faultline coverage` command:

```bash
faultline coverage --format=text       # Human-readable summary
faultline coverage --format=json       # Machine-readable metrics
faultline coverage --domain=dependency # Filter by domain
faultline coverage --depth=shallow     # Show only shallow coverage areas
faultline coverage --gaps              # Show priority gaps
```

---

## Migration Plan

### Phase 1: Define Ontology (This Document)

- [ ] Publish ontology hierarchy and schema
- [ ] Create contributor guidelines
- [ ] Review with team for alignment

**Duration:** Done

### Phase 2: Tag Existing Playbooks (Batch 1)

- [ ] Assign domain/class/mode to existing 60+ playbooks
- [ ] Add `confidence_baseline` to each
- [ ] Design evidence pattern templates
- [ ] Tag by category in one pass (domain should map cleanly to category)

**Example:**
```bash
# Review bundled playbooks and add ontology metadata
for playbook in playbooks/bundled/log/*/*.yaml; do
  # Extract category from path
  # Open and add: domain, class, mode
  # Run tests to confirm no change in behavior
done
```

**Duration:** 1-2 sprints

### Phase 3: Add Fixture Metadata

- [ ] Tag each fixture with playbook ID and confidence
- [ ] Identify positive vs. negative fixtures
- [ ] Add coverage depth labels (deep/medium/shallow)
- [ ] Remove or improve low-confidence fixtures

**Duration:** 1 sprint

### Phase 4: Build Coverage Reporting

- [ ] Implement `faultline coverage` command
- [ ] Generate domain/class/mode coverage tables
- [ ] Add confidence distribution reporting
- [ ] Publish automated coverage reports in CI

**Duration:** 1 sprint

### Phase 5: Gap Analysis and Planning

- [ ] Use coverage reports to identify priority gaps
- [ ] Design 3-5 new playbooks from gap analysis
- [ ] Plan next batch of coverage expansion
- [ ] Update ROADMAP with ontology-driven priorities

**Duration:** 1-2 sprints

### Phase 6: Continuous Improvement

- [ ] Review new playbooks for ontology alignment
- [ ] Monitor confidence distribution in production
- [ ] Add new domains/classes as needed
- [ ] Maintain coverage reports as living docs

**Duration:** Ongoing

### Backwards Compatibility Guarantee

- Playbooks **without** ontology metadata remain fully functional
- Old playbooks continue to match and rank exactly as before
- No breaking changes to YAML schema
- `faultline analyze` output unchanged unless `--coverage` flag used
- Older packs and external playbooks don't need to adopt ontology

---

## Contributor Guidance

When authoring a new playbook, follow this process:

### Step 1: Choose Domain

**Question:** Where in the CI/deployment pipeline does this failure manifest?

Options:

- **`dependency`**: Package/module resolution, lockfiles, version conflicts
- **`runtime`**: Runtime errors, permission issues, resource exhaustion
- **`container`**: Image building, registry, daemon
- **`auth`**: Credentials, tokens, secrets
- **`network`**: DNS, connectivity, TLS
- **`ci-config`**: Workflow validation, pipeline setup, resource limits
- **`test-runner`**: Test framework, coverage, isolation
- **`database`**: Connection, migration, test data
- **`filesystem`**: paths, permissions, disk space
- **`platform`**: Provider-specific scheduling, quotas, deployment
- **`source`**: Code quality, compilation, type checking

**Decision rule:** The domain should match the operational subsystem where the error originates, not where it's detected.

- ✓ `docker-auth` → `auth` (origin: credential validation)
- ✗ `docker-auth` → `container` (wrong; it's not a container build issue)
- ✓ `postgres-connection-refused` → `database` (origin: DB service)
- ✗ `postgres-connection-refused` → `network` (wrong; specific to DB service layer)

### Step 2: Choose Class

**Question:** What is the root cause family?

Examine existing classes in the domain. If your failure fits an existing class, reuse it.

Example: You're writing a new playbook for `pip install` lockfile errors.

- Domain: `dependency`
- Existing classes: `lockfile-drift`, `version-conflict`, `missing-package`, `cache-corruption`
- Your failure: lockfile hash mismatch in pip
- **Decision:** Reuse `lockfile-drift` class (same root cause family)

If no existing class fits:

- Name it descriptively: `<adjective>-<noun>`
- Document why it's distinct from nearby classes
- Suggest it for ontology expansion
- Use `<domain>-/<class>` pattern: `dependency/pip-hash-validation`

### Step 3: Choose Mode

**Question:** What is the specific error signature that distinguishes this mode?

Your mode should be:

- **Specific:** Not "python error" but "python-venv-not-activated"
- **Actionable:** Points directly to the fix
- **Unique:** Only this root cause produces this exact signal
- **Testable:** You can write a fixture that matches this mode and nothing else

Format: `<tool>-<specific-error>` or `<operation>-<blocker>`

Examples:

- ✓ `npm-ci-requires-package-lock` (specific tool, specific requirement)
- ✓ `github-actions-env-not-persisted` (specific platform, specific behavior)
- ✗ `npm-error` (too generic)
- ✗ `ci-failure` (not specific)

### Step 4: Define Evidence Pattern

**Question:** What log signals deterministically identify this failure?

Use the evidence schema:

```yaml
evidence:
  required:
    - log.contains: "exact text from log"
    - log.regex: "pattern.*for.*variable.*parts"
  optional:
    - delta.signal: dependency.npm.lockfile.changed
  exclusions:
    - log.contains: "nearby false positive signature"
  confidence: 0.92
  false_positive_risks:
    - "this pattern can appear when... [scenario]"
```

**Guidelines:**

- **Required signals:** Must unambiguously point to this mode. Two required signals > one required + many optional.
- **Optional signals:** Boost confidence but shouldn't be present in every instance.
- **Exclusions:** Explicitly list false positives (other modes that might match the same text).
- **Confidence:** Express your certainty (0.95 = very certain, 0.70 = needs work).
- **False positives:** Document known edge cases.

### Step 5: Choose Remediation Strategy

**Question:** What is the fix approach?

Use an existing strategy if possible:

- `align-lockfile` → regenerate and commit
- `install-missing-tool` → add to PATH or CI setup
- `activate-venv` → source environment, set vars
- `wait-for-service-readiness` → add retry logic, health checks, timeout
- `use-service-hostname` → update DNS/service discovery
- `configure-token-scope` → rotate credential, add scopes
- `persist-ci-env-correctly` → use platform-specific env persistence
- `clear-corrupt-cache` → delete and regenerate

If none fit exactly, propose a new strategy:

- Name it as `<operation>-<outcome>`
- Document the steps
- Suggest it for ontology expansion

### Step 6: Write the Playbook (Existing Process Unchanged)

Your playbook now has full context:

```yaml
id: my-new-failure

# Ontology (NEW)
domain: dependency
class: lockfile-drift
mode: pnpm-frozen-lockfile-outdated
aliases: [pnpm-lockfile-stale]
severity: medium
confidence_baseline: 0.93

# Matching (EXISTING)
match:
  any:
    - pnpm install --frozen-lockfile
    - pnpm ERR! .*frozen lockfile

# Content Guidance (EXISTING, now informed by ontology)
summary: |
  pnpm install --frozen-lockfile failed because pnpm-lock.yaml is outdated.

diagnosis: |
  ## Diagnosis
  
  pnpm --frozen-lockfile enforces strict lockfile fidelity. When pnpm-lock.yaml
  is older than package.json, pnpm refuses to install.

fix: |
  ## Fix steps
  
  1. Regenerate the lockfile:
     ```bash
     pnpm install
     ```
  2. Commit pnpm-lock.yaml
  3. Push and re-run CI

workflow:
  likely_files:
    - package.json
    - pnpm-lock.yaml
  local_repro:
    - pnpm install --frozen-lockfile
  verify:
    - pnpm install --frozen-lockfile
```

### Step 7: Add Fixtures

- **Positive fixture:** Sample log that should match this mode
- **Negative fixture:** Similar log that should NOT match (e.g., pnpm ERR with different reason)
- **Confidence label:** Mark expected confidence (0.93 in example above)

### Step 8: Review and Test

Run the existing playbook tests:

```bash
make test
make review  # Check for overlaps with other playbooks
```

Ontology consistency checks:

```bash
# Validate ontology assignments (future)
faultline playbooks validate --ontology

# Check coverage impact
faultline coverage --mode=pnpm-frozen-lockfile-outdated
```

---

## Quality Bar Checklist

When evaluating a playbook for ontology consistency:

- [ ] **Domain is correct** - matches the operational subsystem, not the symptom
- [ ] **Class is justified** - fits an existing class or proposes a new one with clear reasoning
- [ ] **Mode is specific** - describes the exact error, not a family
- [ ] **Evidence is tight** - required signals are specific, optional signals are genuine boosters
- [ ] **Confidence is honest** - 0.95+ only if truly tight; otherwise 0.70-0.85
- [ ] **Exclusions are explicit** - false positives are documented
- [ ] **Remediation is actionable** - steps lead to verification
- [ ] **Fixtures are positive + negative** - tests both match and non-match cases
- [ ] **Related modes are noted** - nearby confusables are documented
- [ ] **No overfit** - playbook doesn't over-match on weak signals

---

## Frequently Asked Questions

### Q: If I add ontology metadata to an existing playbook, will the behavior change?

**A:** No. Ontology fields are purely additive metadata. They don't affect matching, ranking, or output unless explicitly enabled with `--coverage` or `--stats` flags.

### Q: What if my playbook matches multiple modes?

**A:** This suggests the playbook should be split. Each playbook should correspond to one mode (one specific error signature). If your log evidence could match multiple root causes, write separate playbooks and use `hypothesis.discriminators` to rank them.

### Q: Can I reuse someone else's mode ID?

**A:** No. Mode IDs must be globally unique within Faultline and all installed packs. Use namespacing if needed: `prod-registry-timeout` for your org's production registry.

### Q: How do I report a new domain or class?

**A:** File an issue or discussion in the repository. The ontology can grow, but it should grow **intentionally**, not ad-hoc. Propose new domains/classes only when multiple playbooks would share the same classification.

### Q: Should I add ontology to external packs?

**A:** Optional. If you maintain a pack, adding ontology metadata helps teams understand coverage. Faultline will still work without it.

### Q: How do I know if my confidence baseline is right?

**A:** Test it:

1. Match the fixture against the playbook
2. Count true positives vs. false positives in production logs (if available)
3. Adjust based on false positive rate:
   - If FP rate < 5%, confidence 0.90+
   - If FP rate 5-15%, confidence 0.75-0.85
   - If FP rate > 15%, confidence < 0.70 (consider redesign)

### Q: Can I nest domains or classes?

**A:** No. Hierarchy is strictly these five levels: domain → class → mode → evidence → remediation. Nesting adds complexity without proportional benefit.

---

## Future Extensions

This ontology is designed to enable:

### 1. Enterprise Analytics
- Historical failure rates by domain, class, mode
- Remediation effectiveness tracking
- Team-specific failure patterns

### 2. Team-Level Customization
- Org-specific domains or classes
- Team-specific remediation strategies
- Custom confidence baselines for internal tools

### 3. Integrations
- Slack/Teams notifications with ontology context
- Jira automation rules by failure class
- Long-term trend dashboards

### 4. AI/LLM Assistance (Future)
- Structured input for debugging assistants
- Reproducible failure symptom encoding
- Deterministic remediation workflows independent of LLM

---

## References

- **SYSTEM.md** - Faultline architecture and system design
- **docs/playbooks.md** - Playbook authoring guide
- **docs/failures/** - Failure-specific documentation by domain
- **playbooks/bundled/** - Reference playbooks with ontology tags (post-Phase 2)

---

## Appendix: Ontology Expansion Reference

### When to Add a New Domain

- Multiple playbooks need a new operational subsystem
- Current domains don't fit the subsystem conceptually
- Example: `platform` was added when Kubernetes-specific failures appeared

### When to Add a New Class

- 3+ playbooks share the same root cause family
- The class is distinct from existing classes
- Example: `lockfile-drift` groups npm, pnpm, yarn lockfile issues

### When to Add a New Mode

- A specific error signature requires its own playbook
- The mode is not generic; it describes one root cause
- Example: `github-actions-env-not-persisted` is specific to GitHub Actions env handling

### When to Add a New Remediation Strategy

- The fix approach doesn't fit existing strategies
- Multiple playbooks would share this strategy
- Example: `wait-for-service-readiness` groups retries, health checks, timeouts

---

**Document Version:** 1.0  
**Last Updated:** 2026-04-25  
**Status:** Approved for Phase 2 (Tagging Existing Playbooks)
