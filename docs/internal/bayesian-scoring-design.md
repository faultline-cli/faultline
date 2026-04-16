# Bayesian-Inspired Ranking And Delta Diagnosis Design

Status: proposed for v0.2.0

This document describes a deterministic scoring layer for Faultline. It does
not replace the current detectors, and it does not make the product
probabilistic. Its job is to combine explicit evidence more cleanly after
detection so the product can support better ranking, better drift diagnosis,
and safer automation while preserving the same-input same-output contract.

## Why This Exists

Faultline's v0.2.0 direction is not "become a general smart CI assistant." The
better direction is:

- deterministic prevention
- deterministic diagnosis
- deterministic remediation handoff
- safe automation against explicit evidence

That means the scoring layer should not be treated as an isolated reranking
experiment. It should become the shared evidence-fusion layer behind several
product surfaces:

- candidate reranking in `analyze`
- delta-differential diagnosis for "what changed?"
- better prioritization in `workflow`
- thresholding and summaries for the GitHub Action
- high-confidence local guard surfaces
- conservative fixture-backed priors

The scoring layer should stay invisible when it is not helping. When it does
help, it should explain exactly why.

## Goals

- keep deterministic detection unchanged
- rank only candidates already emitted by the existing detectors
- preserve same-input same-output behavior
- keep fixes and workflows deterministic
- make score explanations human-readable and JSON-serializable
- support optional delta-oriented diagnosis from repo changes
- expose stable signals that downstream automation can gate on
- use the accepted fixture corpus as a conservative offline prior source

## Non-Goals

- no runtime learning or model fitting
- no randomization
- no network calls
- no hidden state
- no change to playbook matching semantics
- no AI or ML in product logic
- no provider-specific logic in the core CLI
- no replacement of playbooks with opaque scoring rules

## Current Architecture

Today the main log-analysis path is:

1. `internal/engine` loads playbooks, reads input, normalizes log lines, and extracts lightweight context.
2. The engine dispatches to a detector through the detector registry.
3. The detector returns ranked `model.Result` values.
4. The engine persists history for the top result and optionally enriches that top result with git context.
5. `internal/output` and `internal/workflow` consume `Analysis.Results` in the final rank order.

Relevant current files:

- `internal/engine/analyzer.go`
- `internal/matcher/matcher.go`
- `internal/detectors/logdetector/logdetector.go`
- `internal/detectors/sourcedetector/sourcedetector.go`
- `internal/model/model.go`
- `internal/output/output_json.go`
- `internal/workflow/workflow.go`
- `internal/repo/correlate.go`
- `internal/repo/signals.go`
- `internal/fixtures/stats.go`

Current evidence and score structures already exist in `model.Result`:

- `Evidence`
- `EvidenceBy`
- `Explanation`
- `Breakdown`
- `Score`
- `Confidence`

That existing shape is already a strong foundation for a second ranking layer
because the raw evidence is explicit and structured.

## Proposed Layering

The proposed runtime layering is:

### Layer 1: Deterministic Detection

Existing detectors remain unchanged.

Input:

- normalized log lines or source snapshot
- deterministic playbooks
- extracted context

Output:

- candidate playbooks already deemed matches
- raw signals and evidence already attached to `model.Result`
- existing detector score and confidence

### Layer 2: Bayesian-Inspired Evidence Fusion

New optional reranking and delta layer.

Input:

- detected candidates from Layer 1
- playbook metadata
- existing evidence bundles
- optional repo and diff-derived signals when available
- checked-in priors and feature weights

Output:

- same candidates, optionally reordered
- additive score contributions
- stable explanation payload
- optional ranked delta cause classes
- optional automation-readiness hints derived from explicit evidence

### Layer 3: Deterministic Consumers

Existing behavior remains authoritative.

Input:

- ranked candidates from Layer 2 when enabled
- ranked candidates from Layer 1 when disabled

Output:

- top result
- top-N rendering
- workflow generation
- fix output
- JSON and human-readable output
- GitHub Action summaries and artifacts
- future guard and IDE consumers

## Integration Point

The smallest integration point is inside `internal/engine` after detector
results are produced and before history persistence, fingerprinting, repo
enrichment, and workflow selection are driven off the top result.

Proposed flow inside `engine.AnalyzeReader` and `engine.AnalyzeRepository`:

