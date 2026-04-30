# Maven Java compilation failure

**Playbook ID:** `maven-compile-error`
**Category:** build
**Severity:** high
**Tags:** `java`, `maven`, `compile`, `build`, `javac`

## What this failure means

The Maven compiler plugin failed during compilation. Common causes are
duplicate class definitions, missing symbols from a recently changed API, or
generated sources that are no longer in sync with the code.

## Common log signals

```text
COMPILATION ERROR :
duplicate class:
is already defined in class
cannot find symbol
Compilation failure: Compilation failure:
Failed to execute goal org.apache.maven.plugins:maven-compiler-plugin
```

## Diagnosis

Maven surfaced a `COMPILATION ERROR` from the `maven-compiler-plugin`. The
javac errors underneath it identify the root cause:

- **duplicate class** — the same class is defined more than once, often because
  generated source directories are also committed, or a class was moved without
  removing the original.
- **is already defined in class** — an inner class or enum constant appears twice
  in the same compilation unit. Usually caused by a failed merge or a code
  generator run that appended to an existing file.
- **cannot find symbol** — a method, field, or type that the code references no
  longer exists. Frequently triggered by a Lombok, MapStruct, or other
  annotation-processor model change that happened in a different commit.

## Fix steps

### Duplicate class

1. Search for all copies of the class file:

   ```bash
   find . -name "DuplicateClass.java" -not -path "*/target/*"
   ```

2. Remove the stale copy or the generated source directory from `pom.xml`
   `<sourceDirectory>` / `<testSourceDirectory>`.

3. Clean the build cache:

   ```bash
   mvn clean compile
   ```

### Cannot find symbol

1. Check whether the missing symbol was removed or renamed in a recent commit:

   ```bash
   git log --oneline -5 -- path/to/ChangedClass.java
   ```

2. Re-run the annotation processor (Lombok, MapStruct, etc.) if the symbol is
   generated:

   ```bash
   mvn clean generate-sources compile
   ```

3. Verify that the IDE-generated `target/generated-sources` is not excluded from
   the Maven source root configuration.

## Validation

- `mvn clean test-compile` — must exit 0 with no `COMPILATION ERROR` lines.
- Re-run the CI job and confirm the `[ERROR] COMPILATION ERROR` line is absent.

## Likely files to inspect

- `pom.xml`
- `src/main/java/**/*.java`
- `src/test/java/**/*.java`


## Run Faultline

```bash
faultline analyze build.log
faultline explain maven-compile-error
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Maven Java compilation failure
- Build: maven java compilation failure
- Failed to execute goal org.apache.maven.plugins:maven-compiler-plugin
- GitHub Actions maven java compilation failure
- faultline explain maven-compile-error
- Java maven java compilation failure


---

*Generated from [playbooks/bundled/log/build/maven-compile-error.yaml](../../../playbooks/bundled/log/build/maven-compile-error.yaml). Do not edit directly — run `make docs-generate`.*
