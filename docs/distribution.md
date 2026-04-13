# Faultline Distribution

Faultline v1 is easiest to ship as two separate products:

1. a public starter CLI repository with public releases
2. a private premium playbook repository or private premium release archive

That keeps the core binary easy to evaluate while preserving a clean paid upgrade path for additional playbooks.

For local validation, this repo can point at a sibling premium repository
through the ignored symlink at `playbooks/packs/premium-local` or by setting
`PREMIUM_PACK_DIR`. Customer delivery should still publish the pack from a
private repository or private archive.

## Product Split

Public repository responsibilities:

- source code for the CLI
- bundled starter playbooks under `playbooks/bundled/`
- public release tarballs
- public Docker build instructions

Private premium repository responsibilities:

- premium playbook pack only
- optional private release archive for buyers who should not clone the repo
- no fork of the CLI codebase unless there is a separate need

The premium repository should be a pack root that Faultline can load recursively.

## Buyer Onboarding

The public starter install remains unchanged:

```bash
curl -L <public-release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

Premium onboarding should add one explicit install step:

```bash
# Option 1: clone the private premium repository
git clone <private-premium-pack-repo> ../faultline-premium

# Option 2: unpack a private premium release archive
tar -xzf faultline-premium.tar.gz -C ..

# Install the premium pack once
./faultline packs install ../faultline-premium

# Verify the install
./faultline packs list
./faultline list
```

`faultline packs install` copies the premium pack into `~/.faultline/packs/<pack-name>`. That gives buyers a stable local upgrade path even when they replace the Faultline binary with a newer public release.

## Premium Upgrades

Premium updates should not require a different CLI build. Buyers update the private pack, then reinstall it in place:

```bash
cd ../faultline-premium && git pull
./faultline packs install --force ../faultline-premium
```

For archive-based delivery, the same pattern applies after unpacking the newer premium archive.

## Verification Checklist

For the initial manual sales flow:

1. Confirm payment.
2. Grant access to the private premium repository or send the premium archive.
3. Ask the buyer to run `faultline packs install <premium-pack-dir>`.
4. Ask the buyer to run `faultline packs list` and confirm the installed pack is shown.
5. Ask the buyer to run `faultline list` and confirm premium playbooks appear in the `PACK` column.

This keeps the upgrade path deterministic and supportable without a hosted entitlement system.

## Docker Distribution

The public Docker image should continue to ship only the starter catalog.

Premium access in Docker should use the same installed-pack location as the host machine:

```bash
docker run --rm \
    -v "$HOME/.faultline":/home/faultline/.faultline \
    -v "$(pwd)":/workspace \
    faultline analyze /workspace/build.log
```

That means one premium install can serve both local CLI runs and containerized runs.

For teams that need premium packs baked into a custom image, create a thin derived image on top of the public base:

```dockerfile
FROM faultline:latest
COPY faultline-premium /home/faultline/.faultline/packs/faultline-premium
```

Use that only for internal delivery or CI images. Keep the public image starter-only.

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

When the private premium repository is available in CI or checked out locally, extend the gate:

```bash
make premium-check PREMIUM_PACK_DIR=../faultline-premium
make premium-review PREMIUM_PACK_DIR=../faultline-premium
make release-check VERSION=v0.1.0 PREMIUM_PACK_DIR=../faultline-premium
```

This catches duplicate IDs and cross-pack load errors before release.

## Later Automation

Only automate premium access after the manual flow has been exercised enough to justify it.

The first reasonable automation step is still:

1. payment webhook or marketplace event
2. small access grant workflow
3. private repo invitation or private archive fulfillment

Avoid license keys or a hosted auth service until the manual repo-or-archive flow becomes operationally expensive.
