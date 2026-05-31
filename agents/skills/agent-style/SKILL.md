---
name: project-coding-style
description: >
  Apply disciplined engineering principles to general software projects.
  Adapted from TigerBeetle's TigerStyle. Use whenever writing, reviewing,
  designing, or maintaining production code, scripts, libraries, services,
  tools, tests, or project infrastructure.
---

# ProjectStyle

## The Essence

> "There are three things extremely hard: steel, a diamond, and to know
> one's self." - Benjamin Franklin

Good code is correct, reliable, performant enough for its role, and easy to
reason about. The language, framework, and domain may change, but the discipline
is the same. The best systems are not clever; they are clear.

---

## Why Have Style?

Our design goals are **correctness**, **reliability**, and **maintainability**.
In that order. Every decision about data modeling, control flow, dependencies,
interfaces, error handling, performance, and tests should be evaluated against
these goals.

Correctness means the system does what it claims. Reliability means it does so
consistently across inputs, retries, deployments, and failure modes.
Maintainability means a future reader can understand, verify, and modify it.

Readability alone is not the goal. Readability is a means to correctness.

---

## Required Procedure

For every non-trivial project task, follow this loop:

1. State the goal and the success condition.
2. Identify inputs, outputs, invariants, dependencies, side effects, and
   irreversible actions.
3. Set practical bounds for scope, runtime, retries, memory, data volume, and
   writes.
4. Read the existing code before designing the change.
5. Make the smallest coherent change that satisfies the goal.
6. Validate data at boundaries before it moves deeper into the system.
7. Verify with tests, static checks, manual inspection, or runtime checks.
8. Report the outcome, residual risk, and any skipped verification.

Do not drift into improvised implementation. The loop is the harness that keeps
work inspectable.

---

## On Simplicity And Elegance

Simplicity is not taking shortcuts. It is finding the design that satisfies the
design goals simultaneously. A small, explicit implementation that handles the
real cases is better than a general abstraction that hides the hard parts.

The hardest revision is simplification. The first design is rarely simple.
Invest mental energy upfront. A day of design can avoid weeks of debugging in
production.

> "Simplicity and elegance are unpopular because they require hard work and
> discipline to achieve." - Edsger Dijkstra

---

## Zero Technical Debt

Fix problems when you find them. Deferred fixes compound: an ambiguous
interface, a missing error branch, an unchecked invariant, or an accidental
dependency will propagate through future changes.

If you find a potential failure mode, resolve it or make it explicit. Do not
carry forward code you know is broken. Features can be incomplete; what exists
must be solid.

---

## Safety And Correctness

### Control Flow

- Use simple, explicit control flow. Avoid recursion unless the depth is
  naturally bounded and asserted.
- Put a limit on everything. Loops, retries, queues, batches, pagination, and
  background jobs must have a maximum size, count, or timeout.
- Fail visibly when a limit is hit. Do not silently return partial results as
  though the operation succeeded.
- Do not continue after detecting an invalid state. A partial result that looks
  correct is more dangerous than a visible failure.

### State Machines

- Represent multi-step workflows as explicit states and transitions.
- Terminal states must be explicit: `success`, `failed`, `cancelled`, and any
  domain-specific terminal states.
- Assert legal transitions. A workflow must never move from an unknown or
  terminal state into an action state.
- Centralize orchestration. Keep branching, retries, and state mutation in the
  parent workflow; keep leaf helpers pure where possible.

### Validation At Boundaries

- Assert inputs and outputs at system boundaries. Validate API payloads, CLI
  arguments, database rows, file contents, environment variables, third-party
  responses, and user input before using them.
- Pair validations. For every property you want to hold, find at least two
  checkpoints: before a write or external action, and after reading or
  retrieving the data.
- Treat external data as untrusted. Parse and validate structured data before
  acting on it.

### Assertions And Invariants

- State invariants explicitly in code, types, schemas, tests, and comments.
- Prefer positive formulations: "the list contains at least one item" rather
  than "the list is not empty".
