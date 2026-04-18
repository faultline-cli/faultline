# Faultline Roadmap

## Roadmap Review

The existing roadmap got the V1 foundations right:

- keep the CLI deterministic and fast
- preserve stable JSON and CI-friendly Docker delivery
- grow coverage through playbooks and fixture-driven validation
- keep Bayesian-inspired reranking optional and downstream of detection

The gap is not the foundation. The gap is what the roadmap makes primary.

Today the repository already contains several pieces that point beyond "log analyzer":

- `faultline workflow` already turns a diagnosis into a deterministic follow-up plan
- repo context and drift signals already exist under `internal/repo`
- pack installation and pack composition already exist
- fixture ingestion, review, promotion, and baseline checks already exist
- stable machine-readable JSON already exists for automation

That means v0.2.0 should not be framed as more of the same. It should prove that Faultline belongs in real workflows as a deterministic prevention, diagnosis, and remediation layer for CI/CD.

## v0.2.0 Theme

**From diagnosis to prevention**

Faultline v0.2.0 should make five things obvious:

- it is safe to automate against
- it produces deterministic next-step artifacts, not just diagnoses
- it can explain likely drift between green and red runs
- it is backed by a visible, growing real-world fixture corpus
- it fits naturally into CI and local developer workflows

## v0.2.0 Pillars

### 1. Reposition Around Deterministic, Audit-Friendly Automation

Current foundation:

- the product is already deterministic, local-first, and explainable
- README copy already gestures at trust, reproducibility, and automation safety

v0.2.0 work:

- rewrite top-level messaging around deterministic, explainable, reproducible analysis
- make "safe to automate against" a first-class promise in README, docs, and release notes
- explicitly frame output as audit-friendly and procurement-friendly for security-conscious teams
- keep user-facing language focused on guarantees rather than implementation trivia

Impact on developer and user experience:

- engineers understand faster when Faultline should be trusted in CI loops
- automation consumers get a clearer contract for when results are stable enough to drive actions
- the product feels more like infrastructure and less like a helper utility

Impact on product:

- sharpens the commercial wedge against probabilistic CI assistants
- improves category clarity for teams that care about traceability and repeatability
- makes future enterprise and policy packaging easier to explain

Impact on technical architecture:

- no large runtime change is required
- stable schemas, deterministic tie-breaks, and evidence provenance become even more important because they are now part of the product promise, not just implementation quality
- the Bayesian-inspired layer should expose explainability and automation-safety signals without changing match semantics

### 2. Make Workflow a First-Class Surface

Current foundation:

- `faultline workflow` already exists
- deterministic workflow text and JSON output already exist
- playbooks already carry workflow hints such as likely files, local repro commands, and verification commands

v0.2.0 work:

- promote workflow alongside `analyze` in README and examples
- add examples that show actionable artifacts instead of diagnosis-only output
- tighten the workflow schema so it is easier for scripts, agents, and CI steps to consume
- let workflow output incorporate ranking and delta context when available

Impact on developer and user experience:

- the user gets a next-step plan, not just a label
- teams can wire workflow JSON into scripts or agent prompts without inventing their own glue
- the CLI feels useful even when the user already understands the raw log

Impact on product:

- this is the bridge from diagnosis to remediation
- it makes Faultline operationally valuable instead of merely informative
- it sets up future deterministic patch suggestions or policy-guided remediation without requiring AI in the core path

Impact on technical architecture:

- `internal/workflow` should remain deterministic and consume analysis results rather than owning diagnosis logic
- the scoring layer can provide stable explanation and prioritization signals that workflow consumes
- any new workflow artifact types should be derived from playbook content, repo context, and checked-in rules only

### 3. Add Delta-Differential Diagnosis as a Core Capability

Current foundation:

- repo history loading and drift correlation already exist
- the Bayesian-inspired design already includes delta-oriented scoring concepts
- the engine already has a natural post-detection integration point for reranking

v0.2.0 work:

- pilot a deterministic Bayesian-inspired reranking layer behind `--bayes`
- add delta-oriented scoring that answers "what changed between green and red?"
- start with a minimal scope: dependency files, runtime files, CI workflow files, environment/config files, and deploy/infra files
- expose ranked drift signals and clear reasons in human output and JSON

