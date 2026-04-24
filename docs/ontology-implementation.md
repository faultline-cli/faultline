# CI Failure Ontology: Implementation Guide

This document provides step-by-step guidance for implementing the CI Failure Ontology across Faultline's architecture, tooling, and workflows.

---

## Overview

The implementation spans six phases:

1. **Definition & Design** (DONE) - Ontology schema and hierarchy
2. **Playbook Tagging** - Add ontology metadata to existing 60+ playbooks
3. **Fixture Organization** - Tag and classify existing fixtures
4. **Coverage Reporting** - Build CLI commands and analytics
5. **Gap Analysis** - Identify and prioritize coverage gaps
6. **Continuous Improvement** - Maintain and evolve

This guide focuses on **Phase 2 (Playbook Tagging)** which is the critical foundation.

---

## Phase 2: Tagging Existing Playbooks

### Goal

Add ontology metadata to all 60+ bundled playbooks without breaking existing functionality.

**Constraints:**
- No changes to matching behavior
- No breaking changes to YAML schema
- Backward compatibility with old playbooks
- Tests must pass unchanged

### Process

#### Step 1: Prepare Domain Mapping

Create a mapping from existing `category` to new `domain`:

```
category → domain mapping:

auth → auth
build → dependency (lockfiles, conflicts, missing packages)
build → runtime (permission denied, executables)
build → filesystem (wrong paths)
build → container (docker/buildkit failures)
build → source (compilation, type errors)
build → ci-config (validation, step setup)

ci → ci-config
deploy → platform
deploy → container
deploy → auth
deploy → database

network → network
runtime → runtime
test → test-runner
test → database
```

**Tool:** Create a `docs/ontology-category-mapping.yaml`:

```yaml
schema_version: category-to-domain.v1

mappings:
  # Playbooks in playbooks/bundled/log/auth/*
  auth:
    domain: auth
    classes:
      - registry-auth
      - credential-mismatch
      - token-expiry
  
  # Playbooks in playbooks/bundled/log/build/*
  build:
    - id: npm-ci-lockfile
      domain: dependency
      class: lockfile-drift
      mode: npm-ci-requires-package-lock
    
    - id: missing-executable
      domain: runtime
      class: missing-executable
      mode: generic-executable-missing
    
    # ... one entry per playbook
```

#### Step 2: Batch Tagging

Process playbooks in batches by category to maintain consistency:

**Batch 1: Auth playbooks (7 total)**

```bash
# playbooks/bundled/log/auth/docker-auth.yaml
domain: auth
class: registry-auth
mode: docker-registry-403
aliases: []
severity: high
confidence_baseline: 0.95
```

Template for all auth playbooks:

```yaml
# === NEW: Ontology Classification ===
domain: auth
class: <class>
mode: <specific-mode>
aliases: []
severity: <critical|high|medium|low>
confidence_baseline: <0.90+>

# === EXISTING: Everything else unchanged ===
# (rest of playbook)
```

**Batch 2: Dependency-class build playbooks (lockfiles, conflicts, missing packages)**

Identify lockfile playbooks:
- `npm-ci-lockfile.yaml`
- `pnpm-lockfile-missing.yaml`
- `pnpm-lockfile.yaml`
- `yarn-lockfile.yaml`
- `poetry-lockfile-drift.yaml`

Tag all with:
```yaml
domain: dependency
class: lockfile-drift
mode: <tool>-lockfile-<specific-error>
```

Identify version-conflict playbooks:
- `npm-peer-dependency-conflict.yaml`
- `npm-eresolve-conflict.yaml`
- `package-manager-mismatch.yaml`

Tag all with:
```yaml
domain: dependency
class: version-conflict
mode: <tool>-<conflict-type>
```

**Batch 3: Runtime-class build playbooks (missing executables, permissions, resource limits)**

Identify missing-executable playbooks:
- `missing-executable.yaml`
- `jest-command-not-found.yaml`
- `python-command-not-found.yaml`

Tag all with:
```yaml
domain: runtime
class: missing-executable
mode: <tool>-missing-from-path
```

Identify permission playbooks:
- `npm-eacces-permission-denied.yaml`
- `docker-permission-denied-nonroot.yaml`

Tag all with:
```yaml
domain: runtime
class: permission-denied
mode: <tool>-insufficient-privilege
```

Identify resource playbooks:
- `node-out-of-memory.yaml`
- `oom-killed.yaml`
- `disk-full.yaml`

Tag all with:
```yaml
domain: runtime
class: resource-exhaustion
mode: <resource>-limit-exceeded
```

**Continue for remaining batches** (ci, deploy, network, runtime, test)

#### Step 3: Add Evidence Pattern Metadata (Optional, Phase 2b)

