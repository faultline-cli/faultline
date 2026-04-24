# Generate Playbook From Gap

Use this procedure to take a CI failure type from gap identification through to a validated, linted playbook with real and synthetic fixtures in a single session.

Called from: `agents/skills/playbook-sprint`

This is the end-to-end operating procedure. Each phase has an explicit stop condition. Do not skip phases.

---

## Inputs Required Before Starting

One of:
- a described failure type in one sentence (e.g. "Maven build fails when local repository cache is corrupted")
- a known coverage gap name (e.g. `maven-cache-poisoning`)
- a public failure URL (GitHub issue, GitLab issue, StackExchange question, Discourse thread, Reddit post)

Plus:
- output of `./bin/faultline list` to orient against the current catalog
- output of `faultline explain <nearest-neighbor-id>` for the most likely adjacent playbook

---

## Phase 1: Orient

**Goal:** confirm this is a genuine gap and name the root cause.

1. Run `./bin/faultline list` and scan for existing coverage in the relevant `domain` and `class`.
2. Identify the root cause in one sentence. If you cannot write one sentence, stop — return to the failure description and clarify before continuing.
3. Find the nearest existing playbook ID. Run `faultline explain <id>` on it.
4. Classify the failure in the ontology (`docs/ontology.md`): name `domain`, `class`, and `mode`.
5. Check `fixtures/real/baseline.json` for existing regression coverage in this class.

**Stop condition:** if the failure maps cleanly to an existing playbook at step 3, go to `playbook-refinement` instead. Do not proceed.

---

## Phase 2: Collect Real Evidence

**Goal:** attempt to find at least one real public log before generating synthetically.

Search for public examples using the five source types in `prompts/collect-coverage-evidence.md`:
- GitHub Actions workflow runs and issues
- GitLab CI public job traces and issues
- StackExchange (DevOps, SO, ServerFault)
- Discourse forums
- Reddit (r/devops, r/golang, r/node, r/docker, etc.)

For each usable source:
```bash
faultline fixtures ingest --adapter <adapter> --url <public-url>
./bin/faultline analyze <staged-file> --json
```

Record the evidence metadata:
```
source_url, adapter, ci_system, os, arch, language, language_version,
toolchain, toolchain_version, trigger, job_or_step, failure_class,
top_match_playbook, top_match_score, classification
```

Classify each sample using the five-tier scheme:
- **Covered** → switch to `playbook-refinement` for that playbook
- **Weakly covered** → this playbook is the refinement target; switch to `playbook-refinement`
- **Gap — known category** → continue; this is the evidence anchor for Phase 4
- **Gap — new category** → continue; this is a new playbook candidate
- **Noise** → reject; do not use

**Acceptable fallback:** if no public source is available or accessible, note "synthetic only" and proceed with a note in the final deliverable.

---

## Phase 3: Triage

**Goal:** confirm the gap warrants a new playbook before writing any YAML.

Apply the acceptance bar from `prompts/triage-unmatched-log.md`:

1. Is the root cause distinct from the nearest existing playbook — not just different wording?
2. Is the evidence stable enough for a deterministic match phrase? (Can you cite a verbatim log line?)
3. Is the fix path concrete enough to encode in `fix` and `workflow.verify`?
4. Is there a clear category placement in `playbooks/bundled/log/<category>/`?

If all four are yes: **keep** — continue to Phase 4.

If any are no: **reject** — record the justification and stop. Do not author a playbook.

---

## Phase 4: Generate Synthetic Fixtures

**Goal:** produce all three fixture variants before authoring any YAML.

Use the `fixture-generation` skill (`prompts/fixture-generation.md`).

Inputs to the skill:
- the target playbook ID (assigned now, even before YAML exists)
- the ontology classification from Phase 1
- the nearest confusable neighbor ID and its `match.any` phrases
- any real log lines collected in Phase 2 (use verbatim as the canonical anchor)

Produce:
| Variant | Path | Expectation |
|---------|------|-------------|
| canonical | `fixtures/minimal/<id>.yaml` | top-1 ≥ planned `base_score` |
| noisy | `fixtures/minimal/<id>-noisy.yaml` | top-1 ≥ planned `base_score` |
| near-miss | `fixtures/minimal/<id>-near-miss.yaml` | NOT top-1 |