Impact on developer and user experience:

- users get a much better first answer than "this failed"
- triage starts from likely drift, not a cold read of a long log
- repeated failures feel less opaque because Faultline can point at the most relevant recent changes

Impact on product:

- this is the most valuable intelligence upgrade that still fits the deterministic product identity
- it moves Faultline closer to prevention and faster remediation without turning it into a generic assistant
- it creates a strong release headline for v0.2.0

Impact on technical architecture:

- add a small `internal/scoring` package behind the engine, after detection and before workflow/output consumption
- keep detection, playbook matching, and fix content unchanged
- persist only minimal, explicit snapshot data when success-state comparison is introduced
- keep tie-breaks, rounding, and explanation ordering fully deterministic

### 4. Ship a GitHub Action as the Main Distribution Surface

Current foundation:

- Faultline already runs well in Docker and from the CLI
- JSON output is already stable enough to act as an external integration contract
- README already documents manual GitHub Actions usage

v0.2.0 work:

- ship a dedicated `faultline-action` repository
- support file input, stdin-like capture, markdown summary output, and JSON artifacts
- make the action a thin integration layer over the CLI rather than a provider-specific branch of the product
- document thresholds and failure behavior clearly so teams can adopt it safely

Impact on developer and user experience:

- integration becomes copy-paste simple
- users see Faultline where the failure already happens instead of having to reproduce setup manually
- CI maintainers get an easy path to structured artifacts and workflow follow-up

Impact on product:

- this is the highest-leverage distribution step
- it creates a natural funnel for external logs, issue reports, fixtures, and credibility
- it gives the product a real workflow foothold without requiring a hosted service

Impact on technical architecture:

- keep provider-specific code out of the core CLI
- treat CLI JSON and markdown output as the stable integration contract
- use the scoring and workflow layers to control summary quality, artifact richness, and automation safety without adding provider logic to `internal/`

### 5. Expose the Fixture Corpus as the Moat

Current foundation:

- the repository already has accepted real fixtures, baseline checks, and stats support
- the regression corpus already acts as the quality gate for ranking behavior

v0.2.0 work:

- publish a fixture stats document and link it from README
- show category coverage, fixture counts, and release-over-release growth
- document the contribution path for new real-world failures
- use the checked-in corpus as the source of conservative priors for the Bayesian-inspired reranker

Impact on developer and user experience:

- contributors understand where new evidence fits
- users can see that coverage is grounded in real failures, not just hand-written demos
- ranking behavior becomes easier to trust because it is visibly regression-tested

Impact on product:

- the corpus becomes an external trust signal, not just an internal engineering asset
- it strengthens the moat around empirical coverage and repeatability
- it supports future commercial packaging around premium and policy packs

Impact on technical architecture:

- keep fixture evaluation deterministic and checked in
- derive priors and weights offline from accepted fixtures, never at runtime
- preserve simple, reviewable data artifacts such as baseline files and versioned scoring weights

## Supporting Tracks for v0.2.0

### Pack Foundations Now, Registry Later

Current foundation:

- pack composition, installation, and auto-discovery already exist

v0.2.0 work:

- document the pack contract more clearly
- add example packs and install/versioning guidance
- make it obvious that packs are the extension and monetization boundary

Impact on developer and user experience:

- teams can extend coverage without forking core
- the extensibility model becomes easier to understand and test locally

Impact on product:

- packs become the clean path to ecosystem depth and future commercial packaging
- the roadmap stays focused on ecosystem readiness rather than registry theater

Impact on technical architecture:

- keep pack metadata simple and deterministic
- avoid registry, signing, or pack-specific runtime weight overrides in v0.2.0

### Pre-Commit and Pre-Push Guard

Current foundation:

- `inspect`, source detection, repo scanning, and workflow generation already exist

v0.2.0 work:

- add a lightweight `faultline guard` path for high-confidence local checks
- focus on known deterministic failure classes that are cheap to detect before CI
- use the same evidence and workflow model rather than inventing a separate subsystem