After basic tagging, enhance with optional evidence metadata:

```yaml
evidence:
  required:
    - log.contains: "exact error message"
  optional:
    - log.contains: "supporting evidence"
  exclusions:
    - log.contains: "false positive pattern"
  confidence: 0.92
  false_positive_risks:
    - "documented edge case"
```

This is optional in Phase 2. Focus on required fields first.

#### Step 4: Testing

After each batch, run tests:

```bash
# Verify no behavioral changes
make test

# Check for new conflicts
make review

# Validate YAML syntax
make lint
```

All tests must pass with zero behavior changes.

### Phase 2 Timeline

```
Week 1: Preparation
  - Create category mapping document
  - Review existing playbooks for consistency
  - Set up CI gate for ontology validation

Week 2: Batches 1-2 (Auth + Dependency)
  - Tag all auth playbooks (7)
  - Tag all lockfile playbooks (5)
  - Tag all version-conflict playbooks (4)
  - Tests + review

Week 3: Batches 3-4 (Runtime + CI-Config)
  - Tag all runtime playbooks (15)
  - Tag all ci-config playbooks (8)
  - Tests + review

Week 4: Batches 5-6 (Network + Deploy) + Finalization
  - Tag all network playbooks (8)
  - Tag all deploy playbooks (12)
  - Tag all test playbooks (16)
  - Final testing and documentation

Total: ~75 playbooks tagged, ~60 lines per playbook, ~4500 YAML changes
```

---

## Phase 3: Fixture Organization

### Goal

Classify existing fixtures by playbook ID, confidence, and depth.

### Process

#### Step 1: Inventory Fixtures

List all fixtures:

```bash
find fixtures/ -name "*.log" -o -name "*.json" | sort
```

Expected: ~80 fixtures across staging and real directories.

#### Step 2: Classify Each Fixture

For each fixture, determine:
- **Playbook ID**: Which playbook should match?
- **Confidence**: Expected match confidence (0.80-1.00)
- **Type**: Positive (should match) or Negative (should NOT match)
- **Depth**: First (shallow), secondary (medium), edge case (deep)

Create `fixtures/manifest.yaml`:

```yaml
schema_version: fixtures.v1

fixtures:
  - id: npm-ci-lockfile-simple
    playbook_id: npm-ci-lockfile
    type: positive
    confidence: 0.95
    depth: primary
    path: fixtures/real/npm-ci-lockfile-simple.log
    description: npm ci fails, lockfile out of sync
  
  - id: npm-ci-lockfile-workspace
    playbook_id: npm-ci-lockfile
    type: positive
    confidence: 0.90
    depth: secondary
    path: fixtures/real/npm-ci-lockfile-workspace.log
    description: workspace with npm 9.x vs 10.x mismatch
  
  - id: npm-enoent-not-lockfile
    playbook_id: npm-ci-lockfile
    type: negative
    confidence: null
    depth: regression
    path: fixtures/real/npm-enoent-file-missing.log
    description: ENOENT error, should NOT match npm-ci-lockfile
    confuses_with: npm-enoent-package-json
```

#### Step 3: Add Missing Negative Fixtures

Identify playbooks with no negative fixtures:

```bash
# For each playbook without a negative fixture, create one
# Example: Create a fixture that looks similar but should NOT match
```

Priority order:
1. High-confidence playbooks (0.90+) must have 1+ negative fixtures
2. Medium-confidence playbooks (0.70-0.89) should have 1+ negative
3. Low-confidence playbooks must have 2+ negatives

---

## Phase 4: Coverage Reporting

### Goal

Build `faultline coverage` command and automated reporting.

### Implementation

#### Step 1: Add Coverage CLI Command

**Location:** `internal/cli/coverage.go`

```go
package cli

import (
    "faultline/internal/playbooks"
    "faultline/internal/coverage"
)

// CoverageCmd reports ontology coverage metrics
type CoverageCmd struct {
    Domain   string `help:"Filter by domain"`
    Depth    string `help:"Filter by depth (deep/medium/shallow)"`
    Gaps     bool   `help:"Show only coverage gaps"`
    Format   string `default:"text" help:"Output format (text/json/csv)"`
}

func (c *CoverageCmd) Run() error {
    pbs, err := playbooks.LoadDir(...)
    if err != nil {
        return err
    }
    
    report := coverage.Analyze(pbs,
        coverage.WithDomain(c.Domain),
        coverage.WithDepth(c.Depth),
        coverage.WithGaps(c.Gaps),
    )
    
    return coverage.Render(report, c.Format)
}
```

#### Step 2: Implement Coverage Analysis

**Location:** `internal/coverage/analysis.go`

