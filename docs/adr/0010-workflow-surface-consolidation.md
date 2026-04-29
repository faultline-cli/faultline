# ADR 0010: Workflow Surface Consolidation

- Status: Accepted
- Date: 2026-04-30

## Context

Faultline's core product story is deterministic diagnosis plus a clear
follow-up handoff. Workflow output should support that story without becoming a
second automation engine inside the CLI.

## Decision

`faultline workflow` is the workflow surface. It emits a deterministic handoff
artifact from the matched playbook's `workflow:` metadata, diagnosis,
remediation plan, and artifact.

The workflow surface does not execute remediation steps or persist remediation
execution records. Local store state remains scoped to analysis, signatures,
playbook matches, artifacts, and hook history.

## Consequences

- `internal/workflow/workflow.go` owns workflow handoff generation.
- Playbooks should keep using `workflow.likely_files`, `workflow.local_repro`,
  and `workflow.verify` for deterministic handoff data.
- Faultline core does not execute remediation workflows or persist remediation
  execution history.
- Future workflow work should start hidden or experimental unless it has
  deterministic tests, release-grade verification, and docs that match shipped
  behavior.

## References

- `internal/workflow/workflow.go` - workflow handoff generation
- `internal/app/service.go` - `Service.Workflow` call site
- `docs/workflows.md` - workflow surface contract
