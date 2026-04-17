# Agent Workflows

This repository should offer agents a small set of concrete, deterministic loops instead of generic software prompts.

The useful workflows in Faultline are grounded in the shipped CLI, fixture corpus, and playbook review gates:

- diagnose a known failure from a log with `faultline analyze` or `faultline workflow`
- turn a public failure report into a staged fixture with `faultline fixtures ingest`
- review staged evidence against the current catalog with `faultline fixtures review`
- promote accepted evidence into the checked-in corpus with `faultline fixtures promote`
- defend the catalog with `make review`, `make test`, and `faultline fixtures stats --class real --check-baseline`
- refine an existing playbook before adding a new one
- audit playbook coverage against an independent stratified sample of real CI failures
- author a new playbook only after a gap has been explicitly justified
- investigate and resolve a failing baseline gate without weakening it

## Current Direction

The repository now treats the prompt files under [`prompts/`](../prompts/) as task-specific operating procedures, not generic coding templates.

Removed:

- generic feature implementation prompts
- generic refactor prompts
- generic polish prompts not tied to Faultline's deterministic trust boundary

Added:

- [`prompts/run-ingestion-pipeline.md`](../prompts/run-ingestion-pipeline.md)
- [`prompts/triage-unmatched-log.md`](../prompts/triage-unmatched-log.md)
- [`prompts/curate-fixture-corpus.md`](../prompts/curate-fixture-corpus.md)
- [`prompts/refine-playbook-from-fixture.md`](../prompts/refine-playbook-from-fixture.md)
- [`prompts/verify-deterministic-change.md`](../prompts/verify-deterministic-change.md)
- [`prompts/collect-coverage-evidence.md`](../prompts/collect-coverage-evidence.md)
- [`prompts/author-new-playbook.md`](../prompts/author-new-playbook.md)
- [`prompts/investigate-baseline-regression.md`](../prompts/investigate-baseline-regression.md)

These are narrower, but they are more useful: they tell an agent exactly how to work inside this repository without inventing its own process.

## Repo Skills

The repository now ships five repo-local skills under [`agents/skills/`](../agents/skills/):

- [`ingestion-pipeline`](../agents/skills/ingestion-pipeline/SKILL.md) for public-source fixture intake, staging review, promotion, and baseline validation
- [`playbook-refinement`](../agents/skills/playbook-refinement/SKILL.md) for fixture-driven playbook tightening and workflow-field improvement
- [`coverage-evidence`](../agents/skills/coverage-evidence/SKILL.md) for auditing playbook coverage against a broad, stratified sample of real CI failures from independent sources
- [`new-playbook-authoring`](../agents/skills/new-playbook-authoring/SKILL.md) for authoring a new playbook after a gap has been explicitly justified, including pattern discipline, fixture pairing, and full validation
- [`baseline-regression`](../agents/skills/baseline-regression/SKILL.md) for isolating and resolving a failing `make fixture-check` gate without weakening the baseline

This location is intentionally agent-neutral so the same skill files can be used from Codex, Copilot, or any other repository-aware assistant workflow.

These skills should stay small and procedural. They are useful because they encode repository-specific judgment that a generic coding assistant would otherwise have to infer each time.

## Recommended Loops

### 1. Diagnose And Hand Off

Use this when there is already a failing log.

```bash
faultline analyze ci.log
faultline workflow ci.log --json --mode agent
```

Why this matters:
- `analyze` establishes the deterministic diagnosis and evidence
- `workflow` turns the same result into a bounded next-step artifact for an engineer or another agent

### 2. Ingest And Curate Evidence

Use this when we find a real public failure worth learning from.

```bash
faultline fixtures ingest --adapter github-issue --url <public-url>
faultline fixtures review
faultline fixtures promote <staging-id> --expected-playbook <id>
./bin/faultline fixtures stats --class real --check-baseline
```

Why this matters:
- it keeps acquisition, acceptance, and regression proof in one deterministic loop
- it prevents random fixture accumulation
- it should pull from a mixed set of public sources instead of overfitting the corpus to one provider or one long thread

