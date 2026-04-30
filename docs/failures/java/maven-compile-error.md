# Maven Java compilation failure in CI

**Target search query:** `maven COMPILATION ERROR duplicate class cannot find symbol CI`

## Error snippet

```text
[ERROR] COMPILATION ERROR :
[ERROR] /builds/example/app/src/test/java/com/example/ServiceTest.java:[1298,5] class com.example.ServiceTest.TestInnerClass is already defined in class com.example.ServiceTest
[ERROR] /builds/example/app/src/test/java/com/example/setup/ContainerInitializer.java:[25,8] duplicate class: com.example.setup.ContainerInitializer
[ERROR] /builds/example/app/src/test/java/com/example/TestController.java:[48,25] cannot find symbol
[ERROR] Failed to execute goal org.apache.maven.plugins:maven-compiler-plugin:3.14.0:testCompile (default-testCompile) on project server: Compilation failure: Compilation failure:
```

## What this error means

The `maven-compiler-plugin` invoked `javac` and it returned compile errors. Maven
wraps the javac output in `[ERROR] COMPILATION ERROR :` and exits non-zero.

### duplicate class / is already defined

The same class is declared more than once in the source roots Maven is compiling.
Common causes:

- A generated source file (e.g. Lombok, MapStruct, jOOQ) was checked into
  version control and the generator also writes it to `target/generated-sources`.
- A file was moved to a new package but the original was not deleted.
- A failed Git merge left both versions of a file.

### cannot find symbol

A method, field, or type referenced in the source no longer exists. Common causes:

- An annotation-processor model class changed (Lombok `@Data` field removed,
  MapStruct source type renamed) without updating the test.
- A dependency was bumped and removed or renamed a public API.
- Generated sources under `target/` are stale or missing (run `mvn generate-sources`).

## Fix steps

1. Clean and recompile to reproduce locally:

   ```bash
   mvn clean test-compile
   ```

2. For duplicate class errors, find all copies:

   ```bash
   find . -name "DuplicateClass.java" -not -path "*/target/*"
   ```

3. For cannot-find-symbol errors, check recent changes to the missing type:

   ```bash
   git log --oneline -10 -- src/main/java/com/example/ChangedClass.java
   ```

## How Faultline detects it

Faultline maps this failure to `maven-compile-error`.

Primary signals:

- `[ERROR] COMPILATION ERROR`
- `duplicate class:`
- `is already defined in class`
- `maven-compiler-plugin.*Compilation failure`

Run:

```bash
cat build.log | faultline analyze
```

## Related failures

- [Maven dependency resolution failure](../build/maven-dependency-resolution.md)