1. call the existing detector
2. if scoring is disabled, keep current behavior
3. if scoring is enabled, pass detected `[]model.Result` plus explicit context into the scoring layer
4. replace the result ordering with the reranked ordering
5. optionally attach delta findings and automation hints
6. continue with fingerprint, history, output, and workflow logic unchanged

This keeps detector implementations and selection semantics stable.

## Product Surfaces This Should Unlock

### `analyze`

Primary use:

- break close calls more reliably
- explain why the top result outranked neighboring candidates

Why it matters:

- this is the direct product quality improvement
- it reduces ranking surprises without changing playbook semantics

### `workflow`

Primary use:

- prioritize files, steps, and evidence using ranking and delta signals
- expose why the workflow chose those priorities

Why it matters:

- turns diagnosis into a deterministic next-step artifact

### Delta-Differential Diagnosis

Primary use:

- answer "what likely changed between green and red?"
- rank drift classes such as dependency, runtime, CI config, environment, deploy, or test data

Why it matters:

- gives users the most valuable next question after the top diagnosis

### GitHub Action

Primary use:

- gate summaries and artifacts on evidence quality
- expose stable machine-readable data for CI automation

Why it matters:

- lets the product ship into real pipelines without provider logic entering the CLI core

### `guard`

Primary use:

- use the same evidence-fusion logic for a high-confidence local prevention surface

Why it matters:

- supports the shift-left story without creating a second diagnosis system

## Minimal Interface Changes

Recommended new option fields:

- `app.AnalyzeOptions.BayesEnabled bool`
- `engine.Options.BayesEnabled bool`

Recommended CLI flags:

- `--bayes`
- optional later: `--delta`

Recommended model additions:

- add optional `Ranking` data to `model.Result`
- add optional `Delta` data to `model.Analysis`
- add optional automation hints only if they can be expressed as deterministic derived data, not policy decisions

Everything else can remain unchanged.

## What Stays Untouched

- detector matching semantics
- `match.any`, `match.all`, and `match.none`
- source-detector trigger and mitigation logic
- playbook YAML detection rules
- fix text and workflow content authored in playbooks
- existing top-N and output selection behavior when Bayes is disabled

## Feature Schema

The scoring layer should use a small, explicit, fixed feature set. Not every
product idea needs its own feature. Some ideas should consume the scoring
output rather than change the scoring model.

### `detector_score`

- Type: numeric
- Extraction: current `result.Score`
- Why useful: preserves the current deterministic ranking signal as the anchor

### `detector_confidence`

- Type: numeric
- Extraction: current `result.Confidence`
- Why useful: summarizes competitive separation already computed by the detector

### `candidate_separation`

- Type: numeric in `[0,1]`
- Extraction: normalized gap between this candidate and the next strongest baseline candidate
- Why useful: important for safe automation and for deciding how assertive workflow or action output should be

### `log_match_coverage`

- Type: numeric in `[0,1]`
- Extraction: proportion of declared `match.any` and `match.all` patterns satisfied for the candidate
- Why useful: broader coverage is usually stronger than a single incidental hit

### `error_exact_match`

- Type: boolean or normalized numeric
- Extraction: exact normalized match between a playbook pattern and the matched log evidence
- Why useful: exact signatures should outrank looser partial matches

### `error_fuzzy_overlap`

- Type: numeric in `[0,1]`
- Extraction: deterministic token overlap between matched patterns and evidence
- Why useful: retains nuance without semantic inference or randomness

### `stage_hint_match`

- Type: boolean
- Extraction: `Analysis.Context.Stage` against `Playbook.StageHints`
- Why useful: stage agreement is already an important discriminator

### `tool_or_stack_match`

- Type: boolean
- Extraction: deterministic tool tokens from command hints, evidence lines, playbook tags, and pattern text
- Why useful: distinguishes similar failures across Node, Python, Java, Go, Docker, Kubernetes, and CI tooling

### `file_path_relevance`

- Type: numeric in `[0,1]`
- Extraction: overlap between source evidence paths, workflow likely files, repo recent files, and playbook path hints
- Why useful: relevant files often separate superficially similar candidates

### `changed_files_relevance`

- Type: numeric in `[0,1]`
- Extraction: overlap between changed files or recent repo files and playbook-related files or patterns
- Why useful: recent edits are often the strongest practical signal

### `dependency_change_relevance`

- Type: boolean
- Extraction: changes to files such as `go.mod`, `go.sum`, `package.json`, lockfiles, `requirements*.txt`, `pyproject.toml`, `Gemfile`, or similar
- Why useful: important for install, resolver, and drift playbooks

