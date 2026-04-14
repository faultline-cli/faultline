# Playbook Packaging Review

## Scope

This pass reviewed the shipped catalog for three things at once:

1. editorial quality of the playbook library
2. a precise starter versus premium boundary
3. a maintainable pack layout that matches the existing CLI composition model

The scored inventory for the commercialization pass inside this repository is in
`docs/playbook-matrix.csv`.

## Audit Summary

The pre-polish catalog was generous, but it mixed three different kinds of
content into one starter pack:

- broad first-run failures that should stay bundled
- specialized ecosystem playbooks that are better sold as deeper expertise
- generic fallback playbooks that needed tighter boundaries so they would not
  shadow stronger rules

The main editorial issues were:

- `runtime-mismatch` overlapped with more specific Node.js and Java rules
- `quality-gate-failure` was too generic and could collide with tool-specific
  lint playbooks
- `disk-full` and `runner-disk-full` were too close in wording and signal scope
- the starter catalog still contained several framework or toolchain-specific
  rules that added count more than first-run value

## Changes Made

### Starter catalog polish

- Narrowed `runtime-mismatch` so it now covers Python; Ruby; and Go runtime
  mismatches instead of shadowing `node-version-mismatch` and
  `java-version-mismatch`.
- Strengthened `node-version-mismatch` so the existing runtime-version fixture
  resolves to the specific Node.js rule instead of the generic fallback.
- Tightened `quality-gate-failure` into an explicit fallback for unknown lint;
  format; or static-analysis gates.
- Tightened `disk-full` so it captures generic host or container filesystem
  exhaustion while staying out of CI-runner-specific failures handled by
  `runner-disk-full`.

### Pack structure

- Kept the public starter catalog in `playbooks/bundled/`.
- Moved premium candidates into the sister premium repository rooted at
  `../faultline-premium/`, which is exposed locally in this repo through the
  ignored symlink `playbooks/packs/extra-local`.
- Updated local pack-resolution helpers so repository validation prefers the
  symlinked premium repo and still supports an explicit external pack path
  through `make extra-pack-link` or `EXTRA_PACK_DIR`.

### Test coverage

- Split fixture coverage so starter tests only exercise the bundled catalog.
- Added premium-pack fixture coverage through pack composition.
- Moved the Terraform noisy corpus check into a premium-pack corpus test so the
  starter release gate reflects the actual starter product.

## Scoring Framework

Every playbook was scored from 1 to 5 on these dimensions:

- `frequency`: how often the issue appears across real CI and developer
  workflows
- `breadth`: how many teams; stacks; and repositories the playbook applies to
- `urgency`: how blocking or painful the failure usually is
- `fix_confidence`: how confidently Faultline can guide a repair path
- `distinctiveness`: how much the playbook differentiates Faultline from a
  generic grep-and-docs tool
- `premiumness`: how strongly the rule feels like paid depth rather than
  baseline coverage

Scoring principles used in practice:

- starter kept the highest-frequency and widest-coverage failure classes even
  when they were not the most differentiated
- premium took specialized ecosystem coverage, deeper operator workflows, and
  higher-leverage long-tail rules
- source-detector starter coverage was kept intentionally small but real so
  `faultline inspect` remains useful without an add-on

## Exact Split

### Bundled starter pack: 67 playbooks

These remain bundled because they create the strongest day-one experience:

- Auth: `aws-credentials`, `docker-auth`, `git-auth`, `missing-env`,
  `ssh-key-auth`
- Build: `cache-corruption`, `dependency-drift`, `docker-build-context`,
  `eslint-failure`, `go-compile-error`, `go-sum-missing`, `install-failure`,
  `merge-conflict`, `node-out-of-memory`, `node-version-mismatch`,
  `npm-ci-lockfile`, `path-case-mismatch`, `pip-install-failure`,
  `pnpm-lockfile`, `python-module-missing`, `quality-gate-failure`,
  `runtime-mismatch`, `syntax-error`, `typescript-compile`,
  `working-directory`, `yarn-lockfile`
- CI: `artifact-upload-failure`, `pipeline-timeout`, `runner-disk-full`,
  `secrets-not-available`
- Deploy: `config-mismatch`, `container-crash`, `health-check-failure`,
  `image-pull-backoff`, `k8s-crashloopbackoff`, `port-conflict`
- Network: `connection-refused`, `dns-resolution`, `firewall-egress-blocked`,
  `network-timeout`, `rate-limited`, `ssl-cert-error`
- Runtime: `disk-full`, `env-var-missing`, `oom-killed`,
  `permission-denied`, `port-in-use`, `resource-limits`, `segfault`
