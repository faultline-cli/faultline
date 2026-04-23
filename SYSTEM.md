# Faultline System

## Core Thesis

Faultline is a deterministic failure reasoning engine for CI failures.

Given a build log from a local run or CI job, Faultline should identify the most likely failure pattern, explain the root cause in plain language, materialize a structured failure artifact, and return a concrete fix with evidence pulled directly from the log.

## Product Shape

- Single-purpose CLI.
- Local-first, but easy to run inside any CI system.
- Deterministic pattern matching for authoritative detection.
- Optional deterministic Bayesian-inspired reranking for evidence fusion.
- Text output for humans.
- JSON output for agents and automation.
- Optional workflow output for local and agentic follow-up.
- Optional quiet guard output for high-confidence local prevention checks.
- First-class failure artifacts for storage, replay, compare, and remediation handoff.
- Docker-first distribution for portable CI usage.

## Main Workflow

1. Observe: read log input from a file path or stdin, then normalize the raw log into stable lines for matching.
2. Resolve: load bundled YAML playbooks, validate playbook structure, and match deterministic patterns against the normalized log.
3. Materialize: score and rank matches using explicit, stable rules, then build a deterministic differential diagnosis and first-class failure artifact.
4. Enrich: attach recent local git repository context, explicit change signals, and local history when available.
5. Act: return the result as formatted text, JSON, workflow output, or a remediation handoff.
6. Learn: refine playbooks through deterministic fixtures, overlap review, and regression gates so future failures resolve faster.

## Primary Commands

- `faultline analyze <logfile>`
- `faultline inspect <path>`
- `cat build.log | faultline analyze`
- `faultline analyze <logfile> --json`
- `faultline analyze <logfile> --bayes`
- `faultline analyze <logfile> --git`
- `faultline analyze <logfile> --git --since 30d --repo .`
- `faultline list`
- `faultline explain <failure-id>`
- `faultline workflow <logfile>`
- `faultline workflow <logfile> --bayes`
- `faultline guard <path>`
- `faultline history`
- `faultline history --signature <hash>`
- `faultline signatures`
- `faultline verify-determinism <logfile>`
- `faultline fixtures ingest`
- `faultline fixtures review`
- `faultline fixtures promote`
- `faultline fixtures stats`

## Architectural Boundaries

- `cmd/main.go` owns CLI startup and command wiring.
- `internal/engine` owns log ingestion, source tree scanning, normalization, and orchestration.
- `internal/engine/hypothesis` owns deterministic differential diagnosis across competing playbooks.
- `internal/detectors` owns detector module interfaces and target contracts.
- `internal/playbooks` owns pack resolution, YAML loading, validation, and deterministic playbook ordering.
- `internal/matcher` owns log-pattern matching, evidence extraction, and scoring.
- `internal/scoring` owns optional Bayesian-inspired evidence fusion, additive ranking explanations, and delta diagnosis.
- `internal/output` owns text formatting and JSON serialization.
- `internal/workflow` owns typed remediation workflow schemas, binding,
  dry-run planning, policy-gated execution, verification, and persisted
  workflow execution records.
- `internal/repo` owns local git scanning, history parsing, derived signals, and diagnosis correlation.
- `internal/repo/topology` owns CODEOWNERS parsing, repository ownership graph construction, and topology signal derivation.
- `internal/fixtures` owns deterministic fixture corpora, source adapters, curation workflow, and regression gates.
- `internal/playbooks` also owns playbook overlap reporting for deterministic review of shared patterns and exclusions.
- `playbooks/` owns bundled and external-pack boundaries and should contain only deterministic rule data or pack metadata.

## Core Entities

### Playbook

```go
type Playbook struct {
    ID         string
    Title      string
    Category   string
    Severity   string
    Detector   string
    BaseScore  float64
    Tags       []string
    StageHints []string
    Match      MatchSpec     // Any (OR), All (AND), None (exclusion)
    Source     SourceSpec    // Trigger/amplifier/mitigation signal matchers
    Summary    string
    Diagnosis  string
    Fix        string
    Validation string
    Workflow   WorkflowSpec  // LikelyFiles, LocalRepro, Verify steps
    Scoring    ScoringConfig
    Contextual ContextPolicy
    Hypothesis HypothesisSpec
}
```

### Result

```go
type Result struct {
    Playbook   Playbook
    Detector   string
    Evidence   []string
    EvidenceBy EvidenceBundle  // Triggers, Amplifiers, Mitigations, Suppressions
    Score      float64
    Confidence float64
    SeenCount  int
    Ranking    *Ranking  // Populated when --bayes is enabled
    Hypothesis *HypothesisAssessment
}
```

### Analysis

```go
type Analysis struct {
    Results     []Result
    Context     Context          // Stage, CommandHint, Step
    Source      string
    Fingerprint string
    RepoContext *RepoContext     // Populated by default local repo enrichment unless disabled
    Delta       *Delta           // Populated when repo-aware scoring has explicit change or git context
    Differential *DifferentialDiagnosis
    Artifact    *FailureArtifact // Stable unit for replay, compare, storage, and remediation handoff
}
```

### FailureArtifact

```go
type FailureArtifact struct {
    Fingerprint         string
    MatchedPlaybook     *ArtifactPlaybook
    Evidence            []string
    Confidence          float64
    Environment         ArtifactEnvironment
    HistoryContext      *ArtifactHistoryContext
    FixSteps            []string
    CandidateClusters   []CandidateCluster
    DominantSignals     []string
    SuggestedPlaybookSeed *SuggestedPlaybookSeed
    Remediation         *RemediationPlan
}
```

### Ranking (--bayes only)

```go
type Ranking struct {
    Mode              string   // always "bayes"
    Version           string
    BaselineScore     float64
    Prior             float64
    FinalScore        float64
    Contributions     []RankingContribution
    StrongestPositive []string
    StrongestNegative []string
}
```

## Key Invariants

- Deterministic output only.
- No ML or LLM dependence in product logic.
- Same log input must yield the same result every time.
- Playbook loading order must be stable.
- Deterministic detectors stay authoritative for matching.
- Optional Bayesian-inspired reranking may assist ranking and delta diagnosis, but it must stay explainable, additive, and reproducible.
- Local repo-aware enrichment is part of the default analysis loop; explicit flags may still disable it for narrow cases.
- Provider-backed CI delta is experimental, requires explicit opt-in, and may call the provider API only when enabled.
- Evidence must come directly from matched log lines.
- JSON output must remain stable and automation-friendly.
- Text output should stay concise and actionable.
- Workflow output should stay deterministic and derived only from the analysis artifact plus local repo context.
- Playbook review output should be deterministic and should highlight exact shared patterns and exclusions.
- Docker execution should not require runtime dependencies beyond the container image.

## Non-Goals

- No hosted webhook service.
- No PR comments or provider integrations in V1.
- No dashboards or frontend UI.
- No speculative rule engine abstractions.
- No fuzzy or semantic matching.
- No runtime network calls during analysis unless explicit provider-backed delta resolution is enabled.

## Delivery Standard

- The CLI should work the same locally and inside CI.
- The Docker image should stay small and fast to start.
- Playbooks should cover a focused set of high-value CI failures first.
- Repo docs must stay aligned with the shipped CLI product.
