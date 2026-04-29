# CircleCI config.yml validation error

**Playbook ID:** `circleci-config-validation`
**Category:** ci
**Severity:** high
**Tags:** `circleci`, `config`, `yaml`, `validation`, `ci`

## What this failure means

CircleCI rejected the `.circleci/config.yml` file because it contains a syntax error, references an undefined parameter or executor, or violates the CircleCI config schema. No jobs ran.

## Common log signals

```text
config compilation failed
Invalid config detected
config.yml is invalid
was not accepted by the CircleCI API
Error parsing config
Configuration error
value of type .* is not assignable to type
Unknown executor
```

## Diagnosis

CircleCI validates the config file before scheduling any jobs. When validation fails the pipeline is blocked immediately with an error such as:

- `Error: config compilation failed: ...`
- `Invalid config detected: ...`
- `was not accepted by the CircleCI API`
- An undefined anchor reference or YAML alias error.
- A required field (e.g., `working_directory`, `executor`, `docker`) is missing or mistyped.
- A version mismatch: `version: 2.1` features used in a `version: 2` file.

## Fix steps

1. Validate the config locally before pushing:

   ```bash
   circleci config validate .circleci/config.yml
   ```

2. Fix any YAML syntax errors (indentation, missing colons, invalid anchors).

3. Ensure all referenced executors, commands, and orbs are declared in the same file or imported via the `orbs:` stanza.

4. Check that the `version:` key is `2.1` if you are using orbs, parameters, or reusable commands.

5. Remove or qualify any deprecated keys flagged in the validation output.

6. For orb-related errors, confirm the orb namespace and version exist at `circleci.com/developer/orbs`.

## Validation

- `circleci config validate .circleci/config.yml` exits 0 with no errors.
- The pipeline triggers and at least one job is scheduled.

## Likely files to inspect

- `.circleci/config.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain circleci-config-validation
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- CircleCI config.yml validation error
- Ci: circleci config.yml validation error
- value of type .* is not assignable to type
- GitHub Actions circleci config.yml validation error
- faultline explain circleci-config-validation


---

*Generated from [playbooks/bundled/log/ci/circleci-config-validation.yaml](../../playbooks/bundled/log/ci/circleci-config-validation.yaml). Do not edit directly — run `make docs-generate`.*
