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
- `diagnosis`
- `fix`
- `validation`
- `why_it_matters` (optional)

Use YAML block scalars for each field:

```yaml
summary: |
  One-line summary for ranked output.

diagnosis: |
  ## Diagnosis

  Explain the likely root cause in plain language.

fix: |
  ## Fix steps

  1. Keep steps short and operational.
  2. Use short code fences only when they clarify the action.

validation: |
  ## Validation

  - Re-run the relevant command.
  - Confirm the original failure signature is gone.
```

Optional delta-aware ranking fields:

```yaml
requires_delta: true
delta_boost:
  - signal: delta.dependency.changed
    weight: 1.2
```

Use these only when a playbook becomes meaningfully more precise once the
failure is compared against a baseline successful run.

Optional differential-diagnosis fields:

```yaml
hypothesis:
  supports:
    - signal: dependency.resolution.conflict
      weight: 0.7
  contradicts:
    - signal: dependency.lockfile.sync_error
      weight: -0.6
  discriminators:
    - description: Resolver wording points to incompatible requirements.
      signal: dependency.resolution.conflict
  excludes:
    - signal: dependency.hash.mismatch
```

Use `hypothesis` when a playbook has nearby confusable rivals and you want the
detailed output to explain why one wins, what evidence weakens it, and what
would rule it out entirely.

Signal IDs must be deterministic. Faultline currently supports curated signal
aliases plus a small generic set:

- named aliases such as `dependency.cache.corrupt`, `runtime.node.version.mismatch`, and `test.timeout.detected`
- `log.contains:<text>`
- `log.absent:<text>`
- `delta.signal:<id>`
- `delta.absent:<id>`
- `context.stage:<stage>`
- `context.stage.absent:<stage>`

## Writing guidelines

- Keep `summary` to one or two sentences.
- Prefer short headings such as `## Diagnosis` and `## Validation`.
- Keep bullet lists and numbered steps concise.
- Use short code fences for exact commands, not long scripts.
- Put deterministic commands in `workflow.local_repro` and `workflow.verify` as well as the markdown if they matter operationally.
- Do not hide branching logic or detector assumptions inside markdown prose.
- `summary`, `diagnosis`, `fix`, and `validation` are required for shipped playbooks.

## Improvement pipeline

Treat playbook growth as a deterministic review loop, not a content-volume goal.

1. Ingest evidence from bundled playbooks, clean fixtures, noisy corpus logs, missed detections, false positives, and repository inspection findings.
2. Normalize each candidate into a root-cause record with likely category, distinctive signatures, confusable neighbors, and an actionable fix path.
3. Cluster by underlying failure mechanism, not by wording. Reject vague or duplicate clusters before authoring anything.
4. Prefer improving the strongest nearby playbook over adding a shallow variant. Add a new playbook only when the root cause is distinct and the signals are defensible.
5. For every accepted playbook, add at least one positive fixture and one nearby negative or adversarial regression so ranking stays stable in noisy logs.
6. Re-run `make review` after edits to inspect shared patterns and `make test` to confirm fixture, corpus, and ranking regressions remain deterministic.

Review interpretation:

- `make review` and the composed-pack review checks are overlap inspection tools, not a requirement that the catalog reach zero shared phrases.
- Expect some stable overlap between adjacent rules such as timeout families, restart families, and generic-versus-specialized build failures.
- Investigate new broad phrases, new duplicate IDs, or ranking regressions; do not churn mature rules only to drive the overlap count down.

Acceptance bar:

- one dominant playbook per root cause unless the detection boundary is genuinely different
- distinctive signals over broad wording
- short, ordered fixes tied to the root cause
- explicit negative signals when a nearby false positive is known
- shipped playbooks must be defendable against at least one confusable example

## Pack composition

Faultline ships the default catalog from `playbooks/bundled/` and can compose team-specific or extended packs on top of it.

Use this boundary when deciding where a playbook belongs:

- bundled: high-frequency failures across common stacks or CI systems, plus enough baseline source coverage for `inspect` to produce useful default results
- extra pack: provider-specific workflows, advanced deployment and platform operations, security-heavy rules, and deeper source-detector coverage beyond the default baseline

There are three supported ways to add extra packs:

1. `faultline packs install <dir>` to copy a pack into `~/.faultline/packs/` for automatic loading on future runs
2. repeat `--playbook-pack <dir>` for one-off or scripted composition
3. set `FAULTLINE_PLAYBOOK_PACKS` for environment-driven composition

Use `--playbooks <dir>` only for full catalog overrides such as testing a pack in isolation.

For installed extra packs, prefer `faultline packs install` as the long-lived path. It survives binary upgrades, avoids repeated flags, and works with the Docker image when `~/.faultline` is mounted into `/home/faultline/.faultline`.

## Minimal Example Pack

A small example pack lives under `examples/packs/minimal/`.

It is intentionally outside `playbooks/bundled/` so it does not affect the default catalog or fixture gates. Use it as a starting point for pack structure, field naming, and local validation:

```bash
./bin/faultline list --playbook-pack examples/packs/minimal
./bin/faultline explain example-cache-prime-missing --playbook-pack examples/packs/minimal
```

When you are ready to install a real pack persistently:

```bash
./bin/faultline packs install ./examples/packs/minimal --name example-pack
./bin/faultline packs list
```
