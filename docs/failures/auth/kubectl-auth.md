# Kubernetes cluster authentication failure

**Playbook ID:** `kubectl-auth`
**Category:** auth
**Severity:** high
**Tags:** `kubernetes`, `kubectl`, `auth`, `kubeconfig`

## What this failure means

`kubectl` could not authenticate with the Kubernetes API server. This prevents cluster operations such as `apply`, `rollout`, or `get`.

## Common log signals

```text
error: you must be logged in
You must be logged in to the server
unable to connect to the server
no configuration has been provided
error loading config file
context not found
kubectl cluster-info
```

## Diagnosis

`kubectl` could not authenticate with the Kubernetes API server. This prevents cluster operations such as `apply`, `rollout`, or `get`.

## Fix steps

1. Decode and export the kubeconfig from the CI secret: `echo "$KUBECONFIG_B64" | base64 -d > ~/.kube/config && chmod 600 ~/.kube/config`.
2. Verify the active context and cluster endpoint with `kubectl config current-context && kubectl config view --minify`.
3. Test connectivity before deploy steps with `kubectl cluster-info`.
4. For expired certificate credentials, regenerate the kubeconfig from the cloud provider CLI.
5. For service account token auth, confirm the token is still valid and attached to the job context you expect.
6. For RBAC errors immediately after auth succeeds, switch to the `k8s-rbac-forbidden` diagnosis instead.

## Validation

- `kubectl cluster-info` succeeds.
- `kubectl auth whoami` or an equivalent access check confirms the expected identity.
- `kubectl get pods -n <namespace> --request-timeout=10s` returns successfully.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `kubeconfig`
- `k8s/**/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain kubectl-auth
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Kubernetes cluster authentication failure
- Auth: kubernetes cluster authentication failure
- You must be logged in to the server
- faultline explain kubectl-auth
- Kubernetes kubernetes cluster authentication failure


---

*Generated from [playbooks/bundled/log/auth/kubectl-auth.yaml](../../../playbooks/bundled/log/auth/kubectl-auth.yaml). Do not edit directly — run `make docs-generate`.*
