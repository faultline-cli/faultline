# CI Failure Ontology for Faultline

**Status:** ✅ Design Complete, Ready for Phase 2 Implementation

---

## What Is This?

Faultline's **CI Failure Ontology** is a canonical taxonomy that transforms the playbook collection into a structured CI failure knowledge system.

**Before:** Faultline is *"a collection of playbooks"*  
**After:** Faultline is *"a structured CI failure knowledge system"*

It enables:
- Deterministic playbook classification across domains, classes, and modes
- Coverage gap analysis and automated reporting
- Fixture organization with clear depth expectations
- Contributor guidance for consistent ontology adoption
- Enterprise reporting on CI failure patterns and remediation effectiveness

---

## Core Documents

### 1. **docs/ontology.md** — Complete Design
The authoritative reference for the ontology.

**Contains:**
- Design principles (determinism, backwards compatibility, extensibility)
- Five-level hierarchy (Domain → Class → Mode → Evidence → Remediation)
- 11 domains, 40+ classes, and mode naming conventions
- Complete YAML schema with examples
- Coverage reporting model (what can be measured)
- 6-phase migration plan
- Contributor guidance (step-by-step playbook authoring)
- Quality bar checklist
- FAQ and future extensions

**Read this if:**
- You need to understand the full architecture
- You're designing new domains or classes
- You want to contribute new playbooks
- You're implementing the system

**Lines:** ~1,200 | **Read time:** 30-45 minutes

---

### 2. **docs/ontology-examples.md** — Real-World Examples
Six complete, production-ready examples with every field populated.

**Examples:**
1. **NPM Lockfile Mismatch** (dependency → lockfile-drift)
   - Root cause, evidence patterns, fixtures, remediation
   - How to integrate into existing playbook
   
2. **Missing Local Binary** (runtime → missing-executable)
   - Missing Node, Python, Docker in CI
   - Multiple fixture scenarios
   
3. **Python Interpreter Mismatch** (runtime → interpreter-mismatch)
   - Virtual environment issues
   - Negative fixture examples
   
4. **Docker COPY Missing File** (filesystem → wrong-working-directory)
   - Dockerfile context issues
   - Confusable error patterns
   
5. **GitHub Actions Env Not Persisted** (ci-config → env-not-persisted)
   - Environment variable isolation
   - Deprecated vs modern syntax
   
6. **Postgres Service Not Ready** (database → service-not-ready)
   - Connection timing issues
   - Startup lag scenarios

**Read this if:**
- You want to author a new playbook
- You're tagging existing playbooks in Phase 2
- You need to understand the full record structure
- You want templates to copy/paste

**Lines:** ~800 | **Read time:** 30-40 minutes

---

### 3. **docs/ontology-implementation.md** — Implementation Roadmap
Step-by-step guide to rolling out the ontology across Faultline.

**Phases:**
- **Phase 1** (DONE): Define and design ✅
- **Phase 2**: Tag existing 60+ playbooks (2-4 weeks)
- **Phase 3**: Organize fixtures by classification (1 week)
- **Phase 4**: Build coverage reporting CLI (1 week)
- **Phase 5**: Gap analysis and planning (1-2 weeks)
- **Phase 6**: Continuous improvement (ongoing)

**Contains:**
- Detailed Phase 2 batch tagging process
- Category-to-domain mapping strategy
- Fixture classification process
- Coverage reporting implementation (Go code skeleton)
- CI/CD gate configuration
- Developer workflow integration
- Success metrics and validation

**Read this if:**
- You're implementing the ontology
- You're a maintainer planning the rollout
- You need to set up CI validation gates
- You're building the coverage reporting system

**Lines:** ~500 | **Read time:** 25-35 minutes

---

### 4. **docs/ontology-quick-reference.md** — Cheat Sheet
Condensed reference for quick lookup during playbook authoring.

**Quick lookups:**
- All 11 domains (one-page table)
- Class examples by domain
- Minimal playbook schema
- Confidence baseline guidelines
- Evidence signal types
- Mode naming conventions
- 8 core remediation strategies
- New playbook flow (6 steps)
- Coverage command examples
- Editor checklist
- Migration FAQ

**Read this if:**
- You're authoring a quick playbook
- You're reviewing someone else's playbook
- You need to look up domain definitions
- You're running coverage reports

