# Faultline Product Spec

> **AGENTS: Do not derive architecture, module boundaries, or data models from this document.**
> For implementation decisions, [`SYSTEM.md`](../SYSTEM.md) is the single authoritative source.
> This document describes product positioning and user-facing behavior only.

## Positioning

### Faultline

Run it when CI fails. Fix it in seconds.

### Core Value

- Deterministic CI failure detection
- Actionable fixes
- Works locally and in CI through Docker
- Agent-compatible JSON output

## Target User

- IC engineers
- Startup backend developers
- Indie hackers
- Teams using GitHub Actions, GitLab CI/CD, and similar systems

## Core Use Case

CI fails, the user runs Faultline against the log, and Faultline returns the most likely root cause plus an immediate fix.

## V1 Feature Set

### 1. Log Analysis

Commands:

```bash
faultline analyze build.log
cat build.log | faultline analyze
```

Text output:

```text
Detected: Docker registry authentication failure (67%)
Category: auth

Cause:
The CI job could not authenticate with the container registry before pulling or pushing an image.

Fix:
1. Verify the registry username, token, or password configured in CI.
2. Ensure the registry login step runs before any docker pull or docker push command.
3. Confirm the token has the correct repository scope for the image being accessed.

Evidence:
- pull access denied
- Error response from daemon: authentication required
```

### 2. JSON Output

Command:

```bash
faultline analyze build.log --json
```

Schema:

```json
{
  "matched": true,
  "failure_id": "docker-auth",
  "title": "Docker registry authentication failure",
  "category": "auth",
  "confidence": 0.67,
  "cause": "The CI job could not authenticate with the container registry before pulling or pushing an image.",
  "fix": [
    "Verify the registry username, token, or password configured in CI.",
    "Ensure the registry login step runs before any docker pull or docker push command.",
    "Confirm the token has the correct repository scope for the image being accessed."
  ],
  "evidence": [
    "pull access denied",
    "Error response from daemon: authentication required"
  ]
}
```

No-match schema:

```json
{
  "matched": false,
  "message": "No known failure pattern matched."
}
```

### 3. Playbook System

- YAML-based
- Deterministic matching
- Broad bundled coverage for high-frequency CI failures across common stacks, with additional packs reserved for deeper provider-specific and advanced workflows

### 4. CLI Utilities

```bash
faultline list
faultline explain docker-auth
```

### 5. CI Integration

Docker-first usage:

```bash
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/build.log
```

Example GitHub Actions step:

```yaml
- name: Faultline analysis
  if: failure()
  run: |
    docker run --rm -v ${{ github.workspace }}:/workspace faultline analyze /workspace/build.log
```

## System Architecture

### High-Level Flow

```text
[CLI]
   ↓
[Analyzer Engine]
   ↓
[Playbook Loader]
   ↓
[Matcher]
   ↓
[Scoring + Ranking]
   ↓
[Formatter (text/json)]
```

### Modules

1. CLI layer
2. Engine layer
3. Playbook layer
4. Matcher
5. Output layer

### Module Responsibilities

- CLI layer: command parsing, flags, input and output handling
- Engine layer: log ingestion, normalization, orchestration
- Playbook layer: YAML loading, validation, structure
- Matcher: deterministic pattern matching, evidence extraction, scoring
- Output layer: text formatting and JSON serialization

## Data Model

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

### JSON Result

```go
type JSONResult struct {
    Matched    bool     `json:"matched"`
    FailureID  string   `json:"failure_id,omitempty"`
    Title      string   `json:"title,omitempty"`
    Category   string   `json:"category,omitempty"`
    Confidence float64  `json:"confidence,omitempty"`
    Cause      string   `json:"cause,omitempty"`
    Fix        []string `json:"fix,omitempty"`
    Evidence   []string `json:"evidence,omitempty"`
    Message    string   `json:"message,omitempty"`
}
```

### Optional Future Context

```go
type AnalysisContext struct {
    Source    string
    Timestamp int64
}
```

## File Structure

```text
faultline/
  cmd/
    main.go
  internal/
    engine/
    matcher/
    loader/
    output/
  playbooks/
    *.yaml
  Dockerfile
  README.md
```

## Packaging

Primary bundle:

- CLI binary
- bundled playbooks
- docs
- Docker instructions

Delivery format:

- tar.gz archive

Indicative price:

- $19 to $29

## Messaging

### Core hook

CI failed? Run this. Get the fix instantly.

### Secondary message