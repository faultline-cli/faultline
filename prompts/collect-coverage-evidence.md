# Collect Coverage Evidence

Use this workflow when the goal is to discover and collect raw CI failure evidence to identify playbook coverage gaps, not to promote a pre-selected fixture.

## Inputs

Read before starting:
- [`SYSTEM.md`](../SYSTEM.md)
- [`docs/fixture-corpus.md`](../docs/fixture-corpus.md)
- [`docs/playbooks.md`](../docs/playbooks.md)
- [`docs/agent-workflows.md`](../docs/agent-workflows.md)
- [`fixtures/staging/README.md`](../fixtures/staging/README.md)

## Goal

Build a broad, diverse batch of real CI failure samples from public sources and measure how well the current playbook catalog handles them. The output is either:

- confirmed coverage: the catalog matches the failure well
- a gap candidate: an unmatched or weakly matched case worth escalating to the ingestion or refinement workflow

This is an auditing workflow, not a promotion workflow. Its job is to produce a clear map of where coverage is strong and where it is thin.

## Required Approach

### 1. Orient Against The Current Catalog

Before collecting evidence, understand what the catalog already covers:

```bash
./bin/faultline list
```

Note which categories and failure modes are well represented and which are thin or absent. Resolve the existing playbook IDs against `playbooks/bundled/` to understand current pattern boundaries.

Use `fixtures/real/baseline.json` to see which failure classes already have accepted regression coverage.

### 2. Select Evidence Sources

Collect from at least three distinct source types per run. Prefer sources with direct, machine-produced log output over community discussions with excerpts.

Prioritized source types:
- **GitHub Actions public workflow runs** — direct job logs, full environment context, step timing, and exact tool versions via `GITHUB_*` env vars and setup step output
- **GitHub issues referencing CI failures** — user-reported with partial but real log content
- **GitLab CI public job traces** — direct log output with section markers and full runner metadata
- **StackExchange (DevOps, SO, ServerFault)** — structured questions with reproducible failure context
- **Discourse forums (tech communities)** — longer failure threads with environment metadata in replies
- **Reddit (r/devops, r/golang, r/node, r/docker)** — rapid coverage of emerging failure patterns

For each source, note:
- the CI system type (GitHub Actions, GitLab CI, CircleCI, Jenkins, etc.)
- the runtime environment (OS, architecture, language version, toolchain version)
- the job or step name where the failure occurred
- any environment variables or setup steps visible in the log
- the trigger event (push, pull_request, schedule, workflow_dispatch)

### 3. Collect A Stratified Sample

Target a minimum of 8–12 distinct failure samples per run, distributed across:

| Axis | Target |
|------|--------|
| CI system | ≥2 distinct systems |
| Language / ecosystem | ≥3 distinct languages or toolchains |
| Failure class | ≥4 distinct failure categories (auth, network, dependency, runtime, etc.) |
| Source type | ≥3 distinct adapters |

Do not collect more than two samples from the same repository, thread, or discussion.

### 4. Ingest And Analyze Each Sample

For each candidate:

```bash
faultline fixtures ingest --adapter <adapter> --url <public-url>
```

After staging, run the analyzer directly against any raw log you have access to:

```bash
./bin/faultline analyze <log-or-staged-file> --json
```

Record:
- `top_match.playbook_id` and `top_match.score` (or `null` if unmatched)
- `top_3` if more than one candidate is close
- whether the result is a confident top-1, a weak match, or no match

### 5. Classify Each Sample

For each collected sample, assign one of:

| Classification | Criteria |
|---|---|
| **Covered** | Top-1 score ≥ 0.7, correct playbook, no close confusable neighbor |
| **Weakly covered** | Top-1 matches but score < 0.7, or a wrong playbook ranks above the correct one |
| **Gap — known category** | Unmatched, but the failure class maps to an existing playbook category with a plausible authoring path |
| **Gap — new category** | Unmatched and the failure class represents a clear root cause with no existing playbook analog |
| **Noise** | Insufficient log content, workaround-only, ambiguous environment, or duplicate of existing fixture |

### 6. Escalate Actionable Gaps

For each gap candidate:

- Run `faultline fixtures review` to check for duplicate staging hints.
- Noise and duplicates: reject immediately.
- Gap — known category: continue with [`refine-playbook-from-fixture.md`](./refine-playbook-from-fixture.md).
- Gap — new category: continue with [`triage-unmatched-log.md`](./triage-unmatched-log.md) to confirm the case warrants a new playbook before authoring one.

Do not promote directly from this workflow. Coverage evidence collection ends at the classification step. Promotion happens in the ingestion pipeline or playbook refinement workflow.

### 7. Run The Baseline Check

After any staging or analysis activity:

```bash
make build
./bin/faultline fixtures stats --class real --check-baseline
```

Confirm the real corpus is still stable before escalating.

## Evidence Metadata To Capture

For each sample, record the following in your session notes:

```
source_url: <url>
adapter: <github-issue | gitlab-issue | stackexchange-question | discourse-topic | reddit-post>
ci_system: <GitHub Actions | GitLab CI | CircleCI | Jenkins | other>
os: <ubuntu-latest | macos-14 | windows-2022 | ...>
arch: <amd64 | arm64 | ...>
language: <go | python | node | rust | java | ...>
language_version: <exact version string if visible>
toolchain: <docker | npm | pip | cargo | maven | ...>
toolchain_version: <exact version string if visible>
trigger: <push | pull_request | schedule | workflow_dispatch | ...>
job_or_step: <name of the failing job or step>
failure_class: <auth | network | dependency | runtime | config | env | tls | ...>
top_match_playbook: <playbook ID or null>
top_match_score: <float or null>
classification: <covered | weakly-covered | gap-known | gap-new | noise>
notes: <any relevant environment specifics or confusable neighbors>
```

## Command Skeleton

```bash
# Audit current catalog coverage
./bin/faultline list

# Ingest each candidate
faultline fixtures ingest --adapter github-issue --url <url-1>
faultline fixtures ingest --adapter stackexchange-question --url <url-2>
faultline fixtures ingest --adapter gitlab-issue --url <url-3>

# Analyze staged or raw logs
./bin/faultline analyze <staged-file> --json

# Review all staged candidates together
faultline fixtures review

# Baseline gate
make build
./bin/faultline fixtures stats --class real --check-baseline
```

## Acceptance Bar

- the batch covers at least 3 distinct source types
- at least 3 distinct language ecosystems are represented
- every sample has a recorded classification
- gap candidates are escalated to the appropriate follow-on workflow rather than left unresolved
- the real corpus baseline still passes at the end of the run

## Deliverable

- a per-sample evidence table with source URL, environment metadata, top-match result, and classification
- a summary of the catalog's coverage across the batch (covered, weak, gap, noise counts)
- the specific gap candidates selected for follow-on action and which workflow they were handed to
- the baseline check result
