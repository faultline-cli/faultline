# Asciidoctor JBehave integration test failure

**Playbook ID:** `asciidoctor-jbehave-test-failure`
**Category:** test
**Severity:** medium
**Tags:** `asciidoctor`, `jbehave`, `java`, `documentation`, `integration-test`

## What this failure means

An Asciidoctor Java integration test written in JBehave (BDD scenario style) failed.
These tests cover attribute substitution, Java extension registration, CLI invocation,
and Asciidoctor class instantiation. A failure typically indicates a regression in
the core rendering pipeline or extension API.

## Common log signals

```text
WhenAttributesAreUsedInAsciidoctor
WhenJavaExtensionIsRegistered
WhenAnAsciidoctorClassIsInstantiated
WhenAsciidoctorIsCalledUsingCli
```

## Diagnosis

The asciidoctor-java project uses JBehave for behaviour-driven integration tests.
Failing scenario classes include:

- `WhenAttributesAreUsedInAsciidoctor` — tests attribute substitution in documents.
- `WhenJavaExtensionIsRegistered` — tests the extension registration API.
- `WhenAnAsciidoctorClassIsInstantiated` — tests basic library instantiation.
- `WhenAsciidoctorIsCalledUsingCli` — tests the command-line interface.

Common causes:

- **API change** — a public method signature or attribute name changed.
- **Extension loading failure** — a required Java extension is missing from the classpath.
- **Template regression** — an AsciiDoc template produces unexpected output.

## Fix steps

1. Run the failing JBehave story locally:
   ```bash
   mvn test -Dtest=WhenAttributesAreUsedInAsciidoctor
   ```
2. Check recent changes to the attribute substitution or extension API.
3. Inspect the JBehave report in `target/jbehave/` for the failing step.

## Validation

Re-run all JBehave integration stories:
```bash
mvn integration-test
```

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain asciidoctor-jbehave-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Asciidoctor JBehave integration test failure
- Test: asciidoctor jbehave integration test failure
- WhenAnAsciidoctorClassIsInstantiated
- faultline explain asciidoctor-jbehave-test-failure
- Java asciidoctor jbehave integration test failure


---

*Generated from [playbooks/bundled/log/test/asciidoctor-jbehave-test-failure.yaml](../../../playbooks/bundled/log/test/asciidoctor-jbehave-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
