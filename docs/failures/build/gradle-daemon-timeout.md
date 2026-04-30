# Gradle daemon wrapper lock timeout

**Playbook ID:** `gradle-daemon-timeout`
**Category:** build
**Severity:** high
**Tags:** `gradle`, `maven`, `jvm`, `build`, `daemon`, `timeout`

## What this failure means

The Gradle or Maven daemon could not acquire an exclusive lock on cached wrapper files, causing the build tool initialization to timeout.

## Common log signals

```text
Timeout waiting for exclusive access to file
gradle.*daemon.*closed
connection to daemon lost
ERROR 500 in daemon process
Gradle daemon not started
```

## Diagnosis

Gradle and Maven use a daemon process for efficiency, but the daemon must acquire exclusive locks on its cached wrapper JAR files during initialization. When multiple builds run concurrently on the same runner (or a previous build wasn't fully cleaned up), the lock acquisition times out and the build fails.

This is distinct from a general pipeline timeout — the JVM is running but cannot proceed because of file system lock contention on cached build tool artifacts.

Common causes:
- Multiple concurrent builds on the same runner sharing cache directories
- Previous build daemon still holding locks (incomplete cleanup or crash)
- Stale lock file from interrupted build
- Runner filesystem performance degradation

## Fix steps

1. Clear the Gradle or Maven daemon and wrapper cache:

   ```bash
   rm -rf ~/.gradle/wrapper/dists ~/.gradle/caches ~/.m2/repository/org/gradle
   ```

2. Add a cleanup step at the start of your CI job to ensure no stale daemon processes are running:

   ```bash
   pkill -9 -f 'java.*gradle' || true
   pkill -9 -f 'java.*maven' || true
   sleep 2
   ```

3. If using matrix builds or parallel jobs on shared runners, ensure cache keys include the job ID to avoid lock contention:

   ```yaml
   cache:
     key: gradle-cache-${{ runner.os }}-${{ github.run_id }}-${{ matrix.java-version }}
     paths:
       - ~/.gradle
   ```

4. Increase the Gradle daemon startup timeout:

   ```bash
   export ORG_GRADLE_PROJECT_org_gradle_daemon_performance_warn_time_in_ms=60000
   ./gradlew build
   ```

## Validation

- Run `rm -rf ~/.gradle/wrapper/dists ~/.gradle/caches` and re-run the build.
- Confirm the build progresses past the "Starting a Gradle Daemon" message.
- Check that no `Timeout waiting for exclusive access` errors appear in the output.

## Likely files to inspect

- `gradle.properties`
- `build.gradle`
- `build.gradle.kts`
- `.github/workflows/*.y*ml`
- `settings.gradle`


## Run Faultline

```bash
faultline analyze build.log
faultline explain gradle-daemon-timeout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Gradle daemon wrapper lock timeout
- Build: gradle daemon wrapper lock timeout
- Timeout waiting for exclusive access to file
- GitHub Actions gradle daemon wrapper lock timeout
- faultline explain gradle-daemon-timeout
- Gradle gradle daemon wrapper lock timeout


---

*Generated from [playbooks/bundled/log/build/gradle-daemon-timeout.yaml](../../../playbooks/bundled/log/build/gradle-daemon-timeout.yaml). Do not edit directly — run `make docs-generate`.*