Verify each variant before proceeding:
```bash
./bin/faultline analyze fixtures/minimal/<id>.yaml --json
./bin/faultline analyze fixtures/minimal/<id>-noisy.yaml --json
./bin/faultline analyze fixtures/minimal/<id>-near-miss.yaml --json
```

Note: the first two will not match the target playbook yet (it doesn't exist). This step confirms the nearest neighbor score and identifies what exclusion the near-miss will rely on. Use this to inform `match.none` in Phase 5.

**Stop condition:** if canonical and noisy both score below 0.4 against any playbook, the log lines may be too generic. Return to Phase 2 and find a more specific real example, or narrow the failure type.

---

## Phase 5: Author The Playbook

**Goal:** write the playbook YAML grounded in fixture evidence.

Follow the full authoring procedure in `prompts/author-new-playbook.md`.

Required YAML fields:

```yaml
id: <category>-<hyphenated-noun-phrase>
title: <Sentence case>
category: <auth|build|ci|deploy|network|runtime|test>
severity: <critical|high|medium|low>
base_score: <float 0.0–1.0>
tags: []
stage_hints: []
summary: |-
diagnosis: |-
fix: |-
validation: |-
match:
  any:
    - <phrase verbatim from canonical or noisy fixture>
  none:
    - <phrase from near-miss fixture that separates this from the neighbor>
workflow:
  likely_files: []
  local_repro: []
  verify: []
# Ontology fields (required for new playbooks)
domain: <from docs/ontology.md>
class: <from docs/ontology.md>
mode: <concrete root cause slug>
```

Pattern discipline:
- Every `match.any` phrase must appear verbatim in at least one fixture from Phase 4.
- Every `match.none` exclusion must appear verbatim in the near-miss from Phase 4.
- `base_score` must reflect how tightly the signal discriminates; prefer 0.7–0.9 for well-anchored patterns.

After writing the YAML, re-run fixture analysis to confirm scores:
```bash
./bin/faultline analyze fixtures/minimal/<id>.yaml --json
./bin/faultline analyze fixtures/minimal/<id>-noisy.yaml --json
./bin/faultline analyze fixtures/minimal/<id>-near-miss.yaml --json
```

Canonical and noisy must now rank the new playbook top-1 at or above `base_score`. Near-miss must withhold it.

---

## Phase 6: Lint

**Goal:** apply the quality gate before `make review`.

Run the `playbook-linter` skill against:
- the playbook YAML from Phase 5
- the canonical and noisy fixtures from Phase 4
- the near-miss fixture from Phase 4
- the `faultline explain <nearest-neighbor-id>` output from Phase 1

All 9 criteria must PASS. Criterion 9 (ontology completeness) is required for new playbooks.

**A single critical FAIL blocks Phase 7.** Fix all critical issues and re-run the linter before continuing.

---

## Phase 7: Validate

**Goal:** confirm the playbook holds in the full repository context.

```bash
make review
make test
make build
make fixture-check
```

If `make review` reports a pattern conflict with an existing playbook:
1. Identify which playbook is conflicting.
2. Return to Phase 5 and tighten `match.any` or add a `match.none` exclusion.
3. Re-run the linter (Phase 6) after changes.
4. Re-run validation.

Do not proceed past a `make review` conflict.

---

## Deliverable

A structured summary covering all phases:

```
Failure type: <one sentence root cause>
Ontology: domain=<> class=<> mode=<>
Nearest neighbor: <id>

Phase 2 — Real evidence:
  [evidence table or "no real source found — synthetic only"]

Phase 3 — Triage: keep | reject
  Justification: <why>

Phase 4 — Fixtures:
  canonical:  <path>  score=<>
  noisy:      <path>  score=<>
  near-miss:  <path>  correctly withholds=yes|no

Phase 5 — Playbook: <path> id=<>

Phase 6 — Linter: PASS | FAIL
  Issues resolved: <list or none>

Phase 7 — Validation:
  make review:       pass | conflict (resolved)
  make test:         pass | fail
  make build:        pass | fail
  make fixture-check: pass | fail

Boundary maintenance: <how the near-miss separates this from nearest neighbor>
```
