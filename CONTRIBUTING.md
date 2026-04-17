# Contributing

Faultline accepts focused fixes, deterministic playbook improvements, and regression fixtures grounded in real CI failures.

## Workflow

1. Keep changes small and explicit.
2. Run `make test` after code or playbook changes.
3. Run `make cli-smoke` when user-facing output, examples, packaging, or release scripts change.
4. Run `make review` after playbook changes.
5. Run `make fixture-check` if fixture or ranking behavior changes.

Keep the shipped user narrative aligned with [`docs/release-boundary.md`](docs/release-boundary.md): core log diagnosis first, companion surfaces second, and experimental paths clearly flagged.

## Fixture Handling

Use `faultline fixtures ingest` to collect candidate cases into `fixtures/staging/` locally.

Before anything is committed:

1. Remove or redact credentials, emails, signed URLs, internal hostnames, and other identifying data.
2. Keep only the evidence needed to preserve the failure mode.
3. Promote accepted cases into `fixtures/real/` with a stable expectation.

Do not commit raw staging fixture YAML files.

## Pull Requests

Pull requests should explain the failure mode being fixed, the expected behavior change, and the validation used.

## Issues

When opening an issue, prefer the repository templates.

- Use the diagnosis-gap template when Faultline misses a CI failure or ranks the wrong playbook first.
- Use the bug-report template for crashes, broken commands, packaging problems, or documentation defects.

For diagnosis issues, include the smallest sanitized log snippet that still reproduces the problem, the expected playbook or explanation, and the actual result.