```go
package coverage

type Report struct {
    DomainCoverage   map[string]*DomainReport
    ClassCoverage    map[string]*ClassReport
    ModeCoverage     map[string]*ModeReport
    ConfidenceDistribution
    Gaps             []Gap
    Metrics          CoverageMetrics
}

type DomainReport struct {
    Domain     string
    Count      int
    Depth      string  // deep/medium/shallow
    Classes    []string
    Playbooks  int
}

type Gap struct {
    Domain   string
    Class    string
    Priority string  // critical/high/medium/low
    Reason   string
}

func Analyze(pbs []playbooks.Playbook, opts ...Option) *Report {
    report := &Report{
        DomainCoverage: make(map[string]*DomainReport),
        // ...
    }
    
    for _, pb := range pbs {
        if pb.Domain == "" {
            continue // Skip playbooks without ontology
        }
        
        // Aggregate by domain, class, mode
        // Calculate confidence distribution
        // Identify gaps
    }
    
    return report
}
```

#### Step 3: Implement Rendering

**Rendering support:**

1. **Text format:**
   ```
   Domain Coverage Summary
   ========================
   
   dependency: 22 playbooks (deep)
     - lockfile-drift: 5 playbooks ✓
     - version-conflict: 4 playbooks ✓
     - missing-package: 3 playbooks ○
     - cache-corruption: 1 playbook ✗
   
   runtime: 18 playbooks (deep)
     - missing-executable: 4 playbooks ✓
     - interpreter-mismatch: 3 playbooks ✓
     ...
   ```

2. **JSON format:**
   ```json
   {
     "domains": {
       "dependency": {
         "count": 22,
         "depth": "deep",
         "classes": {
           "lockfile-drift": {
             "count": 5,
             "modes": 5,
             "fixtures": 12,
             "confidence_avg": 0.93
           }
         }
       }
     },
     "confidence_distribution": {
       "0.95-1.00": 38,
       "0.85-0.94": 42
     }
   }
   ```

3. **CSV format:** For data pipeline integration

#### Step 4: Add to CLI Root

**Location:** `cmd/root.go`

```go
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use: "faultline",
    }
    
    cmd.AddCommand(&cli.CoverageCmd{})  // NEW
    
    return cmd
}
```

### Coverage Command Usage

```bash
# All
faultline coverage

# By domain
faultline coverage --domain=dependency
faultline coverage --domain=runtime

# Show only shallow coverage
faultline coverage --depth=shallow

# Show gaps
faultline coverage --gaps
faultline coverage --gaps --domain=database

# JSON output for dashboards
faultline coverage --format=json > coverage.json

# CSV for spreadsheet
faultline coverage --format=csv > coverage.csv
```

---

## Phase 5: Gap Analysis

### Goal

Use coverage reporting to identify priority gaps and plan playbook expansion.

### Process

#### Step 1: Generate Coverage Report

```bash
faultline coverage --format=json > coverage-snapshot.json
```

#### Step 2: Analyze Results

Use `prompts/coverage-gaps.md` to guide human analysis:

- What domains have 0 playbooks?
- What classes within domains are missing?
- Which playbooks have low confidence?
- Which areas lack negative fixtures?

#### Step 3: Prioritize Gaps

Create a prioritization matrix:

| Gap | Domain | Impact | Detectability | Differentiation | Priority |
|-----|--------|--------|----------------|-----------------|----------|
| Cache validation | dependency | high | medium | high | P0 |
| Rate limiting | network | medium | medium | low | P1 |
| Job cancellation | ci-config | low | low | low | P2 |

#### Step 4: Plan Playbook Expansion

For each P0/P1 gap, create a playbook design document:

```markdown
# New Playbook: npm-cache-validation-failure

**Gap:** dependency / cache-poisoning

**Root Cause:** npm cache becomes corrupted (mismatched hashes, 
partial downloads, concurrent access)

**Evidence:**
- log.contains: "npm ERR! code EINTEGRITY"
- log.contains: "cache verification failed"

**Remediation:** clear-corrupt-cache

**Why Now:** Cache issues are high-frequency in monorepos; 
no existing coverage.
```

---

## Phase 6: Continuous Improvement

### Goal

Maintain and evolve the ontology as new failure modes emerge.

### Process

#### Ontology Review Gates

Add to `make test` and `make review`:

```bash
# Validate all playbooks have ontology metadata
faultline playbooks validate --ontology

# Check confidence distribution hasn't degraded
faultline coverage --check confidence-floor=0.75

# Ensure all high-confidence playbooks have negative fixtures
faultline fixtures validate --require-negatives --confidence-floor=0.90
```

#### New Playbook Checklist

When authoring a new playbook:

