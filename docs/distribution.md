# Faultline Distribution

Faultline is distributed as a public CLI with deterministic playbooks bundled into release artifacts.

This document covers how the public binary, playbooks, Docker image, and validation workflow are packaged for launch.

## Public Release Contents

This repository ships:

- source code for the CLI
- bundled playbooks under `playbooks/bundled/`
- public release tarballs
- public Docker build instructions

Extra playbook packs can be composed on top when needed, but the default release should be useful on its own.

## Install Flow

The release archive flow is:

```bash
curl -L <public-release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

If you move the binary elsewhere, keep `playbooks/bundled/` beside it or set `FAULTLINE_PLAYBOOK_DIR`.

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
3. `make release-snapshot VERSION=<tag>`
4. `make smoke-release VERSION=<tag>`
5. `make docker-smoke IMAGE=faultline-release-smoke`
6. publish release archives


Internal validation and future pack-delivery notes are parked in `docs/monetisation.md` so the public launch path stays focused.