- Test: `coverage-gate-failure`, `database-test-isolation`, `flaky-test`,
  `go-data-race`, `missing-test-fixture`, `order-dependency`,
  `parallelism-conflict`, `snapshot-mismatch`, `test-timeout`
- Source: `missing-error-propagation`, `panic-in-http-handler`

### Premium additions moved in this pass: 17 playbooks

These now live in the sister premium repository and are resolved locally through
`playbooks/packs/extra-local` when that symlink is present. They are
materially deeper and more specialized than the starter catalog:

- Build ecosystems: `cargo-build`, `dotnet-build`, `gradle-build`,
  `java-version-mismatch`, `maven-dependency-resolution`,
  `rubocop-failure`, `ruby-bundler`
- Deploy and IaC: `helm-chart-failure`, `terraform-init`,
  `terraform-state-lock`
- Network and runtime depth: `cors-error`,
  `database-connection-pool-exhausted`, `nodejs-unhandled-rejection`
- Test-runner specialization: `jest-worker-crash`, `pytest-fixture-error`,
  `rspec-failure`, `vitest-failure`

The current sister premium repository now contains 84 playbooks in total after
this move. The 17 playbooks above are the specific additions moved out of the
starter catalog during this pass.

## Why This Boundary Is Commercially Sound

The starter pack is still strong enough to win trust because it keeps:

- the core CI failure archetypes every team expects on day one
- strong coverage across auth; network; generic build; generic runtime; and
  generic test failures
- the highest-frequency JavaScript; Python; Go; and container issues
- two meaningful source-detector rules so `inspect` is not a stub
- core Kubernetes failures that create an immediate aha moment in modern deploy
  workflows

The premium pack is materially better rather than just larger because it adds:

- deeper JVM; Ruby; Rust; and .NET ecosystem expertise
- higher-value IaC and Helm troubleshooting
- richer test-runner coverage for mature engineering teams
- runtime and browser failures that are common in larger service or frontend
  estates but not required for a credible starter experience

Recommended count split for the starter repository itself:

- Starter: 67 playbooks shipped publicly in this repo
- Newly migrated premium additions from this pass: 17 playbooks
- Current sister premium repository total: 84 playbooks

Commercially this still lands near the target value split: the starter pack
contains roughly two-thirds of the practical first-run user value, while the
premium repository carries the denser ecosystem and operator expertise that is
most defensible to sell. Raw premium count is higher because the private pack is
where the long-tail catalog continues to accumulate.

## Current Review Baseline

After the starter-versus-premium cleanup pass, `EXTRA_PACK_DIR=/home/jake/workspace/faultline-premium make extra-pack-review`
reports 186 cross-pack pattern conflicts across the composed starter and premium
catalogs.

This report should be treated as an inspection artifact, not as a zero-defect
requirement. At this point the remaining conflicts are concentrated in adjacent
rule families that intentionally share some vocabulary:

- timeout families such as test, network, and CI timeout wording
- deploy and runtime restart families such as `container-crash`,
  `k8s-crashloopbackoff`, and `oom-killed`
- generic build failures versus ecosystem-specific rules such as Maven,
  Gradle, .NET restore, and language-specific compile failures
- artifact, cache, and CI housekeeping families where platforms emit similar
  operational phrases

The catalog should keep being reviewed for newly over-broad phrases, duplicate
IDs, or obviously shadowed rules. It should not be forced toward zero reported
conflicts when the remaining overlap reflects real neighboring root causes that
still rank correctly in fixtures, adversarial tests, and corpus tests.

## Boundary Rationale By Theme

- Keep in starter when the failure is common; broad; easy to validate; and
  central to the first impression of a CI diagnosis tool.
- Move to premium when the failure is tied to a specific framework; toolchain;
  runtime; or operator workflow and benefits from deeper workflow-specific
  remediation.
- Avoid moving iconic high-frequency failures such as `oom-killed`,
  `connection-refused`, `pipeline-timeout`, `image-pull-backoff`, or
  `k8s-crashloopbackoff`, because doing so would make the starter product feel
  intentionally weakened.

## Future Additions

Use these rules to keep the boundary clean:

- Add to starter only when the new playbook represents a common cross-team
  failure mode or materially improves first-run trust.
- Add to premium when the rule is provider-specific; framework-specific;
  infrastructure-heavy; or requires dense workflow knowledge to be genuinely
  useful.
- Prefer improving an adjacent starter playbook over adding a shallow starter
  variant.
- Keep generic fallback playbooks narrow and clearly subordinate to
  tool-specific rules.
- Run `make review`, `make extra-pack-check`, and `make test` after any future
  boundary change so the starter and premium catalogs remain deterministic.
- Treat `make extra-pack-review` as a change-detection tool: new spikes or new
  broad phrases deserve investigation, but some stable overlap between adjacent
  rule families is expected.
