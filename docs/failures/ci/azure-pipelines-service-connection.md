# Azure Pipelines service connection failure

**Playbook ID:** `azure-pipelines-service-connection`
**Category:** ci
**Severity:** high
**Tags:** `azure-pipelines`, `azure-devops`, `service-connection`, `auth`, `ci`

## What this failure means

An Azure Pipelines job failed because the service connection required by a task does not exist, is not authorized, or has expired credentials. The task cannot authenticate to the external service (Azure Resource Manager, GitHub, Docker Hub, etc.).

## Common log signals

```text
service connection.*not found
Service connection .* does not exist
could not be found or does not exist
This service connection is not authorized
is not authorized for use
pipeline is not authorized to use this resource
service endpoint
Service Endpoint .* not found
```

## Diagnosis

Azure Pipelines uses service connections to store credentials for external integrations. Tasks such as `AzureRmWebAppDeployment`, `Docker`, `AzureCLI`, and `Kubernetes` reference a named service connection. When the connection is missing, unauthorized, or misconfigured:

- The task log shows `Service connection '[name]' could not be found` or `This service connection is not authorized`.
- An ARM service principal may have expired, been deleted, or had its password rotated without updating the connection.
- The pipeline is not granted permission to use the service connection in Azure DevOps Project Settings → Service connections.
- The pipeline was migrated across projects or organizations and the service connection reference is stale.

## Fix steps

1. In Azure DevOps, go to Project Settings → Service connections and verify the connection named in the pipeline task exists and is not expired.

2. For ARM connections: check the app registration in Azure Entra ID (formerly AAD) and confirm the service principal client secret has not expired.

3. Authorize the pipeline to use the service connection:
   - Open the service connection → Security → Pipeline permissions → Add the affected pipeline.

4. For expired secrets, rotate the client secret and update the service connection credentials.

5. For cross-project usage: grant explicit cross-project permissions on the service connection.

6. Verify the service endpoint URL and authentication settings match the current target environment.

## Validation

- Manually test the service connection via Project Settings → Service connections → [connection] → Verify connection.
- Re-run the pipeline and confirm the failing task authenticates without a service connection error.

## Likely files to inspect

- `azure-pipelines.yml`
- `.azuredevops/pipelines/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain azure-pipelines-service-connection
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Azure Pipelines service connection failure
- Ci: azure pipelines service connection failure
- pipeline is not authorized to use this resource
- GitHub Actions azure pipelines service connection failure
- faultline explain azure-pipelines-service-connection


---

*Generated from [playbooks/bundled/log/ci/azure-pipelines-service-connection.yaml](../../playbooks/bundled/log/ci/azure-pipelines-service-connection.yaml). Do not edit directly — run `make docs-generate`.*
