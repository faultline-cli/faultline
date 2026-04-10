# Faultline System

## Core Thesis

Faultline is a deterministic CLI for CI failure analysis.

Given a build log from a local run or CI job, Faultline should identify the most likely failure pattern, explain the root cause in plain language, and return a concrete fix with evidence pulled directly from the log.

## Product Shape

- Single-purpose CLI.
- Local-first, but easy to run inside any CI system.
- Deterministic pattern matching only.
- Text output for humans.
- JSON output for agents and automation.
- Optional workflow output for local and agentic follow-up.
- Docker-first distribution for portable CI usage.

## Main Workflow

1. Read log input from a file path or stdin.
2. Normalize the raw log into stable lines for matching.
3. Load bundled YAML playbooks.
4. Validate playbook structure before matching.
5. Match deterministic patterns against the normalized log.
6. Score and rank matches using explicit, stable rules.
7. Optionally enrich the likely diagnosis with recent local git repository context.
8. Return the result as formatted text or JSON.

## Primary Commands

- `faultline analyze <logfile>`
- `cat build.log | faultline analyze`
- `faultline analyze <logfile> --json`
- `faultline analyze <logfile> --git`
- `faultline analyze <logfile> --git --since 30d --repo .`
- `faultline list`
- `faultline explain <failure-id>`
- `faultline workflow <logfile>`

## Architectural Boundaries

- `cmd/main.go` owns CLI startup and command wiring.
- `internal/engine` owns log ingestion, normalization, and orchestration.
- `internal/playbooks` owns YAML loading, validation, and deterministic playbook ordering.
- `internal/matcher` owns pattern matching, evidence extraction, and scoring.
- `internal/output` owns text formatting and JSON serialization.
- `internal/workflow` owns deterministic next-step planning for local and agentic workflows.
- `internal/repo` owns local git scanning, history parsing, derived signals, and diagnosis correlation.
- `playbooks/` owns bundled failure definitions and should contain only deterministic rule data.

## Core Entities

### Playbook

```go
type Playbook struct {
    ID       string
    Title    string
    Category string

    Match struct {
        Any []string
    }

    Explanation string
    Fix         []string
}
```

### Result

```go
type Result struct {
    Playbook   Playbook
    Evidence   []string
    Score      int
    Confidence float64
}
```

## Key Invariants

- Deterministic output only.
- No ML or LLM dependence in product logic.
- Same log input must yield the same result every time.
- Playbook loading order must be stable.
- Matching and ranking must use explicit rules, never probabilistic inference.
- Evidence must come directly from matched log lines.
- JSON output must remain stable and automation-friendly.
- Text output should stay concise and actionable.
- Workflow output should stay deterministic and derived only from analysis results plus local repo context.
- Docker execution should not require runtime dependencies beyond the container image.

## Non-Goals

- No hosted webhook service.
- No PR comments or provider integrations in V1.
- No dashboards or frontend UI.
- No speculative rule engine abstractions.
- No fuzzy or semantic matching.
- No runtime network calls during analysis.

## Delivery Standard

- The CLI should work the same locally and inside CI.
- The Docker image should stay small and fast to start.
- Playbooks should cover a focused set of high-value CI failures first.
- Repo docs must stay aligned with the shipped CLI product.
