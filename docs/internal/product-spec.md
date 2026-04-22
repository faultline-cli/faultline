# Faultline Product Spec

> **AGENTS: Do not derive architecture, module boundaries, or internal data models from this document.**
> [`SYSTEM.md`](../../SYSTEM.md) and the shipped CLI help remain authoritative for implementation details.
> This document describes the current product shape and user-facing behavior.

## Positioning

Faultline is a deterministic CLI for CI failure diagnosis.

Run it against a failing build log and it returns:

- the most likely failure class
- evidence pulled directly from the log
- checked-in diagnosis and fix guidance
- stable text, markdown, JSON, and workflow artifacts for humans and automation

The product stays deliberately narrow:

- local-first by default
- deterministic matching remains authoritative
- optional ranking assistance is additive, never a second matcher
- no ML or LLM dependence in shipped product logic

## Default User Story

The release boundary is intentionally small:

1. Run `faultline analyze <logfile>` on a failing CI log.
2. Review the evidence-backed top diagnosis.
3. Run `faultline workflow <logfile>` when you want a deterministic follow-up artifact.
4. Use `faultline list`, `faultline explain <id>`, or `faultline fix <logfile>` to inspect the catalog or narrow to remediation.

That boundary is documented in [`docs/release-boundary.md`](../release-boundary.md) and should remain the default narrative in user-facing docs.

## Command Surface

### Ship-ready core

```text
faultline analyze [file]
faultline workflow [file]
faultline list
faultline explain <id>
faultline fix [file]
```

### Supported companion surfaces

```text
faultline trace [file]
faultline replay <analysis.json>
faultline compare <left-analysis.json> <right-analysis.json>
faultline inspect [path]
faultline guard [path]
faultline packs install <dir>
faultline packs list
```

### Hidden maintainer workflows

The repository also includes hidden `faultline fixtures ...` workflows and a hidden scaffold helper for corpus curation and playbook authoring. These are supported for maintainers, but they are not part of the first-run product story.

## Core Behaviors

### 1. Log diagnosis

`faultline analyze` accepts a file path or stdin and returns deterministic ranked results.

Current user-facing characteristics:

- terminal output supports `quick` and `detailed` modes
- markdown output is available for CI summaries and docs snapshots
- JSON output is stable and automation-friendly
- focused views are supported through `--view summary|evidence|fix|raw|trace`
- ranked drill-down is supported through `--top`, `--select`, `--trace`, and `--trace-playbook`
- `--git` can enrich the diagnosis with recent local repository context
- `--bayes` can rerank already-matched candidates additively

### 2. Deterministic workflow handoff

`faultline workflow` turns the top diagnosis into a follow-up plan.

Current modes:

- `--mode local` for a practical local triage checklist
- `--mode agent` for a structured agent handoff artifact

The JSON workflow artifact uses `workflow.v1` and may include additive hints such as `ranking_hints`, `delta_hints`, `metrics_hints`, and `policy_hints` when the underlying analysis contains that context.

### 3. Catalog inspection

`faultline list` and `faultline explain <id>` expose the checked-in playbook catalog.

The catalog is authored in YAML, stored in version control, and loaded deterministically from the bundled pack plus any explicitly configured extra packs.

### 4. Narrow remediation view

`faultline fix` prints the remediation guidance for the top diagnosis without the rest of the analysis view.

### 5. Companion inspection surfaces

These remain important, but they are not the first-run story:

- `trace` for rule-by-rule evaluation
- `replay` for deterministic re-rendering of saved analysis artifacts
- `compare` for deterministic diffing of two saved analysis artifacts
- `inspect` for source-detector findings in a repository tree
- `guard` for quiet high-confidence prevention findings on changed files
- `packs` for optional playbook-pack composition

## Artifact Contracts

### Analysis JSON

`faultline analyze --json` and `faultline inspect --json` emit a stable additive analysis object.

At a high level, the current object includes:

- top-level `matched`
- top-level `source`, `fingerprint`, and `context`
- ranked `results`
- optional `input_hash` and `output_hash`
- optional `pack_provenance`
- optional `metrics` and `policy`

Each ranked result currently carries fields such as:

- `rank`
- `failure_id`
- `title`
- `category`
- `severity`
- `detector`
- `score`
- `confidence`
- `summary`
- `diagnosis`
- `why_it_matters`
- `fix`
- `validation`
- `evidence`
- `evidence_by`
- `breakdown`
- optional recurrence fields including `signature_hash`, `occurrence_count`, `first_seen_at`, and `last_seen_at`

Absent additive fields should stay absent rather than being populated with placeholder values.

### Workflow JSON

`faultline workflow --json` emits a deterministic `workflow.v1` object.

The current schema includes fields such as:

- `schema_version`
- `mode`
- `failure_id`
- `title`
- `source`
- `context`
- `evidence`
- `files`
- `local_repro`
- `verify`
- `steps`
- `agent_prompt` when `--mode agent`

Additive workflow hints remain allowed, but silent renames or removals of established fields should be treated as breaking changes.

## Playbook Model

Faultline playbooks separate deterministic matching from operator-facing guidance:

- structured fields drive matching, scoring, workflow derivation, and detector behavior
- markdown-capable fields such as `summary`, `diagnosis`, `fix`, and `validation` carry the human guidance

The current repository supports both `log` and `source` detectors. `inspect` and `guard` are the main public source-detector surfaces.

## Distribution And Packaging

The product ships as:

- a standalone CLI binary
- bundled playbooks under `playbooks/bundled/`
- release tarballs
- a Docker image build path

The installer script wraps the binary with `FAULTLINE_PLAYBOOK_DIR` pointing at the bundled playbooks under the install prefix.

## Product Boundaries

Faultline is intentionally not:

- a hosted CI platform
- a dashboard product
- a PR-comment bot by default
- a generic automation engine
- a fuzzy or semantic log analysis tool

Provider-backed delta, hooks, and maintainer authoring helpers exist behind explicit or hidden paths, but they do not define the default product story.
