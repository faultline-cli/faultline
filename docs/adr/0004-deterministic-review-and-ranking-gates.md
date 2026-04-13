# ADR 0004: Deterministic Review And Ranking Gates

- Status: Accepted
- Date: 2026-04-13

## Context

As playbook coverage expands, the catalog needs guardrails against duplicate matches, unstable ranking, and accidental regressions. The repository history added explicit playbook conflict management in commit `fccab7e` and refined ranking behavior with explicit weighted matching in commit `68f6587`.

The current system and docs describe these as deterministic review and scoring steps rather than heuristic post-processing.

## Decision

Faultline treats catalog review and result ranking as explicit, deterministic gates.

That includes:

- validating playbook structure before matching
- reviewing overlap conflicts through dedicated tooling
- using explicit scoring and weighting rules for ranking
- keeping evidence tied directly to matched signals
- preserving regression coverage through fixtures, corpus tests, and representative tests

## Consequences

- Adding playbooks requires thinking about confusable neighbors, not just positive matches.
- Release validation should continue to include `make review` and `make test`.
- Ranking changes are architecture-level behavior changes and should be introduced deliberately.
- Operators and automation can trust that repeated runs with the same input preserve ordering and evidence.

## References

- [SYSTEM.md](../../SYSTEM.md)
- [docs/playbooks.md](../playbooks.md)
- [README.md](../../README.md)
- Git history: `fccab7e` playbook conflict management and improvements, `68f6587` IDF weighted matching