# Playbook Authoring

Faultline playbooks separate deterministic matching from human guidance:

- structured YAML fields decide whether a playbook matches
- markdown fields explain the diagnosis and recovery steps

## Authoring rule

Structured fields decide; markdown explains.

Do not hide matching logic, ranking hints, or machine-important state inside prose.

## Recommended content fields

Use these markdown-capable string fields for operator-facing guidance:

- `summary`
- `diagnosis_markdown`
- `fix_markdown`
- `validation_markdown`
- `why_it_matters_markdown` (optional)

Use YAML block scalars for each field:

```yaml
summary: |
  One-line summary for ranked output.

diagnosis_markdown: |
  ## Diagnosis

  Explain the likely root cause in plain language.

fix_markdown: |
  ## Fix steps

  1. Keep steps short and operational.
  2. Use short code fences only when they clarify the action.

validation_markdown: |
  ## Validation

  - Re-run the relevant command.
  - Confirm the original failure signature is gone.
```

## Writing guidelines

- Keep `summary` to one or two sentences.
- Prefer short headings such as `## Diagnosis` and `## Validation`.
- Keep bullet lists and numbered steps concise.
- Use short code fences for exact commands, not long scripts.
- Put deterministic commands in `workflow.local_repro` and `workflow.verify` as well as the markdown if they matter operationally.
- Do not hide branching logic or detector assumptions inside markdown prose.
- `summary`, `diagnosis_markdown`, `fix_markdown`, and `validation_markdown` are required for shipped playbooks.

## Improvement pipeline

Treat playbook growth as a deterministic review loop, not a content-volume goal.

1. Ingest evidence from bundled playbooks, clean fixtures, noisy corpus logs, missed detections, false positives, and repository inspection findings.
2. Normalize each candidate into a root-cause record with likely category, distinctive signatures, confusable neighbors, and an actionable fix path.
3. Cluster by underlying failure mechanism, not by wording. Reject vague or duplicate clusters before authoring anything.
4. Prefer improving the strongest nearby playbook over adding a shallow variant. Add a new playbook only when the root cause is distinct and the signals are defensible.
5. For every accepted playbook, add at least one positive fixture and one nearby negative or adversarial regression so ranking stays stable in noisy logs.
6. Re-run `make review` after edits to inspect shared patterns and `make test` to confirm fixture, corpus, and ranking regressions remain deterministic.

Acceptance bar:

- one dominant playbook per root cause unless the detection boundary is genuinely different
- distinctive signals over broad wording
- short, ordered fixes tied to the root cause
- explicit negative signals when a nearby false positive is known
- shipped playbooks must be defendable against at least one confusable example

## Pack composition

Faultline ships the starter catalog from `playbooks/bundled/` and composes any premium or team-specific packs on top of it.

Use this boundary when deciding where a playbook belongs:

- bundled: high-frequency failures across common stacks or CI systems, plus enough baseline source coverage for `inspect` to produce useful starter results
- premium: provider-specific workflows, advanced deployment and platform operations, security-heavy rules, and deeper source-detector coverage beyond the starter baseline

There are three supported ways to add extra packs:

1. `faultline packs install <dir>` to copy a pack into `~/.faultline/packs/` for automatic loading on future runs
2. repeat `--playbook-pack <dir>` for one-off or scripted composition
3. set `FAULTLINE_PLAYBOOK_PACKS` for environment-driven composition

Use `--playbooks <dir>` only for full catalog overrides such as testing a pack in isolation.

For shipped premium packs, prefer `faultline packs install` as the customer-facing path. It survives binary upgrades, avoids repeated flags, and works with the Docker image when `~/.faultline` is mounted into `/home/faultline/.faultline`.