Impact on developer and user experience:

- catches some failures before cloud CI time is spent
- reinforces the shift-left story without requiring a heavy IDE integration first

Impact on product:

- strengthens the prevention narrative
- increases usage frequency and habit formation

Impact on technical architecture:

- reuse source detector and scoring primitives
- keep the scope narrow and high-confidence to avoid noisy local checks

### Soft Exposure for Complete Surfaces

Current foundation:

- `trace`, `replay`, `compare`, `inspect`, and `guard` are already implemented and validated
- release-boundary policy already separates default onboarding from companion surfaces

v0.2.0 work:

- add a clear "already built but non-default" section in top-level docs
- expose one explicit opt-in discovery path for experimental provider-backed delta
- keep default help and default narrative focused on stable commands
- track usage and feedback from power users before graduation decisions

Impact on developer and user experience:

- power users can discover depth without digging through source
- contributors get clearer promotion targets for command graduation
- new users keep a simple first-run path with low cognitive load

Impact on product:

- improves discoverability of already-shipped depth
- avoids feature rot caused by zero user pressure on complete surfaces
- preserves trust by keeping release discipline intact

Impact on technical architecture:

- no runtime architecture change required
- preserves hidden-flag and release-boundary enforcement in CLI behavior
- requires docs, examples, and promotion metrics to stay synchronized

### Fixture Sanitizer for Safe Contribution

Current foundation:

- fixture ingestion, review, and promotion flows already exist
- real-world fixture quality is already a core trust boundary

v0.2.0 work:

- add a simple `faultline fixtures sanitize` path for contribution prep
- mask obvious secrets and sensitive identifiers before fixture intake
- start with conservative deterministic rules (tokens, emails, auth headers, obvious credential patterns)
- document sanitizer limitations clearly so contributors know what still needs manual review

Impact on developer and user experience:

- lowers friction for submitting real CI failures
- makes contribution safer for teams with strict data-handling policies
- increases fixture throughput without requiring perfect automation

Impact on product:

- accelerates growth of the real-world corpus
- strengthens regression evidence and trust in diagnosis coverage
- improves community participation in playbook refinement

Impact on technical architecture:

- keep sanitizer deterministic and rule-based
- keep masking rules explicit, versioned, and test-backed
- integrate sanitizer output directly into existing fixtures ingest/review flow

## Suggested Build Order

### Slice 1: Action-First Adoption

- ship the GitHub Action as the default onboarding route
- publish copy-paste workflow examples that emit markdown summary plus JSON workflow artifacts
- document deterministic gating policy for downstream automation

### Slice 2: Workflow-First Product Story

- promote analyze -> workflow as the hero path in README and examples
- tighten workflow JSON compatibility guidance as an API contract
- keep schema changes additive by default

### Slice 3: Messaging and Proof

- rewrite README and launch framing around deterministic, audit-friendly automation
- make "safe to automate against" the top-level trust promise
- publish fixture corpus stats with category coverage and release snapshots
- expose complete-but-non-default surfaces without expanding default onboarding

### Slice 4: Distribution and Extensibility

- publish pack examples and pack authoring guidance

### Slice 5: Deterministic Intelligence

- land the Bayesian-inspired reranking layer behind `--bayes`
- add delta-differential diagnosis and explanation payloads
- compare legacy and Bayes modes across the fixture corpus before promotion

### Slice 6: Shift-Left Follow-Through

- add a lightweight guard command
- add `faultline fixtures sanitize` for safe fixture submission
- polish docs, examples, thresholds, and automation handoff contracts

## Later, Not v0.2.0

- deep IDE or LSP integration
- hosted pack registry
- premium pack delivery infrastructure
- AI-generated fixes in the core execution path
- signed packs and enterprise governance layers
- cryptographic or ZK-style attestations

## Roadmap Standard

Changes on this roadmap should keep the same core invariants:

- deterministic detection stays authoritative
- same input and same checked-in rule set produce the same output
- scoring only reranks explicit candidates and remains explainable
- workflow and automation output stay stable and auditable
- fixture-backed evaluation gates any ranking change before it becomes the default
