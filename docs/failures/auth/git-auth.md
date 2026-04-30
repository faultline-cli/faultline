# Git authentication failure

**Playbook ID:** `git-auth`
**Category:** auth
**Severity:** high
**Tags:** `git`, `github`, `ssh`, `auth`

## What this failure means

The CI runner could not authenticate with the remote Git repository. This prevents cloning, fetching, or pushing code.

## Common log signals

```text
terminal prompts disabled
could not read username
authentication failed for
repository not found
remote: invalid username or password
fatal: credential helper
exit code 128
```

## Diagnosis

The CI runner could not authenticate with the remote Git repository. This prevents cloning, fetching, or pushing code.

## Fix steps

1. For HTTPS, prefer the CI provider's checkout or credential helper instead
   of embedding a token in a global URL rewrite. In GitHub Actions, use the
   `actions/checkout` step with `token: ${{ secrets.GITHUB_TOKEN }}` so the
   credential is injected for the job only.

2. For SSH: add the deploy key or SSH key as a CI secret, then load it:

   ```bash
   eval "$(ssh-agent -s)"
   echo "$SSH_KEY" | ssh-add -
   ```

3. Verify the token has the required scope:
   - GitHub PAT: needs `repo` (private repos) or `read:packages`.
   - Deploy key: needs read (or read+write for push) on the specific repo.

4. Confirm the remote URL scheme matches the credentials you have:

   ```bash
   git remote -v
   ```

5. Test the credential before the full workflow step:

   ```bash
   git ls-remote origin HEAD
   ```

## Validation

- `git ls-remote origin HEAD` exits 0 and prints the current HEAD SHA.
- Re-run the failing clone, fetch, or push step and confirm no authentication
  error appears.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitmodules`
- `.ssh/config`


## Run Faultline

```bash
faultline analyze build.log
faultline explain git-auth
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Git authentication failure
- Auth: git authentication failure
- remote: invalid username or password
- faultline explain git-auth


---

*Generated from [playbooks/bundled/log/auth/git-auth.yaml](../../../playbooks/bundled/log/auth/git-auth.yaml). Do not edit directly — run `make docs-generate`.*