**Lines:** ~300 | **Read time:** 10-15 minutes

---

## How to Use These Documents

### 👤 I'm a Contributor
1. Read **Quick Reference** → understand domains/classes
2. Read **ontology.md** → "Contributor Guidance" section
3. Review **examples.md** → find a similar failure mode
4. Write your playbook with full ontology metadata
5. Run `make test && make review`

### 👨‍💼 I'm a Maintainer
1. Read **ontology.md** → full design
2. Read **implementation.md** → phase-by-phase roadmap
3. Start with **Phase 2**: Pick a batch, tag playbooks
4. After each batch: `make test`, `make review`
5. Build coverage reporting in **Phase 4**

### 👨‍💻 I'm Implementing the System
1. Read **ontology.md** → schema and principles
2. Read **implementation.md** → detailed implementation roadmap
3. Phase 2: Batch playbook tagging
4. Phase 3: Fixture organization
5. Phase 4: Implement `faultline coverage` (Go skeleton provided)
6. Phase 5-6: Build reporting and analytics

### 🏢 I'm an Enterprise User
1. Read **Quick Reference** → domains, remediation strategies
2. Read **ontology.md** → "Coverage Reporting Model" section
3. Run: `faultline coverage --format=json`
4. Analyze domains, classes, confidence distribution
5. File issues for missing coverage in your stack

---

## Key Concepts at a Glance

### Five-Level Hierarchy

```
Domain (operational subsystem)
  ↓ Are you having dependency/runtime/auth/network issues?
  ↓
Class (family within domain)
  ↓ Is it lockfile-drift or version-conflict?
  ↓
Mode (specific root cause)
  ↓ Is it npm-ci-lockfile or pnpm-lockfile?
  ↓
Evidence Pattern (deterministic signals)
  ↓ What exact log messages prove this failure?
  ↓
Remediation Strategy (fix approach)
  ↓ align-lockfile? install-missing-tool? wait-for-service?
```

### 11 Domains

| Domain | Subsystem | Count |
|--------|-----------|-------|
| `dependency` | Package resolution, lockfiles | ~22 playbooks |
| `runtime` | Executables, permissions, OOM | ~18 playbooks |
| `container` | Docker build/pull | ~8 playbooks |
| `auth` | Credentials, tokens | ~12 playbooks |
| `network` | DNS, connectivity, TLS | ~9 playbooks |
| `ci-config` | Workflow validation, env vars | ~7 playbooks |
| `test-runner` | Test framework, coverage | ~15 playbooks |
| `database` | Connection, migration | ~5 playbooks |
| `filesystem` | Paths, permissions, disk | ~6 playbooks |
| `platform` | K8s, Terraform, etc. | ~3 playbooks |
| `source` | Code quality, compilation | ~3 playbooks |

### Minimal Schema

```yaml
# NEW: Ontology (4 required fields)
domain: dependency       # Which subsystem?
class: lockfile-drift    # Which family?
mode: npm-ci-requires-package-lock  # Which specific root cause?
confidence_baseline: 0.95            # How certain?

# EXISTING: Unchanged
match:
  any: [...]
summary: ...
diagnosis: ...
fix: ...
validation: ...
```

### Coverage Insights

Ask the system:
- ✅ What domains do we cover deeply?
- ✗ Which domains have no playbooks?
- 🟡 Which playbooks have low confidence?
- 🔍 Where are the coverage gaps?
- 📊 What's the confidence distribution?

---

## Rollout Timeline

```
Phase 1 (DONE)
  ✅ Ontology design & documentation

Phase 2 (2-4 weeks)
  Batch 1: Auth playbooks (7)
  Batch 2: Dependency playbooks (10)
  Batch 3: Runtime playbooks (15)
  Batch 4: CI-config playbooks (8)
  Batch 5: Network, Deploy, Test playbooks (36)

Phase 3 (1 week)
  Fixture classification & metadata

Phase 4 (1 week)
  Build faultline coverage command
  Rendering (text, JSON, CSV)

Phase 5 (1-2 weeks)
  Coverage analysis session
  Gap prioritization & planning

Phase 6 (ongoing)
  CI gates, new playbook reviews
  Quarterly coverage reports

Total: 6-8 weeks from start to production
```

---

## Success Criteria

✅ **Technical:**
- 95%+ of bundled playbooks have ontology metadata
- Coverage reports are deterministic and reproducible
- Confidence distribution is justified
- All tests pass with zero behavior changes

