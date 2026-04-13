# ADR 0002: Starter Catalog With External Premium Packs

- Status: Accepted
- Date: 2026-04-13

## Context

Faultline needs a public, easy-to-evaluate starter experience without collapsing all playbook coverage into a single public catalog. The repository history shows an explicit split between the bundled starter pack and premium coverage moved into separate distribution paths.

The strongest historical markers are the premium split in commit `27b6919`, pack installation support and related docs updates, and the current distribution model described in [docs/distribution.md](../distribution.md).

## Decision

Faultline ships a bundled starter catalog from `playbooks/bundled/` and composes extra packs on top of it rather than embedding premium coverage in the public starter release.

The supported composition model is:

- bundled starter catalog included in releases and Docker images
- extra packs installed persistently under `~/.faultline/packs/`
- one-off composition through repeatable `--playbook-pack` flags or `FAULTLINE_PLAYBOOK_PACKS`
- full catalog override only through `--playbooks <dir>` or `FAULTLINE_PLAYBOOK_DIR`

## Consequences

- Public releases stay simple to evaluate and distribute.
- Premium or team-specific playbooks can evolve independently of the binary.
- Docker uses the same installed-pack convention at `/home/faultline/.faultline/packs`, which keeps local and containerized behavior aligned.
- Starter playbooks should stay broad and useful on first run, while premium packs concentrate on deeper provider-specific, operations, security, or ecosystem-specific coverage.

## References

- [docs/architecture.md](../architecture.md)
- [docs/distribution.md](../distribution.md)
- [README.md](../../README.md)
- Git history: `27b6919` split out premium pack, `1eeaf9e` pack manifests, `364b6a4` install-path documentation refinement