### `runtime_toolchain_change_relevance`

- Type: boolean
- Extraction: changes to version files, CI setup actions, Docker base images, runner image references, or toolchain config files
- Why useful: important for runtime mismatch and missing executable failures

### `ci_config_change_relevance`

- Type: boolean
- Extraction: changes to workflow files, CI scripts, Makefiles, deploy scripts, and related orchestration config
- Why useful: identifies setup and ordering problems introduced in CI config

### `environment_drift_relevance`

- Type: numeric in `[0,1]`
- Extraction: repo hotfix, revert, repeated-edit, and repeated-directory signals combined with env and config changes
- Why useful: highlights likely drift and churn-driven failures

### `repo_delta_agreement`

- Type: numeric in `[0,1]`
- Extraction: overlap between the candidate's likely failure area and the top-ranked delta cause class
- Why useful: when log evidence and repo drift agree, automation and workflow output can be more confident without changing the diagnosis contract

### `historical_fixture_support`

- Type: numeric in `[0,1]`
- Extraction: static accepted-fixture support derived offline and stored in a checked-in weights file
- Why useful: conservative prior support for common real-world failures

### `mitigation_presence`

- Type: numeric in `[0,1]`
- Extraction: existing mitigation, suppression, and safe-context evidence on the candidate
- Why useful: serves as a bounded negative signal when evidence weakens a match

## Scoring Model

The model should be additive and fully deterministic:

`ranking_score = prior + sum(weight(feature) * value(feature))`

This is Bayesian-inspired rather than a runtime probabilistic model. The
important behavior is that the score is decomposed into a prior plus explicit
feature contributions. User-facing output should prefer terms like "ranking"
and "evidence contributions" over "probability."

## Priors

Priors should be defined offline from checked-in fixture data and stored in a
static file.

Recommended approach:

- start from accepted fixture counts per playbook
- smooth heavily toward category-level counts
- clamp to a narrow range so priors never dominate strong evidence
- optionally include pack-level counts only after pack metadata and fixture provenance are stable

Recommended shape:

- playbook prior
- category prior
- blended prior

The prior should be conservative and mainly break close calls between otherwise
plausible candidates.

## Feature Weights

Weights should live in a checked-in, versioned data file, for example:

- `internal/scoring/weights/bayes_v1.json`

Versioning goals:

- explicit diffs when weights change
- reproducible historical behavior
- simple rollback

V1 should use:

- global feature weights
- conservative fixture-derived priors
- optional category-level overrides only if clearly justified

V1 should avoid:

- runtime fitting
- per-playbook custom weights
- hidden generated artifacts
- per-pack runtime overrides

## Smoothing And Overfitting Control

To avoid overfitting on the current fixture corpus:

- use Laplace-style smoothing on priors
- clamp every feature contribution to a bounded range
- keep the feature set small
- keep `detector_score` as the strongest feature in V1
- require a minimum fixture count before category overrides exist
- avoid per-playbook feature weights in V1

The scoring layer should behave as a reranker and evidence synthesizer, not a
replacement detection engine.

## Determinism Rules

The scoring layer must:

- read only current input, playbooks, repo context, and checked-in weight data
- never write hidden state
- iterate in stable order
- sort ties deterministically
- round consistently at fixed points

Recommended tie-break order:

1. reranking score descending
2. original detector score descending
3. detector confidence descending
4. playbook ID ascending

## Explainability Model

Each ranked result should expose a stable explanation payload in addition to the
existing detector-oriented breakdown.

Recommended new result field:

- `Ranking`

Recommended fields inside it:

- `mode`
- `version`
- `baseline_score`
- `prior`
- `final_score`
- `contributions`
- `strongest_positive`
- `strongest_negative`

Each contribution should include:

- `feature`
- `value`
- `weight`
- `contribution`
- `direction`
- `reason`
- `evidence_refs`

Explainability requirements:

- human-readable
- JSON-serializable
- stable across runs
- sorted deterministically by absolute contribution, then feature name

Suggested human output shape:

- final score
- prior
- baseline detector score
- strongest positive signals
- strongest negative signals
- delta agreement when available

Suggested JSON behavior:

- keep existing output fields unchanged
- add an optional `ranking` field per result
- add an optional `delta` field at the analysis level

## Delta-Differential Diagnosis

The same framework should rank likely cause classes from repo state and changes
without changing the playbook detector path.

