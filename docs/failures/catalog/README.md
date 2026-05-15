# Faultline Failure Catalog

This index links to every bundled playbook. Each page explains the failure, shows the log signals Faultline matches, and lists diagnosis and fix steps.

**177 playbooks** across 8 categories.

Run `faultline list` to browse the catalog inside the terminal.

## Auth (11)

- [AWS credentials missing or invalid](../auth/aws-credentials.md) — The job could not authenticate with AWS.
- [Docker registry authentication failure](../auth/docker-auth.md) — CI could not authenticate to the container registry before an image pull or push.
- [Expired credentials or rotated secrets](../auth/expired-credentials.md) — A CI job is using credentials that have expired, been revoked, or rotated
outside of CI.
- [Git authentication failure](../auth/git-auth.md) — The CI runner could not authenticate with the remote Git repository.
- [HTTP request rejected with 401 or 403 status](../auth/http-auth-failure.md) — An HTTP request was rejected with a 401 Unauthorized or 403 Forbidden status.
- [Kubernetes cluster authentication failure](../auth/kubectl-auth.md) — `kubectl` could not authenticate with the Kubernetes API server.
- [Missing required environment variable](../auth/missing-env.md) — A required environment variable was not set in the CI environment.
- [npm or yarn registry authentication failure](../auth/npm-registry-auth.md) — The Node.
- [OIDC token request failed or token invalid](../auth/oidc-token-failure.md) — An OIDC token request failed or the issued token was rejected by the cloud provider.
- [SSH key authentication failure](../auth/ssh-key-auth.md) — An SSH operation (git clone, deploy, rsync) failed because the SSH key was not found, not loaded into the agent, or the remote host does not accept it.
- [SSH permission denied (publickey) authentication failed](../auth/ssh-permission-denied.md) — SSH authentication to a Git host or remote server failed with explicit "Permission denied (publickey)" message.

## Build (67)