✅ **Process:**
- New playbooks authored 20% faster
- False positive rate decreases 15%+
- Ranking stability improves

✅ **Knowledge:**
- Contributors understand domain/class/mode
- Coverage gaps are discovered systematically
- Remediation strategies are reused
- Enterprise reporting is ready

---

## Next Steps

### Immediate (This Sprint)
1. ✅ **Read & Review** — Review the four documents
2. ✅ **Feedback** — Discuss domain definitions, classes
3. ⏳ **Alignment** — Confirm hierarchy matches Faultline architecture

### Short-term (Next Sprint)
1. ⏳ **Phase 2 Start** — Begin batch 1 playbook tagging
2. ⏳ **Category Mapping** — Create domain mapping document
3. ⏳ **First Batch** — Tag 7 auth playbooks, test, commit

### Medium-term (Following Sprints)
1. ⏳ **Phase 2 Completion** — All playbooks tagged
2. ⏳ **Phase 3** — Fixture organization
3. ⏳ **Phase 4** — Coverage reporting CLI
4. ⏳ **Phase 5** — Gap analysis & roadmap update

---

## Questions & Feedback

The ontology is designed to evolve. If you have questions:

- **On design philosophy:** See `docs/ontology.md` → "Design Principles"
- **On specific examples:** See `docs/ontology-examples.md`
- **On implementation:** See `docs/ontology-implementation.md`
- **On naming:** See `docs/ontology-quick-reference.md`

File issues or discussions in the repository. All feedback is welcome.

---

## Document Map

```
docs/
├── ontology.md                  # 📖 Complete design (authoritative)
├── ontology-examples.md         # 📋 6 real-world examples
├── ontology-implementation.md   # 🛠️ Implementation roadmap
├── ontology-quick-reference.md  # ⚡ Quick lookup cheat sheet
├── ONTOLOGY_README.md           # 👈 You are here
└── [Auto-generated]
    ├── ONTOLOGY_COVERAGE.md     # Coverage report (Phase 4)
    └── ontology-category-mapping.yaml  # Domain mapping (Phase 2)
```

---

## Related Documents

- **docs/SYSTEM.md** — Faultline architecture and system design
- **docs/playbooks.md** — Playbook authoring guide (still valid)
- **docs/adr/** — Architecture Decision Records
- **AGENTS.md** — Agent operating rules for deterministic development
- **prompts/coverage-gaps.md** — Guidance for coverage analysis

---

## Version & Status

| Aspect | Status |
|--------|--------|
| Design | ✅ Complete |
| Documentation | ✅ Complete (4 docs, 2,800+ lines) |
| Examples | ✅ Complete (6 detailed examples) |
| Implementation Guide | ✅ Complete |
| Implementation | ⏳ Ready to start Phase 2 |
| Coverage Reporting | ⏳ Ready for Phase 4 |

**Current:** Design Phase (complete)  
**Next:** Phase 2 — Playbook Tagging  
**Timeline:** 6-8 weeks from Phase 2 to Phase 6 completion

---

## Document Statistics

| Document | Lines | Read Time | Audience |
|----------|-------|-----------|----------|
| ontology.md | 1,200 | 30-45 min | Architects, Maintainers, Contributors |
| ontology-examples.md | 800 | 30-40 min | Contributors, Implementers |
| ontology-implementation.md | 500 | 25-35 min | Maintainers, Implementers |
| ontology-quick-reference.md | 300 | 10-15 min | Contributors, Reviewers |
| ONTOLOGY_README.md (this) | 400 | 15-20 min | Everyone |
| **Total** | **3,200+** | **110–160 min** | — |

---

## Final Note

This ontology is not paperwork. It's the backbone of:
- **Better playbooks** (clearer intent, tighter signals)
- **Better fixtures** (positive AND negative test coverage)
- **Better docs** (aligned remediation guidance)
- **Better product reporting** (coverage metrics, gap analysis)
- **Better agentic workflows** (structured input, deterministic output)

It evolves with Faultline. Start with Phase 2, learn from the first batch, refine the hierarchy, and grow the coverage systematically.

---

**Created:** 2026-04-25  
**Status:** Ready for adoption and implementation  
**Version:** 1.0

For issues, feedback, or questions: File an issue in the repository.
