# fatal: repository not found in GitHub Actions or GitLab CI

**Target search query:** `fatal: repository not found CI`

## Error snippet

```text
fatal: repository 'https://github.com/org/private-repo/' not found
```

## What this error means

The job reached the Git host, but the repository URL was wrong or the token did not have access to the target repository.

## Fix steps

1. Verify the repository URL in the checkout step or remote configuration.
2. Confirm the CI token or deploy credential can access the target repository.
3. If the repository is private, use the correct checkout token or deploy key.
4. Re-run `git ls-remote origin HEAD` with the same credential before the full workflow.

## How Faultline detects it

Faultline maps this failure to `git-auth`.

Primary signals:

- `repository not found`
- `could not read username`
- authentication failure during clone or fetch

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [Host key verification failed during git clone in CI](host-key-verification-failed.md)
- [pull access denied for ghcr.io image: fix Docker registry auth failures in CI](../docker/pull-access-denied-ghcr-authentication-required.md)