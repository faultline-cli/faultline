---
name: coverage-evidence
description: Use this skill when the task is to audit playbook coverage by collecting a broad, stratified sample of real CI failure evidence from public sources. Trigger it for requests about finding coverage gaps, sampling CI failures across ecosystems, measuring catalog robustness against external evidence, identifying unmatched or weakly matched failure classes, or building an evidence batch for playbook discovery. Do not trigger it for promoting a specific known fixture (use ingestion-pipeline) or for refining an individual playbook (use playbook-refinement).
---

# Coverage Evidence Collection

This skill is for auditing Faultline's playbook catalog against a broad, independent sample of real CI failures sourced from public sources.

Use it when the task involves:

- discovering coverage gaps across failure categories, CI systems, or language ecosystems
- sampling public CI failure logs to test playbook robustness without starting from a pre-selected fixture
- identifying unmatched or weakly matched failure classes across a diverse evidence set
- measuring catalog coverage before or after a playbook authoring sprint
- producing a per-sample evidence record for follow-on ingestion or refinement work

Do not use it for:

- promoting a specific pre-selected fixture (use `ingestion-pipeline`)
- refining a single playbook based on a known regression (use `playbook-refinement`)
- generic coding tasks unrelated to corpus coverage or playbook quality

## Read First

- [`SYSTEM.md`](../../../SYSTEM.md)
- [`docs/fixture-corpus.md`](../../../docs/fixture-corpus.md)
- [`docs/playbooks.md`](../../../docs/playbooks.md)
- [`docs/agent-workflows.md`](../../../docs/agent-workflows.md)
- [`prompts/collect-coverage-evidence.md`](../../../prompts/collect-coverage-evidence.md)
- [`fixtures/staging/README.md`](../../../fixtures/staging/README.md)

## Workflow

1. Run `./bin/faultline list` to orient against the current catalog. Note thin or absent failure categories.
2. Check `fixtures/real/baseline.json` to understand which failure classes already have accepted regression coverage.
3. Select evidence from at least three distinct source types per run. Prefer sources with direct machine-produced log output.
4. Collect a stratified sample of 8–12 failure cases, targeting:
   - ≥2 distinct CI systems
   - ≥3 distinct language ecosystems
   - ≥4 distinct failure categories
   - ≥3 distinct source adapters
   - no more than 2 samples from the same repository, thread, or discussion
5. For each sample, run `faultline fixtures ingest --adapter ... --url ...` to stage it, then `./bin/faultline analyze <file> --json` to capture match results.
6. Classify each sample using the five-tier scheme:
   - **Covered** — top-1 score ≥ 0.7, correct playbook, no close confusable neighbor
   - **Weakly covered** — top-1 matches but score < 0.7, or wrong playbook ranks above the correct one
   - **Gap — known category** — unmatched but maps to an existing failure class with a plausible authoring path
   - **Gap — new category** — unmatched and the root cause has no existing playbook analog
   - **Noise** — insufficient log content, workaround-only, ambiguous environment, or duplicate
7. Run `faultline fixtures review` to check for staging duplicates.
8. Escalate gap candidates to the appropriate follow-on workflow:
   - Gap — known category → `playbook-refinement` skill
   - Gap — new category → `triage-unmatched-log.md` prompt, then `playbook-refinement` skill if confirmed
9. Run `make build && ./bin/faultline fixtures stats --class real --check-baseline` to confirm the real corpus is stable.

## Evidence Metadata To Capture Per Sample

Record these fields for every sample:

```
source_url, adapter, ci_system, os, arch, language, language_version,
toolchain, toolchain_version, trigger, job_or_step, failure_class,
top_match_playbook, top_match_score, classification, notes
```

## Guardrails

- Do not promote fixtures directly from this workflow; hand off to the ingestion pipeline.
- Do not author new playbooks during this workflow; hand off to triage or refinement.
- Do not collect more from a single source just because it is producing results; diversity is the goal.
- Do not skip the baseline gate after staging activity.
- Treat a weakly covered case as a gap, not a success.

## Deliverable

Report:

- a per-sample evidence table (source, environment metadata, top-match result, classification)
- a coverage summary (covered / weakly covered / gap / noise counts)
- the subset of gap candidates escalated and which follow-on workflow each was handed to
- the baseline check result
