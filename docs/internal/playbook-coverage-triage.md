# Playbook Coverage Triage

Date: 2026-05-02

This triage started from the 81 bundled playbooks reported by `go run ./cmd coverage`
as missing positive fixture evidence before the low-signal removals. The current
post-cleanup coverage backlog is 57 playbooks.

## Inputs

- `go run ./cmd coverage`
- `go run ./cmd coverage --json`
- `go run ./cmd fixtures review`
- `SYSTEM.md`
- `docs/playbooks.md`
- `docs/fixture-corpus.md`

Coverage snapshot:

- Total bundled playbooks: 173
- Bundled playbooks with positive fixture evidence: 116
- Missing positive fixture evidence: 57
- Fixture assertions: 301 positive, 16 negative
- Strict top-1 fixtures: 54

Staging review found existing top-match candidates for these uncovered
playbooks: `alpine-debian-incompatibility`, `circleci-resource-class-oom`,
`docker-permission-denied-nonroot`, `github-actions-matrix-axis-invalid`,
`http-auth-failure`, `invalid-config-schema`, `package-manager-mismatch`,
`process-killed-no-logs`, `resource-limits`, and `volume-mount-issue`.

## Decision Rules

Choose `fixture promotion` when the cluster is a reusable CI failure mode with
stable machine-produced evidence and a concrete fix path. Promotion may mean
promoting existing staging candidates or collecting a new public fixture before
the rule is considered release-backed.

Choose `removal` when the cluster is low-value, too project-specific, duplicate
with a stronger existing rule, or not defensible as a bundled default without
real evidence. Removal means removing the bundled playbooks and regenerated
failure docs unless future evidence justifies reintroducing them through the
normal fixture pipeline.

## Cluster Decisions

