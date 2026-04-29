# ADR 0008: Playbook Catalog Scalability Through Composition And Inheritance

- Status: Accepted
- Date: 2026-04-25

## Context

As the bundled playbook catalog grew from 77 playbooks at v0.3.0 to 178+ by v0.4.1, naive copy-paste expansion becomes untenable: duplicate matching logic, divergent guidance text, and brittle maintenance.

The repository already handles duplicates and conflict resolution through `make review` (`internal/playbooks/conflicts.go`). But a larger catalog still creates pressure toward:

- redundant match blocks across similar failure signatures
- guidance text that diverges for the same root cause in different environments
- no clear mechanism to express "this is a narrower variant of that"

The scaling model is documented in [docs/playbooks.md](../playbooks.md) and the inheritance resolver is implemented in `internal/playbooks/inheritance.go`.

## Decision

Faultline's catalog scaling model relies on four mechanisms, each addressing a different axis of growth:

1. **`extends` inheritance**: A child playbook can reference a parent by id. The engine merges parent fields into the child at load time, with child values overriding parent values. Cycles are rejected. This keeps a shared root cause in one place while allowing environment-specific guidance and match constraints to live in child playbooks.

2. **Reusable signal fragments** via `match.use` or `faultline-matchers.yaml`: Common evidence patterns can be factored out and composed into multiple playbooks without duplicating signal logic.

3. **`match.partial` groups**: Multiple soft signals that are individually inconclusive can be combined into a decisive threshold. This avoids broadening individual `match` rules while capturing multi-signal evidence.

4. **Explicit constraint fields** (`tags`, `stage_hints`, `context_filters`, `source`): Narrowing without modifying match patterns. Playbooks stay broad in their core signal and use constraint fields to sharpen scope.

The catalog scales through composition of these primitives rather than accumulating monolithic playbooks or unconstrained duplication.

## Current Status

As of v0.4.4, `extends`-based inheritance is in production use in the bundled catalog. `node-missing-executable` extends `missing-executable` and is the first shipped example of a child playbook that narrows a generic root cause with environment-specific match patterns and runner exclusions.

The engine also implements **NativeAny scoring**: when a child playbook is evaluated, only its own pre-inheritance `match.any` patterns (`NativeAny` in `model.Playbook`) contribute to the child's anyScore. Inherited patterns are still collected as evidence but do not increase the child's score, so the parent wins on generic logs and the child wins only when its distinctive patterns fire. IDF weights (`computeAnyWeights`) also use `NativeAny` for child playbooks to prevent inherited patterns from diluting the parent's rarity signal.

Prior to v0.4.4, the engine supported the `extends` field but no bundled playbooks used it.

## Consequences

- New playbooks for the same root cause should extend a shared base rather than duplicating match and guidance content
- The conflict review gate (`make review`) remains the primary quality check for both flat and inheritance-based families
- Inheritance cycles are a load-time error; they will not silently produce wrong results
- The composition model gives a deterministic equivalent of "components" without introducing a second matching language
- `match.partial` is the correct tool for combining weak signals; broadening individual match rules to compensate is an anti-pattern
- Child playbooks must use `match.none` to exclude log lines that the parent handles generically; without exclusions the child competes with the parent on every generic log that fires the inherited patterns

## References

- [docs/playbooks.md](../playbooks.md) — Scaling model and authoring guidance
- [internal/playbooks/inheritance.go](../../internal/playbooks/inheritance.go)
- [internal/playbooks/inheritance_test.go](../../internal/playbooks/inheritance_test.go)
- [internal/playbooks/conflicts.go](../../internal/playbooks/conflicts.go)
- [playbooks/bundled/log/build/node-missing-executable.yaml](../../playbooks/bundled/log/build/node-missing-executable.yaml) — First bundled example of `extends`
- [docs/releases/v0.4.0.md](../releases/v0.4.0.md) — Catalog grew from 77 to 123 playbooks
- [docs/releases/v0.4.1.md](../releases/v0.4.1.md) — Catalog at 170 bundled playbooks
- [docs/releases/v0.4.4.md](../releases/v0.4.4.md) — First production inheritance use, NativeAny scoring
