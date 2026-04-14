# Faultline Distribution

Faultline is distributed as a public CLI with deterministic playbooks bundled into release artifacts.

This document covers how the binary, bundled playbooks, Docker image, and validation workflow are packaged for release.

## Public Release Contents

This repository ships:

- source code for the CLI
- bundled playbooks under `playbooks/bundled/`
- public release tarballs
- public Docker build instructions

Optional extra playbook packs can be composed on top when needed, but the default release should be useful on its own.

## Install Flow

Tagged releases publish tarballs named `faultline_<version>_<os>_<arch>.tar.gz` on the GitHub Releases page. The archive flow is:

```bash
VERSION=<published-tag>
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

If you move the binary elsewhere, keep `playbooks/bundled/` beside it or set `FAULTLINE_PLAYBOOK_DIR`.

For day-zero repository use before a tagged release is published, prefer `make build` or the Docker flow from the root README.

## Docker Distribution

The public Docker image ships the CLI plus the bundled playbooks. A typical run looks like this:

```bash
docker run --rm \
    -v "$(pwd)":/workspace \
    faultline analyze /workspace/build.log
```

Mounted host directories make it easy to analyze local logs without extra runtime dependencies.

## Release Artifacts

Public release tarballs should contain:

- the `faultline` binary
- `playbooks/bundled/`
- `README.md`

Archives are written to `dist/releases/<version>/` by `make release-snapshot VERSION=<tag>`.

## Release Workflow

Tagged releases should continue to run this sequence:

1. `make test`
2. `make review`
3. `make fixture-check`
4. `make release-snapshot VERSION=<tag>`
5. `make smoke-release VERSION=<tag>`
6. `make docker-smoke IMAGE=faultline-release-smoke`
7. publish release archives to the GitHub release created from that tag