| Cluster | Playbooks | Decision | Rationale |
| --- | --- | --- | --- |
| HTTP auth rejection | `http-auth-failure` | Fixture promotion | Generic 401/403 build and deploy failures are common, auth scoped, and staging already has a top-match candidate. |
| Remote build cache | `remote-cache-miss` | Removal | Low-severity cache misses are often performance noise rather than a failing root cause, and no fixture currently proves a deterministic failure boundary. |
| Build config and schema | `config-file-missing`, `invalid-config-schema` | Fixture promotion | Missing or invalid config files are common across tools; schema validation already has a staged top-match candidate. |
| Container image and distro compatibility | `alpine-debian-incompatibility`, `base-image-breaking-change` | Fixture promotion | These are reusable Docker failure modes with stable package, binary, and base-image evidence; the Alpine/Debian rule already has staged evidence. |
| Dependency, package manager, and cache state | `cache-corruption`, `package-manager-mismatch`, `pnpm-lockfile-missing`, `workspace-dependency-mismatch` | Fixture promotion | These extend already central dependency coverage and have concrete remediation paths; package manager mismatch already has multiple staged candidates. |
| Filesystem and path shape | `build-output-path-mismatch`, `line-ending`, `npm-enoent-package-json`, `path-case-mismatch`, `symlink-in-ci` | Fixture promotion | These are common CI portability failures with stable log wording and actionable local verification. |
| Runtime tools and shell environment during build | `encoding-unicode`, `ffmpeg-avconv-missing`, `gradle-daemon-timeout`, `python-command-not-found`, `python-virtualenv-not-activated`, `shell-sh-vs-bash` | Fixture promotion | These are cross-repository environment failures with deterministic evidence lines and clear fixes. |
| Git ref selection | `git-refspec-mismatch` | Fixture promotion | Missing refs and bad branch names are common CI checkout failures with stable Git output. |
| Topology ownership log rules | `topology-ownership-mismatch`, `topology-failure-clustered` | Removal | These require topology context but are modeled as bundled log playbooks without fixture proof; keep topology-backed behavior only where deterministic fixtures exist. |
| tox invocation | `tox-invocation-error` | Fixture promotion | tox failures are reusable Python CI signals and should be backed by a focused real fixture rather than removed. |
| CLI interface drift | `cli-flag-changed` | Fixture promotion | Removed or renamed flags are a common toolchain break with concrete remediation and stable wording. |
| Azure Pipelines provider failures | `azure-pipelines-service-connection`, `azure-pipelines-task-not-found` | Fixture promotion | Provider-specific, but common enough for bundled CI coverage if backed by real Azure log snippets. |
| CircleCI resource failures | `circleci-resource-class-invalid`, `circleci-resource-class-oom` | Fixture promotion | Scheduler and resource-class failures are deterministic provider failures; the OOM variant already has staged evidence. |
| GitHub Actions matrix and runner capacity | `github-actions-matrix-axis-invalid`, `github-actions-runner-capacity` | Fixture promotion | GitHub Actions is central to the corpus, and matrix/capacity failures are stable provider-level modes; matrix axis already has staged evidence. |
| GitLab CI config and runner resources | `gitlab-ci-yaml-invalid`, `gitlab-job-log-limit`, `gitlab-no-runner` | Fixture promotion | These are common GitLab scheduler/config failures with direct machine wording and clear fixes. |
| Jenkins platform failures | `jenkins-agent-offline`, `jenkins-plugin-missing` | Fixture promotion | Jenkins remains a common CI provider; keep only if promoted with direct Jenkins log evidence. |
| Skipped or untriggered CI workflow variants | `step-skipped-unexpectedly`, `workflow-not-triggered` | Removal | These overlap the silent-failure family and are often absence-of-run conditions rather than reliable failing log signatures. |
| Documentation link checking | `link-checker-failure` | Removal | This is a niche documentation workflow and should move out of the bundled default unless a real fixture proves broad value. |
| Network transport and registry availability | `dns-enotfound`, `ipv6-ipv4-resolution`, `proxy-configuration`, `registry-outage` | Fixture promotion | DNS, proxy, and registry failures are recurring CI causes with stable evidence and concrete operator fixes. |
| Docker runtime container setup | `docker-permission-denied-nonroot`, `entrypoint-misconfigured`, `volume-mount-issue` | Fixture promotion | These are reusable container runtime failures; Docker socket permission and volume mount variants already have staged evidence. |
| Runtime termination and limits | `process-killed-no-logs`, `resource-limits` | Fixture promotion | Resource limit failures are high-value CI diagnoses; both IDs already have staged evidence. |
| Source detector rules | `missing-error-propagation`, `panic-in-http-handler`, `unawaited-promise` | Fixture promotion | These should remain bundled, but coverage should be backed by source-fixture accounting rather than real log promotion. |
| Silent CI failure family | `artifact-missing`, `cache-miss-non-fatal`, `continue-on-error`, `empty-deployment-target`, `empty-quality-check`, `ignored-exit-code`, `skipped-critical-step` | Fixture promotion | Silent failures are a first-class product surface; each rule needs positive fixture proof before the family is release-backed. |
| Zero tests executed | `zero-tests-executed` | Fixture promotion | This is a high-value silent test failure with direct log evidence in common runners. |
| Database migration tests | `database-migration-timeout` | Fixture promotion | Migration lock/timeouts are common integration-test failures with concrete database evidence. |
| Test nondeterminism without direct root-cause proof | `clock-drift`, `random-seed-not-fixed` | Removal | Both are low-severity and easy to over-match without strong fixture evidence; reintroduce only after real fixtures prove stable signals. |
| Jest executable setup | `jest-command-not-found` | Fixture promotion | Common Node test setup failure; keep if backed by a positive fixture and a generic missing-executable near-miss. |
| Generic test runner failures | `cucumber-step-failure`, `jest-worker-crash`, `junit-test-failure`, `test-assertion-with-reason`, `timezone-diff` | Fixture promotion | These represent broadly reusable test failure mechanisms and should be backed with concise positive and near-miss fixtures. |
| Project-specific test suites | `asciidoctor-jbehave-test-failure`, `flexget-test-failure`, `mdanalysis-test-failure`, `nikola-build-test-failure`, `nupic-test-failure`, `openai-gym-test-failure`, `pip-install-test-failure`, `python-parser-test-failure`, `readthedocs-build-test-failure`, `sentry-elasticsearch-test-failure`, `translate-toolkit-test-failure`, `youtube-dl-test-failure` | Removal | These are narrow repository or project ecosystems and do not belong in the bundled default without accepted real fixtures. |

## Summary

- Original fixture promotion decision: 61 playbooks
- Removal: 20 playbooks
- Current fixture promotion backlog after four fixture promotions: 57 playbooks
- Total original triage accounted for: 81 playbooks

Promotion should happen in small batches with `make review`, `make test`, and
`make fixture-check` after each batch. Removal batches must also regenerate
failure docs with `make docs-generate` and run `make docs-check`.
