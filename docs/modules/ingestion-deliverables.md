# Ingestion Pipeline Deliverables

This module defines the expected deliverables from a complete ingestion pipeline run.

## Report Structure

Every ingestion pipeline run should produce a clear report including:

### 1. Commands Executed
- The exact ingestion and promotion commands run
- Any validation commands executed (build, stats, etc.)

### 2. Source Mix Analysis
- Which source adapters were used
- Distribution across adapters
- Any bias applied based on corpus statistics

### 3. Fixture Outcomes
- **Promoted fixture IDs** - successfully promoted to `fixtures/real/`
- **Rejected staging IDs** - rejected with reason (duplicate, near-duplicate, setup-only, etc.)
- **Total candidates processed**

### 4. Follow-on Work Required
- **Playbook refinement** - if promoted fixture exposes weak matching or confusable neighbor
- **Source-playbook fixture work** - if repository-local risk was uncovered
- **Regression test extensions** - for source-playbook findings

### 5. Validation Results
- Baseline check outcome (`./bin/faultline fixtures stats --class real --check-baseline`)
- Build status (`make build`)
- Any issues encountered during validation

## Command Skeleton Reference

```bash
# Ingestion commands (example)
faultline fixtures ingest --adapter github-issue --url <public-url>
faultline fixtures ingest --adapter stackexchange-question --url <public-url>
faultline fixtures ingest --adapter reddit-post --url <public-url>

# Review
faultline fixtures review

# Promotion
faultline fixtures promote <staging-id> --expected-playbook <id>

# Validation
make build
./bin/faultline fixtures stats --class real --check-baseline
```

## Quality Metrics

Track these metrics to ensure ingestion quality:
- **Source diversity** - number of distinct adapters used
- **Promotion rate** - percentage of ingested candidates promoted
- **Baseline integrity** - validation passes after promotion
- **Follow-up completeness** - all identified issues have clear next steps