```markdown
- [ ] Domain is clear (not symptom-based)
- [ ] Class is reused or justified as new
- [ ] Mode is specific to this root cause
- [ ] Evidence pattern has 2+ required signals
- [ ] Confidence baseline is honest (not inflated)
- [ ] Exclusions document known false positives
- [ ] Remediation strategy is clear
- [ ] 1+ positive fixture + 1+ negative fixture
- [ ] Related modes are documented
- [ ] Tests pass: make test
- [ ] Review passes: make review
- [ ] Coverage report updated
```

#### Quarterly Ontology Review

Every quarter:

1. Generate coverage snapshot
2. Compare to previous quarter
3. Identify new patterns in production logs
4. Update ROADMAP with ontology-driven priorities
5. Publish coverage report in docs

---

## Integration with Existing Workflows

### Local Development

Developer adds a new playbook:

```bash
# 1. Create playbook with full ontology
vim playbooks/bundled/log/build/my-new-failure.yaml

# 2. Run tests (ontology validation included)
make test

# 3. Check impact on coverage
faultline coverage --domain=<domain>

# 4. Commit with ontology fields
git add playbooks/bundled/log/build/my-new-failure.yaml
git commit -m "feat: add my-new-failure playbook with ontology"
```

### CI/CD Gates

Add to GitHub Actions (`.github/workflows/test.yml`):

```yaml
- name: Validate Ontology
  run: |
    faultline playbooks validate --ontology
    faultline coverage --check confidence-floor=0.75
    faultline fixtures validate --require-negatives --confidence-floor=0.90
```

### Documentation

Public-facing docs auto-generated from ontology:

```go
// Generate docs/failures/ONTOLOGY_COVERAGE.md
func GenerateCoverageDocs() {
    pb, _ := playbooks.LoadDir(...)
    coverage := coverage.Analyze(pb)
    md := coverage.MarkdownReport()
    os.WriteFile("docs/failures/ONTOLOGY_COVERAGE.md", md, 0644)
}
```

---

## Rollout Timeline (6-8 weeks total)

```
Week 1-2: Phase 2 (Playbook Tagging)
   - Batch 1-2 tagged and tested
   - Category mapping documented
   
Week 2-3: Phase 2 (continued)
   - Batch 3-4 tagged and tested
   - CI gate for ontology validation added
   
Week 3-4: Phase 2 completion + Phase 3 activation
   - Batch 5-6 tagged and tested
   - All playbooks decorated
   - Fixtures classified
   - Negative fixture gaps identified
   
Week 4-5: Phase 4 (Coverage Reporting)
   - faultline coverage command implemented
   - Coverage report generated and reviewed
   - Documentation updated
   
Week 5-6: Phase 5 (Gap Analysis)
   - Coverage analysis session with team
   - Priority gaps identified
   - Playbook expansion planned for next cycle
   
Week 6-8: Phase 6 (Continuous Improvement)
   - New CI gates activated
   - Developer guidelines updated
   - Quarterly review process established
   - Public-facing coverage docs published
```

---

## Success Metrics

The implementation is successful when:

1. **Coverage Metrics**
   - 95%+ of bundled playbooks have full ontology metadata
   - Coverage report is deterministic and reproducible
   - Confidence distribution is justified and documented

2. **Process Improvements**
   - New playbooks are authored 20% faster (ontology template reduces ambiguity)
   - False positive rate decreases by 15%+ (better negative fixtures)
   - Ranking stability improves (discriminator signals reduce confusion)

3. **Knowledge System**
   - Contributors understand domain/class/mode hierarchy
   - Coverage gaps are discovered and prioritized systematically
   - Remediation strategies are reused across playbooks
   - Enterprise reporting is ready for Phase 7

---

## Appendix: Schema Validation

Playbook YAML must validate against this schema:

```yaml
required_fields:
  - id
  - title
  - category
  - severity
  - match

ontology_fields:  # Phase 2
  - domain
  - class
  - mode
  - aliases
  - confidence_baseline

ontology_validation:
  domain: enum(dependency, runtime, container, auth, network, ci-config, test-runner, database, filesystem, platform, source)
  class: string (documented in ontology.md)
  mode: string (specific, actionable, unique)
  confidence_baseline: float(0.0-1.0)
  aliases: list[string]

evidence_fields:  # Phase 2b
  - required: list[signal]
  - optional: list[signal]
  - exclusions: list[signal]
  - confidence: float(0.0-1.0)
  - false_positive_risks: list[string]

signal_format:
  - log.contains:<text>
  - log.regex:<pattern>
  - log.absent:<text>
  - delta.signal:<id>
  - delta.absent:<id>
  - context.stage:<stage>
  - file.exists:<path>
  - named_signal:<alias>
```

---

**Document Version:** 1.0  
**Last Updated:** 2026-04-25  
**Status:** Ready for Phase 2 Implementation
