# Maven or Gradle dependency resolution failure

**Playbook ID:** `maven-dependency-resolution`
**Category:** build
**Severity:** high
**Tags:** `java`, `maven`, `gradle`, `dependencies`, `repository`

## What this failure means

The Java build could not resolve one or more dependencies or plugins from the configured artifact repositories.

## Common log signals

```text
could not find artifact
failed to read artifact descriptor
PluginResolutionException
pluginresolutionexception
transfer failed for
sbt.ResolveException
unresolved dependency:
module.*was not found
```

## Diagnosis

The Java build could not resolve one or more dependencies or plugins from the configured artifact repositories.

## Fix steps

1. Verify the artifact coordinates and version in `pom.xml`, `build.gradle`, or version catalogs.
2. Check repository credentials in Maven settings or Gradle properties if the dependency is private.
3. Re-run the build with dependency refresh enabled, for example `mvn -U package` or `./gradlew --refresh-dependencies build`.
4. Confirm the repository URL is reachable from the CI runner.

## Validation

- Re-run the failing workflow step.
- Confirm the original dependency resolution error is gone.

## Likely files to inspect

- `pom.xml`
- `settings.xml`
- `build.gradle`
- `build.gradle.kts`


## Run Faultline

```bash
faultline analyze build.log
faultline explain maven-dependency-resolution
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Maven or Gradle dependency resolution failure
- Build: maven or gradle dependency resolution failure
- failed to read artifact descriptor
- GitHub Actions maven or gradle dependency resolution failure
- faultline explain maven-dependency-resolution
- Java maven or gradle dependency resolution failure


---

*Generated from [playbooks/bundled/log/build/maven-dependency-resolution.yaml](../../playbooks/bundled/log/build/maven-dependency-resolution.yaml). Do not edit directly — run `make docs-generate`.*
