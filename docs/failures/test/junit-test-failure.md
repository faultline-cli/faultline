# JUnit / SBT test suite reported one or more test failures

**Playbook ID:** `junit-test-failure`
**Category:** test
**Severity:** high
**Tags:** `java`, `jvm`, `scala`, `junit`, `sbt`, `gradle`, `test`, `assertion`

## What this failure means

One or more JVM test cases failed. The JUnit runner or SBT reported the failure
with a `java.lang.AssertionError`, `sbt.TestsFailedException`, or a
Gradle/SBT test task summary showing failed tests.

## Common log signals

```text
java.lang.AssertionError
org.junit.Assert
org.junit.ComparisonFailure
AssertionFailedError
sbt.TestsFailedException
Tests unsuccessful
test tasks failed:
testOnly failed:
```

## Diagnosis

A JUnit, SBT, or Gradle test run terminated with a non-zero exit code. Common
output formats seen in CI:

**Gradle / JUnit 4/5:**
```
com.example.MyTest > shouldDoSomething FAILED
    java.lang.AssertionError: expected:<1> but was:<2>
        at org.junit.Assert.fail(Assert.java:88)
```

**SBT:**
```
[error] 1 of 17 test tasks failed:
[error] - it:testOnly failed: sbt.TestsFailedException: Tests unsuccessful
```

Common causes:

- **Assertion failure** — a test assertion (`assertEquals`, `assertTrue`,
  `assertThat`) did not hold; actual output diverged from expected.
- **Memory/resource assertion** — a test that tracks heap growth or object
  counts fails because GC did not reclaim expected memory in time (common
  in reactive/observable library tests like RxJava's `publishNoLeak`).
- **NullPointerException** — a `null` reached a code path that did not
  expect it, typically caused by missing test setup or a changed API.
- **Binary compatibility failure** — a MiMa (`mimaReportBinaryIssues`) task
  detected that a public API signature changed in a binary-incompatible way.
- **Data-dependent failure** — the test relies on external state (database,
  filesystem, network) that is absent or inconsistent in CI.

## Fix steps

1. Find the failing test in the output — look for `FAILED` lines or the
   stack trace that follows `AssertionError`.

2. Reproduce locally:

   ```bash
   # Gradle
   ./gradlew test --tests "com.example.MyTest.shouldDoSomething"

   # SBT
   sbt "testOnly com.example.MySpec"

   # Maven
   mvn -pl module test -Dtest=MyTest#testMethod
   ```

3. For `AssertionError: X -> Y` memory tests (e.g. RxJava `publishNoLeak`):
   - Run the test in isolation with `-XX:+ExplicitGCInvokesConcurrent` or
     add an explicit `System.gc()` call.
   - Check whether the test is inherently flaky (non-deterministic GC) and
     whether the project has a known skip annotation for such tests.

4. For `sbt.TestsFailedException`:
   - Run `sbt testOnly *FailingSpec` and examine the output.
   - Look for `[error]` lines above the exception for the actual test name
     and assertion message.

5. For MiMa binary compatibility (`mimaReportBinaryIssues`):
   - Run `sbt mimaReportBinaryIssues` locally.
   - Add a `ProblemFilters.exclude[...]` entry in `build.sbt` if the API
     change is intentional and the next release is major/minor.

6. Fix the root cause: update the implementation to match the test, or — if
   the test expectation is wrong — update the test and document the change.

## Validation

- `./gradlew test` (or `sbt test` / `mvn test`) exits zero with no failures.
- The specific failing test name no longer appears in `FAILED` output.

## Likely files to inspect

- `build.sbt`
- `build.gradle`
- `build.gradle.kts`
- `pom.xml`
- `src/test/**/*.java`
- `src/test/**/*.scala`


## Run Faultline

```bash
faultline analyze build.log
faultline explain junit-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- JUnit / SBT test suite reported one or more test failures
- Test: junit / sbt test suite reported one or more test failures
- java.lang.IllegalArgumentException
- faultline explain junit-test-failure
- Java junit / sbt test suite reported one or more test failures


---

*Generated from [playbooks/bundled/log/test/junit-test-failure.yaml](../../playbooks/bundled/log/test/junit-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