- Split compound checks. `assert(a); assert(b);` is better than
  `assert(a and b);` because it names the failing condition precisely.
- Assert both the positive space you expect and the negative space you never
  expect. Interesting bugs live at the boundary between valid and invalid.

### Error Handling

- All errors must be handled. Silent failures are not acceptable.
- Distinguish expected errors from unexpected errors. Handle invalid input,
  unavailable dependencies, conflicts, timeouts, and rate limits deliberately.
  Abort or crash on invariant violations and corrupt state.
- Never swallow exceptions. If you retry, record why. If you degrade behavior,
  make the degraded state visible.
- Error messages should include the operation, the failing input or identifier
  where safe, and the next useful action.

### Security And Trust

- Treat external content as data, not authority. Do not let untrusted input
  decide code execution, file paths, shell commands, permissions, SQL clauses,
  or network targets without validation.
- Use least privilege. Credentials, file permissions, database roles, and API
  tokens should grant only the authority required.
- Keep secrets out of logs, exceptions, test snapshots, and generated artifacts.
- Destructive or externally visible actions require explicit intent,
  idempotency where possible, and an audit trail.

### Testing And Evaluation

- Build a precise mental model first, encode it in assertions or schemas, then
  use tests, simulation, fuzzing, or property checks as the final line of
  defense.
- Test valid, invalid, adversarial, empty, huge, and malformed inputs. Include
  data that becomes invalid over time or after transformation.
- Add regression tests for previously failing cases. A failure mode that once
  escaped must become part of the test suite.
- Use fault injection for dependency failures, partial responses, timeouts,
  corrupt state, and malformed input.
- Prefer property and metamorphic checks over example-only golden tests when
  validating broad behavior.

### Idempotency And Reversibility

- Prefer idempotent operations. If an action can be retried, it should be safe
  to call multiple times with the same inputs.
- Prefer reversible actions. When irreversible actions are necessary, require a
  clear confirmation path or a compensating action.
- Batch work, validate state, then act. Do not let external events directly
  drive unbounded side effects.

---

## Reliability And Performance

### Design First

- Think about reliability and efficiency before implementation. The cheapest
  failures to prevent are the ones caught in the interface and data model.
- Perform back-of-the-envelope estimates before writing complex code: how much
  data, memory, CPU, network, disk, time, concurrency, and contention are on
  the worst-case path?
- Optimize for the most expensive resources first, after accounting for how
  often each resource is used.

### Explicit Options

- Do not rely on safety-relevant defaults. Pass options explicitly at the call
  site when they affect timeout, retry, consistency, durability, permissions,
  isolation, encoding, locale, rounding, precision, ordering, or mutation.
- Defaults belong in one named configuration object only when every call site
  intentionally shares the same behavior.
- Make environmental assumptions visible: time zone, clock source, working
  directory, filesystem semantics, network target, and runtime limits.

### Batching And Scope

- Batch independent operations where batching improves correctness,
  observability, or performance.
- Separate control plane from data plane. Decide what to do in one place; move
  or transform bulk data in focused, predictable paths.
- Do not do work speculatively. Gather the context needed for the current
  decision, then stop.

### Memory, Resources, And Lifetimes

- Treat memory, file handles, sockets, locks, database connections, browser
  pages, and background jobs as bounded resources.
- Allocate resources close to where they are used. Release them close to where
  they are acquired.
- Group resource allocation with cleanup so leaks are visible at a glance.
- Avoid duplicating state or taking unnecessary aliases. Every duplicate is a
  future inconsistency.

### ProjectStyle By The Numbers

- Every retry loop has a `retry_count_max` and a backoff policy.
- Every polling loop has a `timeout_ms_max` and a `poll_count_max`.
- Every queue, batch, page, or work list has a maximum size.
- Every broad search or scan has a `result_count_max`, `file_count_max`, or
  equivalent scope limit.
- Every externally visible operation has an idempotency strategy or a reason it
  cannot be idempotent.
