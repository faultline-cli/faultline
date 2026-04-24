# Agent Workflows

This repository should offer agents a small set of concrete, deterministic loops instead of generic software prompts.

The useful workflows in Faultline are grounded in the shipped CLI, fixture corpus, and playbook review gates:

- diagnose a known failure from a log with `faultline analyze` or `faultline workflow`
- turn a public failure report into a staged fixture with `faultline fixtures ingest`
- draft a maintainer-only candidate playbook from sanitized evidence with `faultline fixtures scaffold`
- review staged evidence against the current catalog with `faultline fixtures review`
- promote accepted evidence into the checked-in corpus with `faultline fixtures promote`
- defend the catalog with `make review`, `make test`, and `faultline fixtures stats --class real --check-baseline`
- refine an existing playbook before adding a new one
- audit playbook coverage against an independent stratified sample of real CI failures
- author a new playbook only after a gap has been explicitly justified
- investigate and resolve a failing baseline gate without weakening it

The current shipping boundary is documented in [`docs/release-boundary.md`](./release-boundary.md). Agents should treat that boundary as the default product scope and keep experimental or maintainer-only surfaces out of the main user narrative unless the task explicitly targets them.

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
- [`prompts/refine-source-playbook-from-repo.md`](../prompts/refine-source-playbook-from-repo.md)
- [`prompts/verify-deterministic-change.md`](../prompts/verify-deterministic-change.md)
- [`prompts/collect-coverage-evidence.md`](../prompts/collect-coverage-evidence.md)
- [`prompts/author-new-playbook.md`](../prompts/author-new-playbook.md)
- [`prompts/investigate-baseline-regression.md`](../prompts/investigate-baseline-regression.md)

These are narrower, but they are more useful: they tell an agent exactly how to work inside this repository without inventing its own process.

## Repo Skills

The repository now ships eight repo-local skills under [`agents/skills/`](../agents/skills/):

- [`ingestion-pipeline`](../agents/skills/ingestion-pipeline/SKILL.md) for public-source fixture intake, staging review, promotion, and baseline validation
- [`playbook-refinement`](../agents/skills/playbook-refinement/SKILL.md) for fixture-driven playbook tightening and workflow-field improvement
- [`coverage-evidence`](../agents/skills/coverage-evidence/SKILL.md) for auditing playbook coverage against a broad, stratified sample of real CI failures from independent sources
- [`new-playbook-authoring`](../agents/skills/new-playbook-authoring/SKILL.md) for authoring a new playbook after a gap has been explicitly justified, including pattern discipline, fixture pairing, ontology classification, and full validation
- [`baseline-regression`](../agents/skills/baseline-regression/SKILL.md) for isolating and resolving a failing `make fixture-check` gate without weakening the baseline
- [`source-playbook-refinement`](../agents/skills/source-playbook-refinement/SKILL.md) for repository-local source-detector refinement and bundled source-playbook authoring
- [`fixture-generation`](../agents/skills/fixture-generation/SKILL.md) for producing canonical, noisy, and near-miss fixture trios that defend a specific playbook pattern
- [`playbook-linter`](../agents/skills/playbook-linter/SKILL.md) for applying the deterministic quality gate before `make review` — checks matcher precision, false positive risk, fixture realism, negative coverage, and ontology completeness

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
- that workflow handoff now includes the first-class failure artifact plus
  additive remediation commands, patch suggestions, and CI config diff hints
- workflow artifacts may carry additive `ranking_hints`, `delta_hints`,
  `metrics_hints`, and `policy_hints` when the underlying analysis has enough
  explicit context

### 2. Ingest And Curate Evidence

Use this when we find a real public failure worth learning from.

```bash
faultline fixtures ingest --adapter github-issue --url <public-url>
faultline fixtures sanitize <staging-id>
faultline fixtures scaffold --from-fixture <staging-id> --category <category>
faultline fixtures review
faultline fixtures promote <staging-id> --expected-playbook <id>
./bin/faultline fixtures stats --class real --check-baseline
```