- [Alpine vs Debian/Ubuntu package or binary incompatibility](../build/alpine-debian-incompatibility.md) — A CI build or container fails because it mixes Alpine Linux (musl libc) and
Debian/Ubuntu (glibc) artifacts.
- [Ansible YAML task or variable file syntax error](../build/ansible-yaml-syntax-error.md) — Ansible failed to parse a YAML file — a task file, variable file, role
default, or handler — while loading the playbook.
- [Docker base image update introduced a breaking change](../build/base-image-breaking-change.md) — A Docker build that previously succeeded now fails because the base image
tag (e.
- [Required build input file missing](../build/build-input-file-missing.md) — A build or lint step referenced a file that was never generated, was copied to the wrong path, or is missing from the repository checkout.
- [Build output artifact not found at expected path](../build/build-output-path-mismatch.md) — A CI step that consumes the output of a prior build step cannot find the
expected artifact at the configured path.
- [Docker BuildKit session lost](../build/buildkit-session-lost.md) — Docker BuildKit lost its session while loading the build definition or context, so the image build stopped before it could execute any build steps.
- [Corrupted or stale dependency cache](../build/cache-corruption.md) — CI restored a dependency cache, but the cached contents were stale,
corrupted, or built for an incompatible environment.
- [Rust cargo build compilation failure](../build/cargo-compile-error.md) — The Rust compiler rejected the crate with one or more hard errors — type mismatches, borrow checker violations, unresolved imports,
- [CLI tool flag or command changed between versions](../build/cli-flag-changed.md) — A CI script or Makefile invokes a CLI tool with a flag or subcommand that no
longer exists in the installed version.
- [Required configuration file not found](../build/config-file-missing.md) — A required configuration file is absent from the CI workspace.
- [Dependency version drift](../build/dependency-drift.md) — Dependency constraints have drifted apart enough that the package manager
cannot compute a valid install plan.
- [Dependency removed or yanked from upstream registry](../build/dependency-removed-upstream.md) — A required dependency has been removed, yanked, or unpublished from the
upstream registry.
- [Docker build context or Dockerfile path issue](../build/docker-build-context.md) — Docker could not read the Dockerfile or a required file from the selected build context, so the image build stopped before executing the full build.
- [Docker manifest not found or bad image tag](../build/docker-manifest-not-found.md) — Docker could not pull an image because the tag does not exist in the registry, the manifest is missing, or the image reference is malformed.
- [Docker COPY failed source file missing from build context](../build/dockerfile-copy-source-missing.md) — Docker COPY or ADD instruction failed because the source file does not exist in the build context.
- [.NET NuGet package restore failure](../build/dotnet-restore.md) — `dotnet restore` or a build that implicitly runs restore could not download or resolve one or more NuGet packages.
- [Unicode or character encoding error](../build/encoding-unicode.md) — A CI step failed because a file or data stream contains characters that
cannot be interpreted in the assumed encoding.
- [ESLint or linter check failure](../build/eslint-failure.md) — A linting or formatting check step (`eslint`, `tslint`, `prettier`) found violations in the source code and exited with a non-zero
- [ffmpeg or avconv not available in the CI environment](../build/ffmpeg-avconv-missing.md) — The job requires ffmpeg (or its libav counterpart avconv) for media processing but neither tool is installed or reachable on
- [Docker base image uses a floating latest tag](../build/floating-docker-base-image.md) — A Dockerfile uses a `:latest` base image tag, so rebuilds can pull different bytes over time without a source change.
- [Code formatting check failure](../build/formatting-failure.md) — A CI formatting check failed because one or more source files do not match the project's required code style as
- [Git refspec or branch does not exist](../build/git-refspec-mismatch.md) — A git checkout or clone operation failed because the requested refspec (branch name, tag, or commit reference) does not exist on the remote repository.
- [Go compilation error](../build/go-compile-error.md) — The Go compiler rejected the package because of a semantic error — undefined
identifier, type mismatch, unused import, or interface non-compliance.
- [Missing go.sum entry](../build/go-sum-missing.md) — The build needs a module checksum that is missing from `go.
- [Gradle build failure](../build/gradle-build.md) — Gradle exited with a non-zero status, indicating one or more tasks failed.
- [Gradle daemon wrapper lock timeout](../build/gradle-daemon-timeout.md) — The Gradle or Maven daemon could not acquire an exclusive lock on cached wrapper files, causing the build tool initialization to timeout.
- [Package installation failure](../build/install-failure.md) — The dependency installation step failed because the requested package or version could not be found in the configured registry, package
- [Configuration file fails schema validation](../build/invalid-config-schema.md) — A configuration file exists but fails validation because a required field is missing, an unrecognized field is present, or a
- [Line ending mismatch (CRLF vs LF)](../build/line-ending.md) — A script or source file contains Windows-style CRLF line endings where
Unix LF endings are expected.
- [Maven Java compilation failure](../build/maven-compile-error.md) — The Maven compiler plugin failed during compilation.
- [Maven or Gradle dependency resolution failure](../build/maven-dependency-resolution.md) — The Java build could not resolve one or more dependencies or plugins from the configured artifact repositories.
- [Unresolved merge conflict in source code](../build/merge-conflict.md) — A file contains unresolved merge conflict markers (`<<<<<<<`, `=======`,
`>>>>>>>`).
- [Required executable or runtime binary missing](../build/missing-executable.md) — The job tried to launch a required tool or runtime binary, but that executable was missing from the image, runner, or expected path.
- [MSBuild task or target failure](../build/msbuild-error.md) — MSBuild exited with a fatal `error MSBxxxx` diagnostic.
- [Multi-stage Docker build artifact not copied correctly](../build/multistage-build-missing-artifact.md) — A multi-stage Docker build failed because a `COPY --from=<builder>` instruction references a file or directory that was not produced by
- [node-gyp missing build tools](../build/node-gyp-missing-build-tools.md) — `node-gyp` failed because required build tools (Python, C++ compiler, etc.
- [Node.js runtime or tool missing in CI](../build/node-missing-executable.md) — The CI job tried to run a Node.
- [Node.js JavaScript heap out of memory](../build/node-out-of-memory.md) — The Node.
- [Node.js version mismatch](../build/node-version-mismatch.md) — The Node.
- [npm ci lockfile mismatch](../build/npm-ci-lockfile.md) — `npm ci` found a missing or out-of-sync `package-lock.
- [npm EACCES permission denied in node_modules](../build/npm-eacces-permission-denied.md) — npm failed with an `EACCES` error while trying to write to `node_modules` or the global npm cache.
- [npm ENOENT package.json missing](../build/npm-enoent-package-json.md) — npm could not find a `package.
- [npm ERESOLVE dependency tree conflict](../build/npm-eresolve-conflict.md) — npm encountered a dependency tree conflict that it could not automatically resolve.
- [npm peer dependency conflict](../build/npm-peer-dependency-conflict.md) — npm could not build a valid dependency tree because one package requires a peer version that conflicts with what the project currently installs.
- [Package manager mismatch in lockfile](../build/package-manager-mismatch.md) — CI used one package manager (e.
- [File path case mismatch](../build/path-case-mismatch.md) — A file or module import path uses a different capitalisation from the actual filename on disk.
- [pip hash-checking mode failure](../build/pip-hash-mismatch.md) — pip rejected one or more downloaded packages because their hash did not match the expected value recorded in `requirements.
- [pip package install failure](../build/pip-install-failure.md) — `pip install` could not satisfy one or more package requirements.
- [pnpm lockfile mismatch or frozen install failed](../build/pnpm-lockfile.md) — The pnpm lockfile (`pnpm-lock.
- [pnpm frozen lockfile missing or mismatch](../build/pnpm-lockfile-missing.md) — pnpm failed because the lockfile is missing, out of sync, or was not committed.
- [Poetry lockfile drift](../build/poetry-lockfile-drift.md) — Poetry is installing from a lockfile that no longer matches `pyproject.
- [Python or pip command not found](../build/python-command-not-found.md) — The `python` or `pip` command is not available in the CI environment's PATH.
- [Python externally managed environment error](../build/python-externally-managed.md) — pip refused to install packages because Python 3.
- [Python module not found](../build/python-module-missing.md) — Python could not import a required module because it is missing from the environment used by the failing CI step
- [Python virtualenv not activated or interpreter mismatch](../build/python-virtualenv-not-activated.md) — Python code is using a system interpreter or wrong virtualenv instead of the activated one, causing module import failures or version mismatches.
- [Lint, format, or static-analysis gate failure](../build/quality-gate-failure.md) — A non-compiler quality gate failed during CI.
- [Python, Ruby, or Go runtime version mismatch](../build/runtime-mismatch.md) — The installed Python, Ruby, or Go runtime does not satisfy the version
constraint declared by the project.
- [Bash syntax used under a POSIX sh shell](../build/shell-dialect-mismatch.md) — A CI workflow or script selects `sh` but uses Bash-only syntax such as `[[ .
- [Shell compatibility issue (sh vs bash)](../build/shell-sh-vs-bash.md) — A shell script uses bash-specific syntax but is executed by `/bin/sh`, which
is commonly `dash` on Ubuntu/Debian or `ash` on Alpine.
- [Symlink resolution failure in CI](../build/symlink-in-ci.md) — A symlink in the repository or the build workspace is broken or cannot be
resolved in the CI environment.
- [Syntax error in source code](../build/syntax-error.md) — The compiler or interpreter hit invalid source syntax and stopped before the
build could continue.
- [Cross-ownership boundary failure](../build/topology-boundary-crossed.md) — A change in one ownership zone has broken code in a separate ownership zone, suggesting the failure originates at a
- [tox environment command failed (InvocationError)](../build/tox-invocation-error.md) — tox ran a command inside a virtual environment and the command exited with a non-zero status.
- [TypeScript compile or type-check failure](../build/typescript-compile.md) — The TypeScript compiler rejected the code because a type, symbol, or module import does not match the project's declared types.
- [Wrong working directory](../build/working-directory.md) — A command ran in an unexpected working directory.
- [Monorepo workspace dependency version mismatch](../build/workspace-dependency-mismatch.md) — A package in the monorepo workspace depends on a version of another workspace package that doesn't match the version published
- [Yarn lockfile out of date](../build/yarn-lockfile.md) — Yarn was run with `--frozen-lockfile` (the recommended CI flag) but the `yarn.

## Ci (26)

- [CI artifact upload or download failed](../ci/artifact-upload-failure.md) — A CI job failed to upload a build artifact, or a downstream job failed to download an artifact produced by an earlier job.
- [Azure Pipelines service connection failure](../ci/azure-pipelines-service-connection.md) — An Azure Pipelines job failed because the service connection required by a task does not exist, is not authorized, or has expired credentials.
- [Azure Pipelines task version not found or deprecated](../ci/azure-pipelines-task-not-found.md) — An Azure Pipelines job could not start because a pipeline task references a version that has been deprecated, removed, or
- [CircleCI config.yml validation error](../ci/circleci-config-validation.md) — CircleCI rejected the `.
- [CircleCI machine image or resource class not valid](../ci/circleci-resource-class-invalid.md) — A CircleCI job failed to start because the specified machine image or resource class is no longer valid, was removed,
- [CircleCI job killed — resource class memory limit exceeded](../ci/circleci-resource-class-oom.md) — The CircleCI job was killed because it exceeded the memory limit of the selected resource class.
- [Git shallow clone missing history or tags](../ci/git-shallow-checkout.md) — The CI job performed a shallow Git clone and a later step requires full commit history or tags.
- [Git submodule not initialized or updated](../ci/git-submodule-not-initialized.md) — A Git submodule referenced by the repository has not been initialized or
updated.
- [GitHub Actions job cancelled by concurrency policy](../ci/github-actions-concurrency-cancel.md) — A GitHub Actions workflow run was automatically cancelled because a newer
run for the same concurrency group started.
- [GitHub Actions environment variable not persisted across steps](../ci/github-actions-env-not-persisted.md) — A GitHub Actions step exported an environment variable using the deprecated
`::set-env` workflow command, which is now disabled.
- [GitHub Actions matrix axis variable undefined](../ci/github-actions-matrix-axis-invalid.md) — A GitHub Actions workflow uses an undefined matrix axis variable in a step or action, causing the workflow to fail during variable resolution.
- [GitHub Actions missing actions/checkout before accessing repo files](../ci/github-actions-missing-checkout.md) — GitHub Actions job tried to access repository files (build, test, run scripts) but the repository was never checked out.
- [GitHub Actions GITHUB_TOKEN permission denied](../ci/github-actions-permission.md) — The `GITHUB_TOKEN` used by this workflow does not have the permissions required for the requested GitHub operation.
- [GitHub Actions runner capacity or queue timeout](../ci/github-actions-runner-capacity.md) — The GitHub Actions job waited too long to be picked up by a runner, or the runner became unresponsive mid-job.
- [GitHub Actions workflow validation or action reference error](../ci/github-actions-syntax.md) — GitHub Actions rejected the workflow definition or could not resolve a referenced action, so the workflow could not validate cleanly
- [GitLab CI artifact missing or expired](../ci/gitlab-ci-artifact-expired.md) — A downstream GitLab CI job could not retrieve an artifact produced by an earlier job because it expired, was never
- [GitLab CI pipeline configuration invalid](../ci/gitlab-ci-yaml-invalid.md) — The `.
- [GitLab CI job log size limit exceeded](../ci/gitlab-job-log-limit.md) — The GitLab CI job produced more output than the runner's configured log size limit.
- [GitLab CI job stuck — no matching runner](../ci/gitlab-no-runner.md) — GitLab CI could not find an available runner that matches the job's tag requirements or the project has no registered runners at all.
- [Jenkins build agent went offline during job](../ci/jenkins-agent-offline.md) — The Jenkins build agent disconnected or went offline while running a job.
- [Jenkins required plugin missing or incompatible](../ci/jenkins-plugin-missing.md) — A Jenkins job or pipeline step failed because a required plugin is not installed, is disabled, or has an incompatible version.
- [Orphaned CI resources blocking subsequent runs](../ci/orphaned-resources.md) — A CI job failed because a previous run left behind a container, network,
volume, or bound port that was not cleaned up.
- [CI job or pipeline exceeded time limit](../ci/pipeline-timeout.md) — The CI job ran past its configured time limit and was killed by the CI system.
- [CI runner disk full](../ci/runner-disk-full.md) — The CI runner has run out of available disk space.
- [Self-hosted runner update permission failure](../ci/runner-update-permission-denied.md) — A self-hosted runner tried to update itself or adjust worker process settings, but the environment denied access to the runner
- [CI secret or environment variable not available](../ci/secrets-not-available.md) — A required secret or environment variable is not available in this CI job.

## Deploy (11)

- [Deployment configuration mismatch](../deploy/config-mismatch.md) — The deployment failed because a required configuration value, environment variable, or secret was absent or had an unexpected value for
- [Container exited unexpectedly](../deploy/container-crash.md) — The container started but then exited with a non-zero status code.
- [Service health check failure](../deploy/health-check-failure.md) — The new service instance never became healthy enough for traffic.
- [Container image pull failure](../deploy/image-pull-backoff.md) — The deployment failed because the runtime could not pull the requested container image from the registry.
- [Kubernetes pod in CrashLoopBackOff](../deploy/k8s-crashloopbackoff.md) — A Kubernetes pod is stuck in CrashLoopBackOff: the container starts, exits with a non-zero code, and Kubernetes restarts it with exponential back-off.
- [MySQL connection refused or service unavailable](../deploy/mysql-connection-refused.md) — The application or test suite could not connect to MySQL.
- [Port already in use during deployment](../deploy/port-conflict.md) — A container or service failed to start because the port it needs to bind to is already occupied by another
- [PostgreSQL connection refused or service unavailable](../deploy/postgres-connection-refused.md) — The application could not connect to PostgreSQL.
- [Redis connection refused or service unavailable](../deploy/redis-connection-refused.md) — The application or test suite could not connect to Redis.
- [Terraform initialization failure](../deploy/terraform-init.md) — `terraform init` failed before any plan or apply could begin.
- [Terraform state lock conflict](../deploy/terraform-state-lock.md) — Terraform could not acquire the state lock because another Terraform process is currently holding it.

## Network (11)

- [Service or dependency connection refused](../network/connection-refused.md) — A TCP connection attempt was actively refused by the target host.
- [DNS resolution failed ENOTFOUND](../network/dns-enotfound.md) — DNS resolution failed for a hostname.
- [DNS resolution failure](../network/dns-resolution.md) — The runner could not resolve a hostname to an IP address, so the network call failed before it could connect.
- [Outbound network traffic blocked by firewall or security group](../network/firewall-egress-blocked.md) — An outbound network connection was blocked before reaching its destination.
- [IPv6 vs IPv4 DNS resolution failure in CI container](../network/ipv6-ipv4-resolution.md) — CI container or test environment fails to connect because the system resolves a hostname to an IPv6 address (AAAA record)
- [LDAP server unreachable](../network/ldap-connection-failure.md) — The LDAP client could not reach the directory server.
- [Network request timeout](../network/network-timeout.md) — A network request exceeded its timeout limit.
- [Corporate or CI proxy misconfiguration blocking outbound traffic](../network/proxy-configuration.md) — Outbound HTTP(S) traffic from CI fails because the runner is behind a corporate proxy that is not configured in the
- [Request rate-limited by external service](../network/rate-limited.md) — A request to an external service was rejected with HTTP 429 Too Many Requests
or an equivalent rate-limit error.
- [Package registry or CDN outage](../network/registry-outage.md) — The dependency installation step failed because the upstream package registry
(npm, PyPI, Maven Central, RubyGems, etc.
- [SSL/TLS certificate error](../network/ssl-cert-error.md) — A TLS/SSL certificate error prevented a secure connection.

## Runtime (17)

- [CPU architecture or platform mismatch](../runtime/arch-mismatch.md) — A binary, container image, or build artifact targets the wrong CPU architecture for the current runner.
- [Disk space exhausted](../runtime/disk-full.md) — The host or container filesystem ran out of disk space.
- [Docker daemon configuration conflict prevents startup](../runtime/docker-daemon-config-conflict.md) — The Docker daemon failed to start because conflicting options were supplied — the same setting was configured both in `daemon.
- [Docker daemon unavailable](../runtime/docker-daemon-unavailable.md) — Docker could not reach the daemon socket or service, so the step failed before it could inspect containers or run a build.
- [Docker permission denied running as non-root user](../runtime/docker-permission-denied-nonroot.md) — A Docker container process running as a non-root user encountered a permission denied error while trying to access files or
- [Docker entrypoint or CMD misconfiguration](../runtime/entrypoint-misconfigured.md) — A Docker container exits immediately after start because its `ENTRYPOINT`
or `CMD` is misconfigured.
- [Required environment variable missing or empty](../runtime/env-var-missing.md) — A required environment variable was not set or was empty when the process tried to read it.
- [Error return value ignored or silently discarded](../runtime/missing-error-propagation.md) — A function returns an error that is discarded with `_ =` or left unchecked, meaning failures in I/O, database, or
- [Process killed by OOM killer](../runtime/oom-killed.md) — The process was terminated by the operating system's out-of-memory (OOM) killer.
- [Unrecovered panic in HTTP or RPC handler](../runtime/panic-in-http-handler.md) — A `panic` call appears inside an HTTP or RPC handler function without a corresponding `recover()` in a deferred function.
- [Permission denied](../runtime/permission-denied.md) — The process tried to read, write, execute, or connect to a resource it does not have permission to access.
- [Port already in use](../runtime/port-in-use.md) — A process attempted to bind to a TCP or UDP port that is already occupied by another process on the same CI runner.
- [Process killed without diagnostic output](../runtime/process-killed-no-logs.md) — A CI process was killed externally, producing little or no diagnostic output.
- [Resource limit exceeded](../runtime/resource-limits.md) — The process exceeded an OS-enforced resource limit such as the maximum number of threads or processes.
- [Process terminated by segmentation fault or fatal signal](../runtime/segfault.md) — The process received a fatal signal — most commonly SIGSEGV (segmentation fault), SIGABRT (assertion failure), or SIGBUS (bus error) —
- [Unawaited promise in async JavaScript or TypeScript code](../runtime/unawaited-promise.md) — An async JavaScript or TypeScript function starts a promise-returning operation but never awaits or catches it.
- [Docker volume mount failure or inaccessible mount](../runtime/volume-mount-issue.md) — A Docker container failed to start or operate correctly because a volume
or bind mount is inaccessible.

## Silent Failure (10)

- [Artifact upload step ran but no files were found](../silent-failure/artifact-missing.md) — An artifact upload step was executed but reported that no files matched the
configured path.
- [Cache restore or save failed but job continued](../silent-failure/cache-miss-non-fatal.md) — A cache restore or save step failed (cache not found, restore error, or save
error), but the CI job continued without flagging this as a failure.
- [CI step configured to continue on error](../silent-failure/continue-on-error.md) — A CI step uses `continue-on-error: true` or `allow_failure: true`, which
causes the pipeline to proceed even when the step exits non-zero.
- [Critical CI step allowed to fail](../silent-failure/continue-on-error-critical-step.md) — A CI workflow marks a critical build, test, deploy, artifact, or security step with `continue-on-error: true`, so the workflow can
- [Deploy or publish step ran but had nothing deployable](../silent-failure/empty-deployment-target.md) — A deploy, apply, publish, or release step ran, but the log indicates there
was nothing deployable to apply or publish.
- [Lint, scan, or coverage step ran but checked nothing](../silent-failure/empty-quality-check.md) — A lint, security-scan, or coverage step ran, but the log shows it processed
zero files or produced no meaningful report.
- [Command failure suppressed with exit-code override](../silent-failure/ignored-exit-code.md) — A failing command's exit code was deliberately suppressed using `|| true`,
`set +e`, or similar shell constructs.
- [CI command failure hidden by shell exit-code swallowing](../silent-failure/ignored-shell-exit-in-ci.md) — A CI script or workflow swallows a critical shell command failure with `|| true` or `set +e`, allowing the job to pass after important work failed.
- [Critical CI step skipped due to condition](../silent-failure/skipped-critical-step.md) — A CI step was skipped because its `if:` condition evaluated to false.
- [Test command ran but zero tests were executed](../silent-failure/zero-tests-executed.md) — A test runner command was invoked but discovered or executed zero tests.

## Test (24)

- [Cargo test suite reported one or more test failures](../test/cargo-test-failure.md) — One or more Rust tests failed.
- [Generic coverage gate failure](../test/coverage-gate-failure.md) — The test step completed, but the job failed because measured coverage fell below an enforced minimum.
- [Cucumber/Gherkin scenario step failed](../test/cucumber-step-failure.md) — One or more Cucumber scenarios failed.
- [Database migration timeout or lock contention](../test/database-migration-timeout.md) — Database migration failed to complete within the timeout window, usually due to lock contention or expensive DDL operations during test
- [Test database state pollution between tests](../test/database-test-isolation.md) — Tests are polluting a shared database with committed data that later tests
did not expect to find.
- [Flaky test failure](../test/flaky-test.md) — A test failed but the failure is not consistently reproducible.
- [Go race detector found a data race](../test/go-data-race.md) — The Go race detector identified a data race: two goroutines accessed the same memory location concurrently, with at least one
- [Go test suite reported one or more test failures](../test/go-test-failure.md) — One or more Go test functions failed or panicked.
- [jest command not found or not installed](../test/jest-command-not-found.md) — The `jest` command could not be found.
- [Jest test suite reported one or more test failures](../test/jest-test-failure.md) — One or more Jest tests failed.
- [Jest worker process crashed unexpectedly](../test/jest-worker-crash.md) — A Jest worker process crashed mid-test run.
- [JUnit / SBT test suite reported one or more test failures](../test/junit-test-failure.md) — One or more JVM test cases failed.
- [Missing test fixture or seed data](../test/missing-test-fixture.md) — A test failed because it could not load a required fixture file, seed record, or test helper resource that was expected to exist before the test ran.
- [Test order dependency](../test/order-dependency.md) — A test failed because it depends on side effects or state created by a different test that ran earlier.
- [Test parallelism conflict](../test/parallelism-conflict.md) — Parallel test execution caused a resource conflict: two tests tried to bind the same port, access the same fixture, or
- [PostgreSQL ALTER TYPE ADD VALUE cannot run inside a transaction block](../test/postgres-enum-add-in-transaction.md) — A database migration that adds a value to a PostgreSQL ENUM type fails because
PostgreSQL does not permit ALTER TYPE .
- [PostgreSQL migration failed due to unsupported type or constraint syntax](../test/postgres-version-incompatible-migration.md) — A database migration failed because the PostgreSQL version in CI does not
support a type or constraint syntax used in the migration (e.
- [pytest fixture setup or teardown failure](../test/pytest-fixture-error.md) — A pytest fixture failed during setup or teardown, causing the associated test to be marked as `ERROR` rather than `FAILED`.
- [pytest collected zero tests](../test/pytest-no-tests.md) — pytest exited with code 5 because it collected zero tests.
- [Snapshot or golden-file mismatch](../test/snapshot-mismatch.md) — A test compared current output against a committed snapshot or golden file and found a difference.
- [Test assertion failure with explicit reason message](../test/test-assertion-with-reason.md) — A test case assertion failed with an explicit reason message.
- [Test suite or individual test timed out](../test/test-timeout.md) — A test (or the entire test suite) did not complete within the configured timeout period.
- [Testcontainers container failed to start](../test/testcontainer-startup.md) — A test using Testcontainers could not start the required Docker container.
- [Timezone differences causing test failures](../test/timezone-diff.md) — A test that compares date or time values fails because the CI runner operates
in a different timezone than the development machine.



---

*Generated from `playbooks/bundled/`. Do not edit directly — run `make docs-generate`.*