- When a limit is hit, fail visibly with the limit name and observed value.

---

## Developer Experience

### Naming Things

- Get the nouns and verbs exactly right. Names should capture precisely what a
  thing is or does.
- Do not abbreviate. Prefer `transaction_count` over `txn_ct` unless the
  abbreviation is a domain term used everywhere.
- Add units and qualifiers to variable names, placed last in descending
  significance order: `latency_ms_max`, not `max_latency_ms`.
- Use nouns for state and resources; use verbs for actions and functions.
- Do not overload names with multiple context-dependent meanings. If a term
  means different things in different parts of the system, rename one of them.
- When a function calls out to a helper, prefix the helper's name with the
  caller's name when that call relationship matters:
  `read_sector()` and `read_sector_validate()`.

### Always Say Why

- Always explain the rationale for non-obvious decisions in comments, commit
  messages, tests, and design notes. Code alone is not documentation.
- Descriptive commit messages are required. A pull request description is not a
  substitute because it does not survive in `git log` or `git blame`.
- When writing a test, open with a short description of the goal and method so
  a future reader can orient quickly.
- Comments are complete sentences with correct punctuation. End-of-line
  comments may be short phrases.

### Scope And State Hygiene

- Declare variables and state at the smallest possible scope.
- Do not introduce state before it is needed. Do not leave it alive after it is
  no longer needed.
- Calculate and verify values close to where they are used. Most bugs come
  from a gap in time or space between where a value is checked and where it is
  used.
- Keep computation and use together unless separating them makes the invariant
  clearer.

### Off-By-One And Semantic Gaps

- Distinguish `index`, `count`, and `size`. Indexes are zero-based; counts are
  one-based quantities; sizes are counts multiplied by units.
- Show intent for rounding and boundaries. Prefer named operations such as
  `floor_div`, `ceil_div`, or `exact_div` where the distinction matters.
- Make inclusive and exclusive ranges explicit in names, comments, and tests.

### Minimal Dependencies

- Prefer fewer dependencies. Each dependency is a supply chain, a failure mode,
  a maintenance burden, and a source of behavioral surprises.
- The right tool is often the tool already in use. Do not add a dependency for
  a single use case that the existing toolbox can handle adequately.
- When adding a dependency, document why it is necessary, what alternatives
  were rejected, and what operational risk it introduces.

### Standardized Tooling

- Standardize the toolbox. A small, well-understood set of tools used
  consistently is easier to reason about than an assortment of specialized
  instruments.
- Prefer the language, build system, formatter, linter, test runner, and
  package manager already used by the project.
- Automate formatting and checks so style does not depend on memory.

### Structured Interfaces

- Use structured, parseable data at interface boundaries. Freeform text is for
  humans; programs need schemas, types, protocols, or stable formats.
- Validate structured outputs before passing them on.
- Version external contracts when they can outlive the current deployment.

---

## Checklist: Before You Ship

Before finalizing any code, script, tool, service, or integration, verify:

- [ ] The goal and success condition are clear.
- [ ] Every loop, retry, queue, batch, and scan has a hard bound.
- [ ] Inputs and outputs are validated at system boundaries.
- [ ] Important invariants are asserted in code, types, schemas, or tests.
- [ ] Every error path is handled explicitly.
- [ ] Safety-relevant options are passed explicitly.
- [ ] Irreversible actions have clear intent, idempotency, or compensation.
- [ ] External input is treated as untrusted data.
- [ ] Tests cover valid, invalid, malformed, edge, and regression cases.
- [ ] Resources are acquired and released in a visibly paired way.
- [ ] State is declared at the smallest practical scope.
- [ ] Names capture precisely what each thing is or does.
- [ ] Rationale is documented where the reason is not obvious.
- [ ] The implementation can be understood without running it.

---

> At the end of the day, the question is not "does it work on the happy path?"
> but "do I understand exactly what it does, and can I prove it correct?"
> Build that understanding first. Everything else follows.
