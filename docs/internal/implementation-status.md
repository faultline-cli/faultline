# Implementation Status

Faultline V1 is now implemented as a CLI-only product.

## Shipped Scope

- CLI entrypoint in [`cmd/`](../../cmd)
- deterministic analyzer modules in [`internal/engine/`](../../internal/engine), [`internal/playbooks/`](../../internal/playbooks), [`internal/matcher/`](../../internal/matcher), and [`internal/output/`](../../internal/output)
- bundled YAML playbooks in [`playbooks/`](../../playbooks)
- CLI-first build and container workflow in [`Makefile`](../../Makefile) and [`Dockerfile`](../../Dockerfile)
- tests covering command behavior, loading, ranking, and formatting

## Repository State

- legacy hosted-service, frontend, storage, and scan-era code has been removed
- the repository structure now matches the CLI architecture described in [`SYSTEM.md`](../../SYSTEM.md)
- remaining work should extend the CLI and playbooks directly instead of reviving removed product shapes

## Completion Checklist

The repository currently satisfies the V1 migration goal when these remain true:

- `faultline analyze`, `faultline list`, and `faultline explain` work end to end
- playbooks load deterministically from disk
- playbook overlap conflicts can be reviewed deterministically before rule changes
- matching, ranking, and output remain deterministic across repeated runs
- JSON output stays stable for automation
- Docker packaging includes the binary and bundled playbooks
- docs describe the shipped CLI behavior