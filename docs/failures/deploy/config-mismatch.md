# Deployment configuration mismatch

**Playbook ID:** `config-mismatch`
**Category:** deploy
**Severity:** medium
**Tags:** `config`, `env`, `secret`, `missing-variable`, `deployment`

## What this failure means

The deployment failed because a required configuration value, environment
variable, or secret was absent or had an unexpected value for the target
environment.

## Common log signals

```text
missing configuration
config not found
no such secret
secret not found
configmap not found
invalid configuration
cannot find key
```

## Diagnosis

The deployment failed because a required configuration value, environment
variable, or secret was absent or had an unexpected value for the target
environment.

This usually happens when a deployment references a secret or ConfigMap that
was never provisioned, or when a key name changed in code without a matching
change to the environment definition.

## Fix steps

1. Compare expected configuration keys against what is present in the target
   environment. For Kubernetes:

   ```bash
   kubectl get secret <name> -n <namespace> -o jsonpath="{.data}" | jq "keys"
   kubectl get configmap <name> -n <namespace> -o yaml
   ```

2. List all environment variables the pod sees to confirm keys are mounted:

   ```bash
   kubectl exec <pod> -- env | sort
   ```

3. Add or update the missing keys in the secret or ConfigMap, then
   re-trigger the deployment.
4. If a key was renamed, update both the secret store and the deployment spec
   in the same commit to avoid drift.
5. Add a start-of-process config validation step in the application that
   lists every missing required key at boot time rather than failing later
   with a cryptic error.

## Validation

- `kubectl exec <pod> -- env | grep KEY_NAME` shows the expected value.
- Re-run the deployment and confirm it reaches the Running / Ready state.

## Likely files to inspect

- `k8s/**/*.yaml`
- `deploy/**/*.yaml`
- `helm/**/*.yaml`
- `.env.example`


## Run Faultline

```bash
faultline analyze build.log
faultline explain config-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Deployment configuration mismatch
- Deploy: deployment configuration mismatch
- missing configuration
- faultline explain config-mismatch


---

*Generated from [playbooks/bundled/log/deploy/config-mismatch.yaml](../../../playbooks/bundled/log/deploy/config-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