### Inputs

- repo history and churn signals from `internal/repo`
- changed files or recent files when available
- CI config changes
- dependency and runtime file changes
- optional success-state snapshot data when the release later supports it

### Outputs

Ranked likely causes such as:

- `dependency`
- `runtime_toolchain`
- `ci_config`
- `environment`
- `source_code`
- `test_data`
- `infra_config`

### Delta Feature Set

- changed dependency files
- changed runtime or version files
- changed CI workflow files
- changed environment or secret files
- changed deploy or infra files
- changed test-only files
- revert signal present
- hotfix signal present
- repeated edit hotspot present
- co-change with known failure area

### Delta Scoring

Use the same additive scoring shape:

`cause_score = cause_prior + sum(weight(feature) * value(feature))`

Delta diagnosis should remain optional and degrade cleanly when repo or diff
inputs are unavailable.

### Delta Explainability

The explanation format should match the playbook ranking explanation but use
delta-oriented reasons, for example:

- workflow file changed
- repeated edits in deploy directory
- lockfile changed with package manifest
- revert touched runtime config

## Expansion Areas Beyond Reranking

The roadmap ideas below should not all become new scoring features. Some should
be consumers of the same deterministic payload. This section breaks down where
the scoring layer helps and where it should stay out of the way.

### 1. Trust, Auditability, And Safe Automation

Developer and user experience impact:

- users need to understand why Faultline chose one diagnosis over another
- automation users need a stable signal for "safe enough to act on" without interpreting raw evidence themselves
- explanations should make close calls legible rather than magical

Product impact:

- this is the strongest differentiator versus probabilistic CI assistants
- it supports regulated, security-conscious, and procurement-heavy environments
- it turns determinism into a product guarantee instead of a technical detail

Technical architecture impact:

- scoring explanations and evidence references become part of the external contract
- introduce automation-readiness only as derived metadata from explicit features like match coverage, separation, mitigation presence, and delta agreement
- do not add policy decisions to the core engine; downstream consumers can choose thresholds

### 2. Workflow As A First-Class Surface

Developer and user experience impact:

- workflow output gets better when it can explain why certain files and steps are prioritized
- agents and scripts benefit from structured reasons instead of plain text lists
- deterministic workflow artifacts reduce the need for custom glue code

Product impact:

- this is where Faultline becomes operationally useful rather than diagnostic-only
- it strengthens the prevention and remediation story without introducing AI dependence

Technical architecture impact:

- `internal/workflow` should consume ranking and delta payloads, not recalculate them
- keep workflow generation declarative and deterministic
- if new artifacts such as patch templates or issue-follow-up plans are added later, they should be derived from playbook workflow fields plus ranking context

### 3. Delta-Differential Diagnosis

Developer and user experience impact:

- "what changed?" is often the first question after seeing a failure
- users can start from likely drift signals instead of manually diffing commits, workflows, and lockfiles

Product impact:

- this is the strongest "intelligence" feature that still fits Faultline's identity
- it materially improves triage speed without weakening determinism

Technical architecture impact:

- build on `internal/repo` rather than creating a separate history system
- introduce success-state snapshots only if they can be stored and compared in a minimal, explicit format
- delta should enrich ranking and workflow, not create a second diagnosis engine

### 4. GitHub Action

Developer and user experience impact:

- users get Faultline directly inside the failing workflow
- markdown summaries and JSON artifacts become the default handoff format

Product impact:

- this is the best near-term adoption engine
- it increases the flow of real logs, fixtures, and confidence signals

Technical architecture impact:

- keep GitHub-specific wiring in a separate action repository
- rely on CLI output contracts instead of embedding provider abstractions into `internal/`
- use ranking and delta payloads to improve summary quality, not to add GitHub-only logic

### 5. Fixture Corpus As The Moat

Developer and user experience impact:

- maintainers get clearer evidence when ranking changes help or hurt
- contributors can see where coverage is growing

Product impact:

- visible corpus growth is a stronger trust signal than vague accuracy claims
- empirical coverage becomes part of the product story

Technical architecture impact:

- fixture evaluation becomes the approval path for any weight or prior change
- weight generation should be offline and reviewable
- avoid coupling runtime code to mutable corpora or hidden caches

### 6. Pack Ecosystem And Monetization Boundary

Developer and user experience impact:

- teams can extend Faultline without forking the bundled catalog
- pack authors need predictable ranking behavior even when their playbooks join the candidate set

