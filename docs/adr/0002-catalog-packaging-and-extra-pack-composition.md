# ADR 0002: Catalog Packaging And Extra Pack Composition

- Status: Accepted
- Date: 2026-04-13

## Context

Faultline needs a public, easy-to-evaluate default experience without collapsing all playbook coverage into a single catalog. The repository history shows an explicit split between the bundled catalog and additional pack coverage moved into separate distribution paths.

The strongest historical markers are the pack split in commit `27b6919`, pack installation support and related docs updates, and the current distribution model described in [docs/distribution.md](../distribution.md).

## Decision

Faultline ships a bundled catalog from `playbooks/bundled/` and composes extra packs on top of it rather than embedding all specialized coverage in the default release.

The supported composition model is:

- bundled catalog included in releases and Docker images
- extra packs installed persistently under `~/.faultline/packs/`
- one-off composition through repeatable `--playbook-pack` flags or `FAULTLINE_PLAYBOOK_PACKS`
- full catalog override only through `--playbooks <dir>` or `FAULTLINE_PLAYBOOK_DIR`

## Consequences

- Public releases stay simple to evaluate and distribute.
- Team-specific or extended playbooks can evolve independently of the binary.
- Docker uses the same installed-pack convention at `/home/faultline/.faultline/packs`, which keeps local and containerized behavior aligned.
- Bundled playbooks should stay broad and useful on first run, while extra packs concentrate on deeper provider-specific, operations, security, or ecosystem-specific coverage.

## References

- [docs/architecture.md](../architecture.md)
- [docs/distribution.md](../distribution.md)
- [README.md](../../README.md)
- Git history: `27b6919` split out an external pack, `1eeaf9e` pack manifests, `364b6a4` install-path documentation refinement