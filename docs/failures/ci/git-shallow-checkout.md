# Git shallow clone missing history or tags

**Playbook ID:** `git-shallow-checkout`
**Category:** ci
**Severity:** high
**Tags:** `git`, `ci`, `shallow`, `history`, `tags`, `versioning`, `github-actions`

## What this failure means

The CI job performed a shallow Git clone and a later step requires full commit history or tags. Tools like semantic-release, lerna, changelog generators, and tag-based version bumps fail because the expected refs are absent.

## Common log signals

```text
shallow update not allowed
fatal: shallow file has changed since we read it
git fetch --unshallow
unshallow
fetch-depth: 0
No names found, cannot describe anything
fatal: No tags can describe
You need to fetch more history
```

## Diagnosis

`actions/checkout` (and most CI checkout actions) default to `fetch-depth: 1`, which fetches only the most recent commit. Any step that needs to walk the commit graph, resolve a tag, or compute a version from git history will fail.

Common downstream failures:

- `git describe` returns an error or a wrong result because no tags are reachable.
- `semantic-release`, `standard-version`, or `lerna version` cannot determine the last release tag.
- Changelog generation tools produce empty output because no prior commits are visible.
- `git log` from the tool returns only a single commit.
- A `git fetch --tags` or `git unshallow` step is missing in the workflow.

## Fix steps

1. Set `fetch-depth: 0` in the `actions/checkout` step to fetch the complete history and all tags:

   ```yaml
   - uses: actions/checkout@v4
     with:
       fetch-depth: 0
   ```

2. If you only need tags (not the full history), fetch tags explicitly after a shallow checkout:

   ```bash
   git fetch --tags --force
   ```

3. For self-hosted runners or other CI systems, replace the shallow boundary with a full fetch:

   ```bash
   git fetch --unshallow
   ```

4. If the workflow runs in a detached HEAD state, confirm the branch name is also available to the versioning tool:

   ```bash
   git checkout -B "$BRANCH_NAME" HEAD
   ```

## Validation

- `git log --oneline | wc -l` returns more than 1.
- `git describe --tags` resolves without error.
- The versioning tool produces the expected tag or version string.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`
- `.gitlab-ci.yml`
- `.circleci/config.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain git-shallow-checkout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Git shallow clone missing history or tags
- Ci: git shallow clone missing history or tags
- fatal: shallow file has changed since we read it
- GitHub Actions git shallow clone missing history or tags
- faultline explain git-shallow-checkout


---

*Generated from [playbooks/bundled/log/ci/git-shallow-checkout.yaml](../../playbooks/bundled/log/ci/git-shallow-checkout.yaml). Do not edit directly — run `make docs-generate`.*
