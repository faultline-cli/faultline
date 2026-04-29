# Git refspec or branch does not exist

**Playbook ID:** `git-refspec-mismatch`
**Category:** build
**Severity:** high
**Tags:** `git`, `checkout`, `refspec`, `vcs`, `branch-not-found`

## What this failure means

A git checkout or clone operation failed because the requested refspec (branch name, tag, or commit reference) does not exist on the remote repository.

## Common log signals

```text
couldn't find remote ref
invalid refspec
repository.*does not exist
remote ref.*does not exist
branch.*does not exist
ref.*does not match any
failed to fetch submodule
re:Remote branch .* not found in upstream
```

## Diagnosis

When cloning or checking out a Git repository, the specified ref (branch, tag, or PR reference) must exist on the remote. If the ref is missing or misnamed, Git fails with "couldn't find remote ref" or "invalid refspec" error.

This is distinct from network errors — the remote is reachable, but the specific reference doesn't point to anything.

Common causes:
- Branch name was deleted or renamed
- Typo in branch name (e.g., `feature/auth` instead of `feature/auth-v2`)
- PR reference is stale (PR was closed or rebased)
- Shallow clone with `--depth 1 --branch <name>` where branch doesn't exist
- Incorrect refspec syntax (e.g., `feature/name:feature/name` with wrong colons)
- Branch exists locally but not on origin

## Fix steps

1. Verify the exact branch name on the remote:

   ```bash
   git ls-remote https://github.com/myorg/myrepo.git | head -20
   ```

2. Check if the branch exists:

   ```bash
   git ls-remote --heads https://github.com/myorg/myrepo.git feature/auth
   ```

3. If using a shallow clone, try a full clone first to confirm the branch truly exists:

   ```bash
   git clone https://github.com/myorg/myrepo.git . --no-checkout
   git fetch origin <correct-branch-name>:refs/heads/<correct-branch-name>
   ```

4. Correct the branch name in your CI configuration or Git command.

5. For PR references, ensure the PR still exists and hasn't been closed:

   ```bash
   git fetch origin pull/<PR_NUMBER>/head
   ```

## Validation

- Run `git ls-remote` to confirm the correct ref exists on the remote.
- Re-run the checkout with the correct branch name.
- Confirm the repository is cloned and checked out successfully.

## Likely files to inspect

- `.github/workflows/*.y*ml`
- `.gitlab-ci.yml`
- `.circleci/config.yml`
- `Makefile`
- `scripts/clone.sh`


## Run Faultline

```bash
faultline analyze build.log
faultline explain git-refspec-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Git refspec or branch does not exist
- Build: git refspec or branch does not exist
- re:Remote branch .* not found in upstream
- GitHub Actions git refspec or branch does not exist
- faultline explain git-refspec-mismatch


---

*Generated from [playbooks/bundled/log/build/git-refspec-mismatch.yaml](../../playbooks/bundled/log/build/git-refspec-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