Product impact:

- packs are the clean path to ecosystem depth and future premium offerings
- keeping packs deterministic preserves the trust model across bundled and commercial content

Technical architecture impact:

- packs should compose into the same candidate and scoring pipeline
- pack metadata may later help with priors and visibility, but v0.2.0 should avoid pack-specific runtime weight overrides
- fixture provenance should eventually distinguish bundled versus pack-derived support if premium packs are introduced

### 7. Shift-Left Guard Surface

Developer and user experience impact:

- a local guard is only useful if it stays high-confidence and quiet
- the same explanation payload helps users understand why a guard fired before CI

Product impact:

- expands Faultline from failure analysis into failure prevention
- increases usage frequency and daily habit formation

Technical architecture impact:

- reuse source detector, repo signals, and scoring logic
- keep the surface narrow to avoid building a second general-purpose linter
- prefer explicit high-confidence failure classes over broad speculative checks

### 8. Optional AI Downstream, Never Core

Developer and user experience impact:

- users may still want generated explanations or patch suggestions
- those outputs are safer when grounded in deterministic evidence and workflow data

Product impact:

- creates a path to higher-leverage remediation UX without sacrificing the product's identity

Technical architecture impact:

- AI, if used, should consume exported JSON and workflow payloads downstream of the core engine
- no AI dependency should enter matching, ranking, or fix selection
- the scoring layer's job is to provide clean grounding data, not generated content

### 9. LSP And Deeper IDE Surfaces Later

Developer and user experience impact:

- inline explanations and code actions would benefit from the same ranking payload later

Product impact:

- useful future surface, but not necessary to prove workflow fit in v0.2.0

Technical architecture impact:

- no special architecture work is needed now beyond keeping outputs stable and structured

### 10. ZK Attestations And Cryptographic Trust Layers Later

Developer and user experience impact:

- almost no near-term users are blocked on this

Product impact:

- interesting future enterprise story, poor current return on effort

Technical architecture impact:

- should stay out of v0.2.0 design entirely
- first build the simpler trust layers: explicit evidence, stable schemas, deterministic reports, and policy-friendly outputs

## Package Structure

Keep the package structure small and direct:

- `internal/scoring/model.go`
- `internal/scoring/scorer.go`
- `internal/scoring/features.go`
- `internal/scoring/weights.go`
- `internal/scoring/explain.go`
- `internal/scoring/delta.go`
- `internal/scoring/weights/bayes_v1.json`

This is enough for v0.2.0. Deeper subpackages are not necessary yet.

## Rollout Plan

### Phase 1

- add the scoring layer behind `--bayes`
- leave the default behavior unchanged
- preserve byte-for-byte current output when the flag is off

### Phase 2

- add unit tests for feature extraction, scoring, rounding, and tie-breaks
- add engine tests confirming legacy behavior is unchanged when disabled
- compare legacy and Bayes modes across the existing corpus and fixture sets

### Phase 3

- add delta-differential output backed by repo signals
- route ranking explanations into workflow and JSON output
- use the same payload in the GitHub Action integration contract

### Phase 4

- extend fixture evaluation to track both legacy and Bayes ranking quality
- review top-1, top-3, unmatched, false-positive, and delta-quality metrics before promotion

### Phase 5

- consider enabling reranking by default only after it proves stable, deterministic, and non-regressive on the accepted corpus

## Risks

### Determinism Risk

- unstable map iteration
- inconsistent floating-point rounding
- unsorted evidence references
- ambiguous tie-breaks

Mitigation:

- stable ordering everywhere
- fixed rounding points
- explicit tie-break sequence

### Explainability Risk

- too many features can produce noisy explanations
- weak or redundant features can obscure the strongest signal

Mitigation:

- small fixed feature set
- bounded contribution counts in human output
- clear reason strings derived from real evidence

### Product Drift Risk

- the layer could become a catch-all for every new idea
- "Bayesian" language could be misread as probabilistic diagnosis

Mitigation:

- keep detection authoritative
- describe the layer as deterministic evidence fusion in user-facing contexts
- add new consumers before adding new features whenever possible

## Summary

The Bayesian-inspired layer should improve ranking while staying subordinate to
deterministic detection. Its real value in v0.2.0 is broader than reranking:
it is the smallest clean way to support delta-differential diagnosis, richer
workflow output, safer automation, fixture-backed trust, and future pack-aware
extension without changing the product's core identity.
