# ADR 0010: Legacy Workflow Migration Path

- Status: Accepted
- Date: 2026-04-27

## Context

Two workflow implementations coexist in `internal/workflow/`:

| File | Entry point | Output type | Used by |
|------|------------|-------------|---------|
| `legacy.go` | `BuildWithOptions` | `Plan` (steps `[]string`) | `faultline workflow` default command |
| `workflow.go` | `Explain`, `DryRun`, `Apply` | `PlanDocument` (steps `[]PlanStep`) | `faultline workflow explain/apply/show/history` |

Both emit JSON with `"schema_version": "workflow.v1"`. Because the schema version string is identical, a consumer reading a stored JSON record cannot determine which system produced it. The two output shapes are structurally incompatible: `steps` is an unstructured `[]string` in the legacy `Plan` and a structured `[]PlanStep` (with `phase`, `id`, `type`, `safety_class`, `idempotence`, and `args`) in `PlanDocument`.

The legacy `Plan` struct also carries several fields (`local_repro`, `verify`, `ranking_hints`, `delta_hints`, `metrics_hints`, `policy_hints`, `agent_prompt`) that have no counterpart in `PlanDocument`. The current `PlanDocument` adds `workflow_id`, `definition_schema_version`, `resolved_inputs`, `required_safety`, and `policy_notes` that are absent from `Plan`.

Allowing the two schemas to diverge further while sharing the same version string is the primary risk. The call site in `app/service.go` already carries a comment reading "emits the **legacy** deterministic follow-up workflow", which confirms the intent to freeze the path but does not enforce it structurally.

## Decision

1. **Freeze the legacy path.** No new fields, options, or code paths are added to `legacy.go` or the `Plan` struct. The file gains a `[LEGACY]` header block that documents its frozen status and the removal criteria.

2. **All new workflow features land in `workflow.go`** and the `execute`, `plan`, `bind`, `schema`, and `registry` sub-packages.

3. **Removal is gated on eval corpus verification.** The legacy path is removed only after a full eval-corpus run (`make eval`) confirms zero regression against the stored expected outputs for the `faultline workflow` command. Until that gate passes, the legacy path must remain because it is the only path that produces the text output format consumed by existing integrations and snapshot tests.

4. **Schema version strings stay as-is.** Introducing a `workflow.v2` or a `legacy` suffix before the eval gate passes would create a public breaking change with no benefit. Once the migration is complete and the legacy path is removed, the schema version for `PlanDocument` records is bumped in a single PR that also updates all stored fixtures.

5. **`app/service.go` comment is authoritative.** The `Service.Workflow` method is the only external-facing entry point for the legacy path. Its existing comment ("emits the legacy deterministic follow-up workflow") acts as the boundary marker at the API layer.

## Consequences

- Contributors must not add features or fix non-critical bugs in `legacy.go`. Any feature request that would require touching the legacy path is instead implemented in the current system's sub-commands.
- The `faultline workflow` default command (no sub-command) continues to work unchanged for existing integrations until the eval gate passes.
- The eval corpus gate (`make eval` producing no regressions on `workflow` expected outputs) is the explicit, testable removal criterion. There is no time-based deadline.
- Once removal lands, the `Plan` type, `BuildWithOptions`, `Build`, `schemaVersion`, `Mode`, `ModeLocal`, `ModeAgent`, and `BuildOptions` exported symbols are removed. Callers outside `internal/app` will need to migrate to the `PlanDocument` path.
- Any future schema version bump for `PlanDocument` happens in a single PR that also updates all stored fixtures and eval-corpus expected outputs.

## References

- `internal/workflow/legacy.go` — frozen legacy path with `[LEGACY]` header
- `internal/workflow/workflow.go` — current system entry points (`Explain`, `DryRun`, `Apply`)
- `internal/app/service.go` — `Service.Workflow` call site; boundary comment
- `internal/workflow/schema/schema.go` — `const Version = "workflow.v1"`
- ROADMAP.md TD-3 — original defect description
