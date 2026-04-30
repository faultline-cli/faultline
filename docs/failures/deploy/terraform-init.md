# Terraform initialization failure

**Playbook ID:** `terraform-init`
**Category:** deploy
**Severity:** high
**Tags:** `terraform`, `iac`, `init`, `provider`, `backend`

## What this failure means

`terraform init` failed before any plan or apply could begin. Initialization downloads providers and configures the backend, so errors here block every later Terraform step.

## Common log signals

```text
Error: Failed to install provider
Error: Unsatisfied provider requirements
Error: Could not load plugin
Could not retrieve the list of available versions
Error: Backend configuration changed
Error: No configuration files
Failed to query available provider packages
Could not resolve provider
```

## Diagnosis

`terraform init` failed before any plan or apply could begin. Initialization downloads providers and configures the backend, so errors here block every later Terraform step.

## Fix steps

1. Check the provider version constraints in `required_providers` against what is available at `registry.terraform.io`.
2. For backend failures, verify that the state bucket or container exists and the CI service account has read and write permission on it.
3. Run `terraform init -upgrade` locally to refresh the provider lock file.
4. Check that the Terraform version in CI satisfies the `required_version` constraint in your Terraform configuration.

## Validation

- `terraform init` completes successfully in the same workspace.
- `terraform version` matches the expected `required_version` contract.

## Likely files to inspect

- `*.tf`
- `.terraform.lock.hcl`
- `terraform/**`


## Run Faultline

```bash
faultline analyze build.log
faultline explain terraform-init
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Terraform initialization failure
- Deploy: terraform initialization failure
- Could not retrieve the list of available versions
- faultline explain terraform-init
- Terraform terraform initialization failure


---

*Generated from [playbooks/bundled/log/deploy/terraform-init.yaml](../../../playbooks/bundled/log/deploy/terraform-init.yaml). Do not edit directly — run `make docs-generate`.*
