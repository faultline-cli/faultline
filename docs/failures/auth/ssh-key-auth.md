# SSH key authentication failure

**Playbook ID:** `ssh-key-auth`
**Category:** auth
**Severity:** high
**Tags:** `ssh`, `key`, `git`, `auth`, `deploy-key`

## What this failure means

An SSH operation (git clone, deploy, rsync) failed because the SSH key was not found, not loaded into the agent, or the remote host does not accept it.

## Common log signals

```text
Host key verification failed
no such identity
Authentication failed
ssh: handshake failed
ssh: connect to host
Warning: Permanently added
Pseudo-terminal will not be allocated
```

## Diagnosis

An SSH operation (git clone, deploy, rsync) failed because the SSH key was not found, not loaded into the agent, or the remote host does not accept it.

## Fix steps

1. Add the private key to CI secrets and load it: `ssh-add <(echo "$SSH_PRIVATE_KEY")`.
2. Fix key permissions: `chmod 600 ~/.ssh/id_rsa`.
3. Add the host to known hosts: `ssh-keyscan github.com >> ~/.ssh/known_hosts`.
4. Verify the corresponding public key is added as a deploy key on the repository.
5. Use `ssh -vT git@github.com` to debug the handshake.

## Validation

- Re-run the local reproduction command after the fix.
- ssh -vT git@github.com

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ssh-key-auth
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- SSH key authentication failure
- Auth: ssh key authentication failure
- Pseudo-terminal will not be allocated
- faultline explain ssh-key-auth


---

*Generated from [playbooks/bundled/log/auth/ssh-key-auth.yaml](../../playbooks/bundled/log/auth/ssh-key-auth.yaml). Do not edit directly — run `make docs-generate`.*
