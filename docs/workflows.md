# Workflows

`faultline workflow` is the deterministic handoff surface for the winning
diagnosis. It analyzes a log, selects the top playbook, and renders the
playbook's authored follow-up metadata as either human text or stable JSON.

It does not execute remediation steps. It is meant to give a human, agent,
ticket, or post-mortem tool a clear next-action bundle without adding another
automation engine to the core product.

## CLI

```bash
faultline workflow <logfile>
cat build.log | faultline workflow
faultline workflow <logfile> --json --mode agent
```

The JSON shape uses `schema_version: workflow.v1` and includes:

- `failure_id`, `title`, `status`, and source context
- matched evidence lines
- likely files to inspect
- local reproduction commands from `workflow.local_repro`
- verification commands from `workflow.verify`
- deterministic fix steps from the matched playbook
- an `agent_prompt` in `--mode agent`
- the first-class failure artifact and remediation handoff

## Playbook Metadata

Playbooks describe workflow handoff data with the existing `workflow:` block:

```yaml
workflow:
  likely_files:
    - Dockerfile
    - .github/workflows/*.yml
  local_repro:
    - command -v <tool>
  verify:
    - command -v <tool>
```

Keep these fields concrete and local:

- `likely_files` should point to the files a maintainer or agent should inspect
  first.
- `local_repro` should contain commands that reproduce or confirm the failure
  condition before editing.
- `verify` should contain commands that prove the diagnosis-specific failure is
  gone after the fix.

## History

Workflow history enrichment is opt-in, just like analysis history. Use one of:

```bash
faultline workflow build.log --history
faultline workflow build.log --store auto
FAULTLINE_STORE=~/.faultline/store.db faultline workflow build.log
```

Without one of those opt-ins, repeated default workflow output stays stable and
uses the no-op store.

## Non-Goals

- executing shell commands
- package installation or environment mutation
- workflow schema registries
- persisted remediation execution records
- hidden LLM planning
- rollback orchestration
- remote control plane features

The workflow surface should stay deterministic, local-first, explicit, and
easy to audit.
