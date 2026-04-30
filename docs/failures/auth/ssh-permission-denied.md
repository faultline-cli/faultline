# SSH permission denied (publickey) authentication failed

**Playbook ID:** `ssh-permission-denied`
**Category:** auth
**Severity:** high
**Tags:** `ssh`, `git`, `auth`, `key`, `permission`, `publickey`

## What this failure means

SSH authentication to a Git host or remote server failed with explicit "Permission denied (publickey)" message. The private key is missing, not readable, or the public key is not registered on the server.

## Common log signals

```text
Permission denied (publickey)
git@.*Permission denied
```

## Diagnosis

SSH permission errors occur when:

- The SSH private key file is missing or not in the expected location.
- The private key has incorrect file permissions (too permissive for the key itself or parent directories).
- The public key is not added to the SSH server's `authorized_keys`.
- SSH agent is not running or the key was not added with `ssh-add`.
- The SSH host key (known_hosts) is not recognized.
- The CI environment has no SSH support configured.

The error explicitly shows `Permission denied (publickey)`, indicating the public key authentication path was attempted but rejected.

## Fix steps

1. Confirm the SSH private key file exists and has correct permissions:

   ```bash
   ls -la ~/.ssh/id_rsa
   chmod 600 ~/.ssh/id_rsa
   chmod 700 ~/.ssh
   ```

2. If using GitHub or similar, verify the public key is registered:

   ```bash
   ssh -T git@github.com 2>&1 | head -5
   ```

3. If running in CI (GitHub Actions, GitLab CI, etc.), configure SSH with the correct secret:

   **GitHub Actions example:**
   ```yaml
   - uses: webfactory/ssh-agent@v0.7.0
     with:
       ssh-private-key: ${{ secrets.SSH_PRIVATE_KEY }}
   ```

4. If the SSH host key is not recognized, add it with `ssh-keyscan`:

   ```bash
   ssh-keyscan github.com >> ~/.ssh/known_hosts
   ```

5. Test the SSH connection:

   ```bash
   ssh -v git@github.com
   ```

## Validation

- `ssh -T git@github.com` returns `Hi <username>! You've successfully authenticated...` (or similar success message).
- `ls -la ~/.ssh/id_rsa` shows `600` permissions on the key.
- Re-run the CI job.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `Dockerfile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ssh-permission-denied
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- SSH permission denied (publickey) authentication failed
- Auth: ssh permission denied (publickey) authentication failed
- Permission denied (publickey)
- faultline explain ssh-permission-denied


---

*Generated from [playbooks/bundled/log/auth/ssh-permission-denied.yaml](../../../playbooks/bundled/log/auth/ssh-permission-denied.yaml). Do not edit directly — run `make docs-generate`.*
