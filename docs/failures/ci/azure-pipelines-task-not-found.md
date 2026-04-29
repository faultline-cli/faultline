# Azure Pipelines task version not found or deprecated

**Playbook ID:** `azure-pipelines-task-not-found`
**Category:** ci
**Severity:** medium
**Tags:** `azure-pipelines`, `azure-devops`, `task`, `version`, `deprecated`, `ci`

## What this failure means

An Azure Pipelines job could not start because a pipeline task references a version that has been deprecated, removed, or does not exist in the organization's task catalog. The pipeline is blocked before any work runs.

## Common log signals

```text
Task not found
task.*not found
Task version is not available
version .* is not available
deprecated task
Task '.*' is not recognized
Could not find task
the task is missing
```

## Diagnosis

Azure Pipelines tasks are versioned (e.g., `AzureCLI@2`, `DotNetCoreCLI@3`). When a referenced version is unavailable:

- The run fails with `Task '[name]' with version '...' not found` or `Task version is not available`.
- Microsoft deprecates old task major versions and eventually removes them.
- A self-hosted agent pool may not have the agent software that ships the required task version.
- A custom or marketplace task may have been removed from the organization's extension catalog.

## Fix steps

1. Update the task to the current major version in `azure-pipelines.yml`:

   ```yaml
   # Before (deprecated)
   - task: AzureCLI@1

   # After (current)
   - task: AzureCLI@2
   ```

2. Check the Azure Pipelines task reference documentation for the current supported version of the task.

3. For marketplace tasks, go to Organization Settings → Extensions and verify the extension is still installed and enabled.

4. For self-hosted agent pools, update the agent to a version that ships the required built-in task.

5. Review Microsoft's deprecation announcements for the task family — some tasks require migration to a new task name, not just a version bump.

## Validation

- The pipeline run starts and the previously failing task step begins execution.
- No `Task not found` or `version is not available` message appears.

## Likely files to inspect

- `azure-pipelines.yml`
- `.azuredevops/pipelines/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain azure-pipelines-task-not-found
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Azure Pipelines task version not found or deprecated
- Ci: azure pipelines task version not found or deprecated
- Task version is not available
- GitHub Actions azure pipelines task version not found or deprecated
- faultline explain azure-pipelines-task-not-found


---

*Generated from [playbooks/bundled/log/ci/azure-pipelines-task-not-found.yaml](../../playbooks/bundled/log/ci/azure-pipelines-task-not-found.yaml). Do not edit directly — run `make docs-generate`.*
