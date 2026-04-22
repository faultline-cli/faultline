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

The default adoption path is the local CLI: diagnose the failing log first, then generate the deterministic workflow handoff.

```bash
curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
faultline workflow ci.log --json --mode agent
```

If you need a pinned version instead of the latest release:

```bash
VERSION=v0.3.1 curl -fsSL https://raw.githubusercontent.com/faultline-cli/faultline/main/install.sh | sh
faultline analyze ci.log
```

GitHub Actions is the strongest follow-up path when you want the same CLI artifacts attached automatically to failed workflow runs:

```bash
faultline analyze build.log --format markdown --ci-annotations > faultline-summary.md
faultline analyze build.log --json --bayes > faultline-analysis.json
faultline workflow build.log --json --mode agent > faultline-workflow.json
```

Publish the markdown summary and upload the JSON outputs as artifacts in the failing job.

If you are working from the repository directly, install from source:

```bash
git clone https://github.com/faultline-cli/faultline
cd faultline
go build -o faultline ./cmd
./faultline analyze examples/docker-auth.log
```

Release archives are published as `faultline_<version>_<os>_<arch>.tar.gz` on the GitHub Releases page. The archive flow is:

```bash
VERSION=v0.3.1
curl -fL "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze build.log
```

The installer places bundled playbooks under the install prefix and configures `FAULTLINE_PLAYBOOK_DIR` for the wrapper it places on `PATH`.

For provider-specific wrapper contracts, see `docs/github-action-contract.md` and `docs/gitlab-ci-contract.md`.

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

The current release snapshot script builds archives for:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`

## Release Workflow

Tagged releases should continue to run this sequence:

1. `make release-check VERSION=<tag>`
2. `WITH_DOCKER=1 IMAGE=faultline-release-smoke make release-check VERSION=<tag>` when the container contract changed
3. publish release archives to the GitHub release created from that tag

`make release-check` already includes `make test`, `make fixture-check`, `make review`, `make release-snapshot`, and `make smoke-release`.
