# Faultline Distribution

This repository is set up for a simple commercial distribution model:

1. Sell access on Gumroad.
2. Collect the buyer's GitHub username.
3. Grant access to the private `faultline-cli` repository.
4. Distribute versioned release tarballs from GitHub Releases.

The goal is to keep delivery low-friction without adding a hosted auth system.

## Recommended private repo layout

Keep the repository itself as the product surface:

```text
.
├── cmd/
├── internal/
├── playbooks/
├── docs/
│   └── distribution.md
├── .github/
│   └── workflows/
│       └── release.yml
└── dist/
    └── releases/
        └── v1.2.3/
            ├── faultline_v1.2.3_darwin_amd64.tar.gz
            └── faultline_v1.2.3_linux_amd64.tar.gz
```

Release tarballs include:

- the `faultline` binary
- the bundled `playbooks/` directory
- the repository `README.md`

Bundling `playbooks/` is required because the CLI loads playbooks from disk at
runtime.

## Buyer delivery flow

Use Gumroad for payment and GitHub for access control and updates.

Suggested post-purchase message:

```text
Thanks for purchasing Faultline.

To get access:

1. Reply with your GitHub username or submit it here: <intake link>
2. You'll be granted access to the private repository shortly
3. Open the latest GitHub Release and download the archive for your OS
```

Suggested private repo onboarding snippet:

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

## Manual access checklist

For the first sales, keep the process manual:

1. Confirm payment in Gumroad.
2. Add the buyer to the private GitHub repository.
3. Confirm they can see Releases.
4. Send the onboarding snippet above.

This keeps the system deterministic and avoids premature infrastructure work.

## Release workflow

Tagged releases are built by [release.yml](/home/jake/workspace/faultline/.github/workflows/release.yml).

The workflow:

1. runs `make test`
2. builds release tarballs with `make release-snapshot VERSION=<tag>`
3. uploads the archives as workflow artifacts
4. publishes them to the GitHub Release for tag pushes

To build the same artifacts locally:

```bash
make release-snapshot VERSION=v0.1.0
```

Archives are written to `dist/releases/<version>/`.

## Later upgrade path

Only automate access after the paid flow is validated.

The first upgrade should be:

1. Gumroad webhook
2. small access service or GitHub Action entry point
3. GitHub API call to grant repository access

Do not add license keys or a custom auth backend until distribution volume makes
the manual process unmanageable.
