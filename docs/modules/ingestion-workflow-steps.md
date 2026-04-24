# Ingestion Workflow Steps

This module defines the core steps for running Faultline's deterministic fixture intake pipeline from public URLs.

## Workflow Overview

The ingestion pipeline transforms public failure evidence into one of two outcomes:
- **Rejected from staging** - duplicates, near-duplicates, setup-only snippets, or cases needing sanitization
- **Promoted into `fixtures/real/`** - with explicit expected playbook and validation

## Step-by-Step Process

### 1. Source Validation
- Confirm each source is public and suitable for ingestion
- Ensure material has direct failure evidence, not just speculation

### 2. Corpus Analysis
- Check current corpus mix: `./bin/faultline fixtures stats --class real --json`
- Bias the run toward underrepresented adapters or failure classes when current corpus is skewed

### 3. Batch Construction
- Build a mixed candidate batch across multiple adapters when possible
- Prefer breadth over depth:
  - Take one or two strong candidates from a source before harvesting more from the same thread
  - Avoid spending the whole run on one issue, post, or discussion unless it clearly yields distinct failure signatures

### 4. Ingestion Execution
- Run `faultline fixtures ingest --adapter <adapter> --url <public-url>` for each candidate URL
- Use appropriate adapter from available source adapters

### 5. Staging Review
- Run `faultline fixtures review` to examine staged results
- Classify staged results:
  - Reject duplicates
  - Reject near-duplicates without meaningful new signal
  - Reject setup-only or workaround-only snippets
  - Reject anything that still needs sanitization
  - Keep only candidates with credible expected playbook and useful regression value

### 6. Promotion
- For accepted cases: `faultline fixtures promote <staging-id> --expected-playbook <id>`
- Add `--strict-top-1`, `--expected-stage`, or `--disallow` only when the boundary needs extra protection

### 7. Validation
- Run `make build`
- Run `./bin/faultline fixtures stats --class real --check-baseline`

### 8. Follow-up Actions
- If promoted fixture exposes weak matching or confusable neighbor: continue with playbook refinement
- If investigation surfaces repository-local risk belonging to bundled `source` playbook:
  - Add/update repository fixtures under `internal/engine/testdata/source/`
  - Extend source-playbook regression tests
  - Do not force this signal into real-log corpus