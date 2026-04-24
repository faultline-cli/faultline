# Ingestion Pipeline Guardrails

These guardrails ensure the ingestion pipeline maintains quality, determinism, and appropriate boundaries.

## Core Guardrails

### Quality Standards
- **Do not promote raw staging files just because ingestion succeeded** - staging is for review, not automatic promotion
- **Do not skip sanitization** - all promoted fixtures must be properly sanitized
- **Do not add new product logic when the task is really corpus curation** - keep engineering work focused

### Determinism Requirements
- **Keep output deterministic and grounded in checked-in expectations** - all decisions must be reproducible
- **Maintain explicit expected playbooks for all promoted fixtures** - no implicit or assumed matches

### Source Selection Rules
- **Prefer public reports with direct failure evidence** over discussions with only speculation
- **Prefer one strong case each from several sources** over many similar snippets from one source
- **Treat additional snippets from the same URL as guilty until proven useful** - assume duplication unless distinct value is demonstrated
- **Promote repeated-source snippets only when they add a distinct failure boundary** - not just more wording around the same one

### Corpus Boundaries
- **Do not confuse "more snippets" with "more coverage"** - repeated-source snippets must earn their place
- **Do not force repository-inspection findings into `fixtures/real/`** when they are better represented as source-playbook regression fixtures
- **Handle repository-local risks in source-playbook fixtures** under `internal/engine/testdata/source/` with regression tests

## Acceptance Criteria

A successful ingestion run must meet these criteria:

- **Varied sources** - batch includes varied sources when available, not just repeated pulls from one thread
- **Bias correction** - run is biased toward underrepresented adapters or failure classes when current corpus stats show a gap
- **Sanitization** - staging output is properly sanitized
- **Explicit expectations** - promoted fixture has an explicit expected playbook
- **Baseline integrity** - real corpus still passes its deterministic baseline gate
- **Immediate resolution** - any new ambiguity is handled immediately, not deferred
- **Proper handoff** - any repository-inspection risk uncovered during intake is handed off to source-playbook fixture coverage rather than left implicit