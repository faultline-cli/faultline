# Faultline Failure Pages

These pages turn recurring CI failures into searchable, problem-first docs.

Each page is designed to win searches for exact CI error strings, explain the failure quickly, and show how Faultline maps the log to a deterministic diagnosis.

## Included pages

- [pull access denied for ghcr.io image: fix Docker registry auth failures in CI](docker/pull-access-denied-ghcr-authentication-required.md)
- [Cannot connect to the Docker daemon at unix:///var/run/docker.sock](docker/cannot-connect-to-the-docker-daemon.md)
- [exec /__e/node20/bin/node: no such file or directory in GitHub Actions](github-actions/exec-node20-bin-node-no-such-file-or-directory.md)
- [npm ci can only install packages when package.json and package-lock.json are in sync](npm/npm-ci-package-lock-json-in-sync.md)
- [npm ERR! ERESOLVE unable to resolve dependency tree in CI](npm/npm-err-eresolve-unable-to-resolve-dependency-tree.md)
- [The engine "node" is incompatible with this module](node/the-engine-node-is-incompatible-with-this-module.md)
- [JavaScript heap out of memory in Node.js builds](node/javascript-heap-out-of-memory.md)
- [Your lockfile needs to be updated, but yarn was run with --frozen-lockfile](node/your-lockfile-needs-to-be-updated-yarn-frozen-lockfile.md)
- [ERR_PNPM_OUTDATED_LOCKFILE: frozen-lockfile failed in CI](pnpm/err-pnpm-outdated-lockfile.md)
- [no space left on device in CI builds](runtime/no-space-left-on-device.md)
- [poetry.lock is not consistent with pyproject.toml in CI](python/poetry-lock-is-not-consistent-with-pyproject-toml.md)
- [ModuleNotFoundError: No module named in Python CI jobs](python/module-not-found-error-no-module-named.md)
- [Host key verification failed during git clone in CI](git/host-key-verification-failed.md)
- [fatal: repository not found in GitHub Actions or GitLab CI](git/fatal-repository-not-found.md)
- [no such host and getaddrinfo ENOTFOUND in CI](network/no-such-host-getaddrinfo-enotfound.md)
- [ECONNREFUSED and connection refused in CI services](network/econnrefused-connection-refused.md)
- [context deadline exceeded and read timeout in CI](network/context-deadline-exceeded-read-timeout.md)
- [missing go.sum entry for module providing package](go/missing-go-sum-entry-for-module-providing-package.md)
- [x509: certificate signed by unknown authority in Docker or CI builds](tls/x509-certificate-signed-by-unknown-authority.md)
- [permission denied while trying to connect to the Docker daemon socket](docker/permission-denied-while-trying-to-connect-to-the-docker-daemon-socket.md)

## Recommended content model

Keep each page narrowly scoped to one dominant search intent:

1. Put the exact error string in the title and code block.
2. Keep the log snippet minimal and real.
3. Explain the failure in one short paragraph.
4. Give the smallest fix sequence that gets CI green again.
5. Show the Faultline playbook ID and primary matching signals.
6. Cross-link to adjacent failures instead of bloating the page.

## Recommended repo structure

```text
docs/
  failures/
    README.md
    _template.md
    docker/
    go/
    git/
    github-actions/
    network/
    node/
    npm/
    pnpm/
    python/
    runtime/
    tls/

examples/
  failures/
    docker/
    go/
    git/
    github-actions/
    network/
    node/
    npm/
    pnpm/
    python/
    runtime/
    tls/
```

Mirror the docs slug under `examples/failures/` when you add runnable sample logs and expected output snapshots.

## Authoring template

Use [docs/failures/_template.md](_template.md) for new pages.