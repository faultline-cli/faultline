# CircleCI job killed — resource class memory limit exceeded

**Playbook ID:** `circleci-resource-class-oom`
**Category:** ci
**Severity:** critical
**Tags:** `circleci`, `oom`, `memory`, `resource-class`, `ci`

## What this failure means

The CircleCI job was killed because it exceeded the memory limit of the selected resource class. The process received SIGKILL or the container was OOM-killed by the kernel, and the job exited non-zero without a clear user-visible error.

## Common log signals

```text
exit code: 137
Out of memory: Kill process
OOM killer
Killed
Process killed
exit status 137
SIGKILL
java.lang.OutOfMemoryError
```

## Diagnosis

CircleCI jobs run in containers with a fixed memory ceiling determined by the resource class. When the job exceeds that ceiling, the kernel OOM killer terminates one or more processes. The failure typically appears as:

- A process exits with signal 9 (SIGKILL) or a non-zero code with no error message.
- The log ends abruptly or shows `Killed` without a stack trace.
- CircleCI emits `exit code: 137` (128 + SIGKILL).
- A test runner or build tool exits with `OOM Killer` or `Out of memory: Kill process`.

Common causes include: large in-memory test suites with too much parallelism, Node.js builds without a `--max-old-space-size` flag, Docker-in-Docker image pulls inside small containers, or Maven/Gradle builds that default to large JVM heap sizes.

## Fix steps

1. Upgrade the resource class in `.circleci/config.yml`:

   ```yaml
   jobs:
     build:
       resource_class: medium+   # or large, xlarge, 2xlarge
   ```

2. For Node.js builds, cap the V8 heap explicitly:

   ```yaml
   - run: NODE_OPTIONS=--max-old-space-size=2048 npm run build
   ```

3. For JVM-based builds (Maven, Gradle), set a bounded heap:

   ```bash
   MAVEN_OPTS="-Xmx1g" mvn package
   GRADLE_OPTS="-Xmx1g" gradle build
   ```

4. Reduce test parallelism so each worker uses less memory simultaneously:

   ```yaml
   parallelism: 2    # lower from 4 or 8
   ```

5. Split long test suites across separate jobs and use CircleCI's test splitting to keep each job within budget.

6. Add a memory monitoring step before the failing command:

   ```bash
   free -h
   ```

## Validation

- Re-run the job and confirm no `exit code: 137`, `Killed`, or OOM message appears.
- Verify `free -h` at job start shows adequate available memory for the workload.

## Likely files to inspect

- `.circleci/config.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain circleci-resource-class-oom
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CircleCI job killed — resource class memory limit exceeded
- Ci: circleci job killed — resource class memory limit exceeded
- Fatal error: CALL_AND_RETRY_LAST Allocation failed
- GitHub Actions circleci job killed — resource class memory limit exceeded
- faultline explain circleci-resource-class-oom


---

*Generated from [playbooks/bundled/log/ci/circleci-resource-class-oom.yaml](../../../playbooks/bundled/log/ci/circleci-resource-class-oom.yaml). Do not edit directly — run `make docs-generate`.*
