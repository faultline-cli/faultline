# Host key verification failed during git clone in CI

**Target search query:** `Host key verification failed git clone CI`

## Error snippet

```text
Cloning into 'repo'...
Host key verification failed.
fatal: Could not read from remote repository.
```

## What this error means

The job tried to clone over SSH, but the remote host fingerprint was not trusted or the SSH setup was incomplete.

## Fix steps

1. Add the host to known hosts with `ssh-keyscan github.com >> ~/.ssh/known_hosts`.
2. Load the private key into the agent or write it to the expected key path.
3. Fix private key permissions with `chmod 600`.
4. Re-run `ssh -vT git@github.com` before the clone step.

## How Faultline detects it

Faultline maps this failure to `ssh-key-auth`.

Primary signals:

- `Host key verification failed`
- `fatal: Could not read from remote repository`
- SSH handshake context

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [fatal: repository not found in GitHub Actions or GitLab CI](fatal-repository-not-found.md)
- [exec /__e/node20/bin/node: no such file or directory in GitHub Actions](../github-actions/exec-node20-bin-node-no-such-file-or-directory.md)