### 3. Audit Coverage Before Authoring

Use this when the goal is to find coverage gaps rather than process a specific known fixture.

```bash
./bin/faultline list
faultline fixtures ingest --adapter <adapter> --url <url>
./bin/faultline analyze <staged-file> --json
faultline fixtures review
make build
./bin/faultline fixtures stats --class real --check-baseline
```

Why this matters:
- it separates discovery from promotion so gaps are validated before anything is committed
- it forces stratified sampling across CI systems, ecosystems, and failure categories
- weakly matched cases revealed here feed directly into the refinement workflow rather than being ignored

### 4. Refine Before Expanding

Use this when a fixture is weakly matched or confused with a neighbor.

```bash
faultline explain <candidate-playbook>
make review
make test
```

Why this matters:
- most catalog quality improvements should come from tightening an existing rule
- new playbooks should be the exception, not the default

### 5. Author Only After Justification

Use this after `triage-unmatched-log` or `collect-coverage-evidence` has confirmed a gap warrants a new playbook.

```bash
faultline explain <nearest-neighbor-id>    # pre-flight: confirm refinement can't cover it
make review                               # after YAML is authored
make test
make build
make fixture-check
```

Why this matters:
- the path from "justified" to "correct YAML in the catalog" is the highest-risk step
- pattern authoring without fixture pairing is how false positives are introduced
- `make review` is the only way to see whether a new pattern silently outcompetes an existing one

### 6. Investigate A Failing Gate

Use this when `make fixture-check` exits non-zero after a repository change.

```bash
make build
./bin/faultline fixtures stats --class real --json   # see per-fixture detail
./bin/faultline analyze fixtures/real/<id>.yaml --json  # isolate each regression
make review                                          # if a playbook was changed
make test
make fixture-check
```

Why this matters:
- the gate protects the checked-in trust boundary; papering it over with expectation updates is not a fix
- isolating the regressed fixture before acting prevents cascading changes
- every regression has exactly one correct outcome: fix forward, fix expectations, or revert

### 7. Close The Loop

Use this before considering a repository change complete.

```bash
make test
make review
```

Add this when corpus behavior changed:

```bash
make build
./bin/faultline fixtures stats --class real --check-baseline
```

Why this matters:
- Faultline's trust boundary is checked-in evidence, not optimistic reasoning

### 8. Prepare A Release Candidate

Use this when consolidating the repository for a tagged cut.

```bash
make release-check VERSION=v0.2.0
```

Add Docker validation when the release image also changed:

```bash
WITH_DOCKER=1 IMAGE=faultline-v0.2.0 make release-check VERSION=v0.2.0
```

Why this matters:
- it exercises tests, playbook review, fixture regression, archive packaging, and release smoke in one deterministic path
- it gives agent workflows a single release-grade validation target instead of a hand-built checklist

## Upgrade Path

The next useful upgrades should stay small and repo-native.

1. Make the new prompt set the default workflow surface for agents working in the repo.
2. Keep fixture ingestion and review as the only path for adding real-world evidence.
3. Bias ingestion runs toward source diversity across GitHub, GitLab, Stack Exchange, Discourse, and Reddit when useful evidence is available.
4. Keep playbook growth biased toward refinement over catalog expansion.
5. Strengthen workflow authoring inside playbooks by improving `likely_files`, `local_repro`, and `verify` for weak handoff cases. A structured sweep of the 74 bundled playbooks using `faultline explain <id>` to spot placeholder or thin workflow fields is a high-value, low-risk improvement pass.
6. Add new agent workflows only when they map cleanly to an existing deterministic command or checked-in regression gate.

## What Not To Add

Avoid adding workflows that are only general software advice, such as:

- broad "implement a feature" prompts
- repository-agnostic refactor guidance
- generic polish or cleanup checklists with no Faultline-specific validation path

If a workflow does not clearly improve fixture curation, playbook quality, deterministic validation, or workflow handoff quality, it probably does not belong here.
