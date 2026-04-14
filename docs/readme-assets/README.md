# README Demo Assets

This directory is for generated terminal visuals used in the main README.

The source of truth is the VHS tapes under `docs/readme-assets/tapes/`.

Regenerate everything locally with:

```bash
make demo-assets
```

Render only selected scenarios with:

```bash
sh ./tools/render-demo-assets.sh docker-auth-hero missing-executable
```

Requirements:

- repo-local tools installed under `.tools/bin`
- `vhs` on `PATH`
- or `docker` on `PATH` so the script can use `ghcr.io/charmbracelet/vhs:latest`

The script rebuilds `./bin/faultline` before rendering so the visuals match the current checkout.

The canonical entrypoint lives in `tools/render-demo-assets.sh` so it is grouped with other local developer tools instead of release and packaging scripts.

The tapes use `Set Shell bash` because VHS only accepts a fixed set of named shells, not arbitrary paths like `/bin/sh`.

Current tapes:

- `docker-auth-hero` for noisy log to terminal-rendered diagnosis
- `docker-auth-fix` for concrete terminal-rendered remediation steps
- `docker-auth-modes` for quick terminal output followed by detailed terminal output
- `missing-executable` for missing runtime or binary failures
- `runtime-mismatch` for environment drift diagnosis