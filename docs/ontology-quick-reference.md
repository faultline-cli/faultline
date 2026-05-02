# Ontology Quick Reference

Faultline playbooks may declare lightweight taxonomy fields:

```yaml
domain: dependency
class: lockfile-drift
mode: npm-ci-lockfile-outdated
```

These fields are metadata only. They do not affect matching, ranking, scoring,
or output unless a reporting surface explicitly reads them.

## Domains

Common domain values in the bundled catalog include:

| Domain | Purpose |
|--------|---------|
| `dependency` | Package resolution, lockfiles, registries, caches |
| `runtime` | Missing executables, permissions, resources, crashes |
| `container` | Docker image build, pull, daemon, and manifest issues |
| `auth` | Credentials, tokens, scopes, and secret availability |
| `network` | DNS, connectivity, TLS, proxy, and timeout failures |
| `ci-config` | Workflow syntax, runner setup, provider configuration |
| `test-runner` | Test framework, fixture, isolation, and coverage failures |
| `database` | Connection, migration, version, and isolation failures |
| `filesystem` | Missing paths, permissions, disk, symlink, casing issues |
| `platform` | Deployment, scheduler, infrastructure, and provider failures |
| `source` | Compilation, lint, formatting, and code-quality failures |

## Authoring Guidance

- Use `domain` for the operational subsystem.
- Use `class` for the reusable failure family.
- Use `mode` for the concrete root cause or signature.
- Keep metadata stable and explicit; do not invent a new value for one playbook
  if an existing value already describes the same boundary.
- Add positive fixture evidence for every shipped playbook when possible.
- Add negative or disallowed-playbook assertions for nearby confusable rules.

## Coverage Command

The shipped coverage command reports fixture evidence for the resolved catalog:

```bash
faultline coverage
faultline coverage --json
faultline coverage --fixture-dir fixtures
faultline coverage --playbooks playbooks/bundled
faultline coverage --playbook-pack examples/packs/minimal
```

Human output includes totals, category/domain grouping, missing positive
fixtures, and duplicate-ID status. JSON output exposes the same data through
the `coverage.Report` schema in `internal/coverage/report.go`.

Unsupported historical examples:

- `faultline coverage --domain=dependency`
- `faultline coverage --depth=shallow`
- `faultline coverage --gaps`
- `faultline coverage --format=json`
- `faultline playbooks validate --ontology`

Those commands appeared in earlier planning docs but are not part of the
current CLI.

## Review Checklist

- Does the playbook have a clear `domain`, `class`, and `mode`?
- Does the metadata agree with the root cause described in `summary`,
  `diagnosis`, and `fix`?
- Does `faultline coverage` show positive fixture evidence for the playbook?
- Are known confusable rules protected by negative or disallowed-playbook
  assertions?
- Do `make test` and `make review` pass after the change?
