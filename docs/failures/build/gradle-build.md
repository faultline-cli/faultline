# Gradle build failure

**Playbook ID:** `gradle-build`
**Category:** build
**Severity:** high
**Tags:** `gradle`, `java`, `kotlin`, `android`, `build`

## What this failure means

Gradle exited with a non-zero status, indicating one or more tasks failed. The `FAILURE:` block in the output identifies the first task that failed and the root cause.

## Common log signals

```text
FAILURE: Build failed with an exception
What went wrong:
Could not resolve
Could not download
Execution failed for task
Gradle build daemon
Could not GET
compileSdkVersion
```

## Diagnosis

Gradle exited with a non-zero status, indicating one or more tasks failed. The `FAILURE:` block in the output identifies the first task that failed and the root cause.

## Fix steps

1. Read the `What went wrong:` block. It usually contains the direct cause and file location.
2. For `Could not resolve`: check the repository URL and any private repo credentials.
3. For `Execution failed for task ':compileJava'`: fix the Java or Kotlin compilation errors listed below.
4. Run `./gradlew <task> --stacktrace` locally for the full exception chain.
5. For daemon issues, run `./gradlew --no-daemon` once to rule out cached state.

## Validation

- `./gradlew build --stacktrace` completes successfully.
- The CI build step exits zero without repeating the original `FAILURE:` block.

## Likely files to inspect

- `build.gradle`
- `build.gradle.kts`
- `settings.gradle`
- `gradlew`


## Run Faultline

```bash
faultline analyze build.log
faultline explain gradle-build
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Gradle build failure
- Build: gradle build failure
- FAILURE: Build failed with an exception
- GitHub Actions gradle build failure
- faultline explain gradle-build
- Gradle gradle build failure


---

*Generated from [playbooks/bundled/log/build/gradle-build.yaml](../../../playbooks/bundled/log/build/gradle-build.yaml). Do not edit directly — run `make docs-generate`.*
