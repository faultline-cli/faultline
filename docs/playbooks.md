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

Optional constrained hook fields:

```yaml
hooks:
  verify:
    - id: package-lock-present
      kind: file_exists
      path: package-lock.json
      confidence_delta: 0.05
  collect:
    - id: npmrc-excerpt
      kind: read_file_excerpt
      path: .npmrc
      max_bytes: 200
```

Use inline hooks sparingly. They belong in the playbook when the extra
verification or evidence is part of the same rule definition and should travel
with the playbook itself.

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
- Keep hooks typed, explicit, and inspectable. Hooks are an additive evidence
  layer, not a second matcher.
- Prefer `verify` and `collect` hooks over `remediate`. Remediation hooks are
  schema-supported today but remain execution-blocked in the shipped CLI.
- Use `confidence_delta` only on `verify` hooks, and keep it small. Hook
  verification refines displayed confidence after matching; it does not create
  a match or rerank the catalog in the current implementation.

## Improvement pipeline

Treat playbook growth as a deterministic review loop, not a content-volume goal.

1. Ingest evidence from bundled playbooks, clean fixtures, noisy corpus logs, missed detections, false positives, and repository inspection findings.
2. Normalize each candidate into a root-cause record with likely category, distinctive signatures, confusable neighbors, and an actionable fix path.
3. Cluster by underlying failure mechanism, not by wording. Reject vague or duplicate clusters before authoring anything.
4. Prefer improving the strongest nearby playbook over adding a shallow variant. Add a new playbook only when the root cause is distinct and the signals are defensible.
5. For every accepted playbook, add at least one positive fixture and one nearby negative or adversarial regression so ranking stays stable in noisy logs.
6. Re-run `make review` after edits to inspect shared patterns and `make test` to confirm fixture, corpus, and ranking regressions remain deterministic.

When you are authoring from a new real failure, keep the maintainer workflow
explicit and deterministic:

1. ingest or stage the candidate evidence
2. sanitize the staging fixture
3. optionally generate a draft with `faultline fixtures scaffold`
4. review and hand-edit the YAML before committing anything
5. run `make review` and the relevant regression checks

`faultline fixtures scaffold` is hidden and maintainer-only on purpose. It is a
drafting helper, not an authoritative authoring engine.

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

Installed packs now carry deterministic provenance metadata recorded at install
time in `faultline-pack.yaml`. When available, Faultline preserves:

- semantic version
- install-time source URL or local source path
- pinned git ref

`faultline packs list` exposes that metadata through `VERSION` and
`PINNED REF` columns, and analysis JSON includes additive `pack_provenance`
entries so downstream automation can audit which catalog inputs were active.

## Hook Catalog Overlays

Team-specific hooks should usually live in a pack-level overlay file instead of
forking the shipped playbook YAML.

Create `faultline-hooks.yaml` at the root of the pack:

```yaml
schema_version: hooks.v1

named_hooks:
  repo.node-version:
    kind: command_output_matches
    command:
      - node
      - --version
    pattern: ^v20\.
    confidence_delta: 0.05

  repo.node-version-capture:
    extends: repo.node-version
    kind: command_output_capture

playbook_hooks:
  runtime-mismatch:
    verify:
      - use: repo.node-version
    collect:
      - id: nvmrc-excerpt
        kind: read_file_excerpt
        path: .nvmrc
        max_bytes: 80

  docker-auth:
    disable:
      - old-inline-hook
    verify:
      - id: docker-config
        kind: file_exists
        path: ~/.docker/config.json
        confidence_delta: 0.04
```

Deterministic merge rules:

- Pack order is the existing catalog order: bundled first, then extra packs.
- Named hooks are merged by ID. A later pack with the same name overrides the
  earlier definition.
- `playbook_hooks.<id>` attaches hooks to an existing playbook without
  redefining that playbook.
- `disable` removes previously attached hooks by ID for that playbook.
- Final hook references are resolved after pack composition, so teams can
  override a named hook in a later pack without editing every `use:` site.

This keeps hook customization inside the existing pack boundary, which means
teams can extend shipped playbooks without copying the original playbook files.

## Hook Execution Modes

Hooks never run unless the user enables them explicitly with the hidden
`--hooks` flag on `faultline analyze` or `faultline trace`.

Current modes:

- `off`: do not execute hooks
- `verify-only`: execute only non-command verify hooks
- `collect-only`: execute only non-command collect hooks
- `safe`: execute verify and collect hooks, but block command hooks
- `full`: execute verify and collect hooks including command hooks

Current safety rules:

- read-only typed hooks are the only supported primitive
- command hooks are blocked unless `--hooks full` is selected
- remediation hooks are reported as blocked in all modes in this release
- raw script hooks are not supported
- every hook outcome is surfaced in trace output and in additive JSON fields as
  `executed`, `blocked`, `skipped`, or `failed`

The supported typed hooks in the current implementation are:

- `file_exists`
- `dir_exists`
- `env_var_present`
- `command_exit_zero`
- `command_output_matches`
- `command_output_capture`
- `read_file_excerpt`

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

## Hidden Authoring Helper

Maintainers can also draft a candidate playbook from a sanitized log:

```bash
faultline fixtures scaffold --log build.log --category build
faultline fixtures scaffold --from-fixture <staging-id> --category auth
faultline fixtures scaffold --log build.log --category ci --pack-dir ./packs/team-pack
```

This helper:

- applies the same deterministic sanitizer pass before extracting patterns
- generates a candidate `match.any` block and required markdown fields with
  `TODO` markers
- writes to `<pack-dir>/<id>.yaml` only when `--pack-dir` is provided

The output still requires human review. Use it to accelerate drafting, not to
skip the normal fixture, review, and regression gates.