Why this matters:
- it keeps acquisition, acceptance, and regression proof in one deterministic loop
- it keeps drafting tied to sanitized staging evidence instead of free-form rule writing
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

### 5. Refine Repository-Local Source Findings

Use this when `inspect` or `guard` surfaces a bundled source-detector finding in a repository tree.

```bash
faultline inspect .
faultline guard .
faultline explain <candidate-source-playbook>
make review
make test
make build
make cli-smoke
```

Why this matters:
- source findings belong in the bundled `source` catalog, not the real log corpus
- the source workflow should stay fixture-backed and deterministic
- inspect and guard need the same repeatable refinement loop that ingestion uses for public log evidence

### 6. Author Only After Justification

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

### 7. Investigate A Failing Gate

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

### 8. Close The Loop

Use this before considering a repository change complete.

```bash
make test
make review
make cli-smoke
```

Add this when corpus behavior changed:

```bash
make build
./bin/faultline fixtures stats --class real --check-baseline
```

Why this matters:
- Faultline's trust boundary is checked-in evidence, not optimistic reasoning

### 9. Prepare A Release Candidate

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

## Composed Skill Chains

The skills above are designed to compose. These are the primary chains:

### Author A New Playbook (full chain)

```
coverage-evidence or triage-unmatched-log
  → gap confirmed
  → fixture-generation skill    (canonical + noisy + near-miss)
  → new-playbook-authoring skill (YAML + ontology fields)
  → playbook-linter skill        (PASS required before make review)
  → make review → make test → make build → make fixture-check
```

Trigger: `coverage-evidence` or `triage-unmatched-log` confirms a new playbook is warranted.  
Gate: `playbook-linter` PASS is required before `make review`. Do not skip it.

### Refine An Existing Playbook (full chain)

```
fixture regression or weak match identified
  → playbook-refinement skill   (smallest credible change)
  → [if new fixtures needed] fixture-generation skill
  → playbook-linter skill        (PASS required before make review)
  → make review → make test
  → [if real corpus affected] make build + fixture stats --check-baseline
```

Trigger: a fixture match is wrong, weak, or confusable.  
Gate: `playbook-linter` PASS is required before `make review`.

### Classify A New Playbook With Ontology

New playbooks must carry ontology fields. The linter enforces this, but the values come from `docs/ontology.md`:

```yaml
domain: <dependency|runtime|container|auth|network|ci-config|test-runner|database|filesystem|platform|source>
class:  <e.g. lockfile-drift, missing-executable, interpreter-mismatch>
mode:   <concrete root cause slug, e.g. npm-ci-requires-package-lock>
```

The ontology is read-only metadata on the playbook. It does not change matching behavior.

The next useful upgrades should stay small and repo-native.

1. Make the new prompt set the default workflow surface for agents working in the repo.
2. Keep fixture ingestion and review as the only path for adding real-world evidence.
3. Bias ingestion runs toward source diversity across GitHub, GitLab, Stack Exchange, Discourse, and Reddit when useful evidence is available.
4. Keep playbook growth biased toward refinement over catalog expansion.
5. Strengthen workflow authoring inside playbooks by improving `likely_files`, `local_repro`, and `verify` for weak handoff cases. A structured sweep of the bundled playbooks using `faultline explain <id>` to spot placeholder or thin workflow fields is a high-value, low-risk improvement pass.
6. Add new agent workflows only when they map cleanly to an existing deterministic command or checked-in regression gate.
7. Keep new product surfaces hidden, flagged, or non-default until they are covered by deterministic tests, fixture gates when relevant, and `make cli-smoke` if they change shipped output.

## What Not To Add

Avoid adding workflows that are only general software advice, such as:

- broad "implement a feature" prompts
- repository-agnostic refactor guidance
- generic polish or cleanup checklists with no Faultline-specific validation path

If a workflow does not clearly improve fixture curation, playbook quality, deterministic validation, or workflow handoff quality, it probably does not belong here.
