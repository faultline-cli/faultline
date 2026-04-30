# GitHub Actions workflow validation or action reference error

**Playbook ID:** `github-actions-syntax`
**Category:** ci
**Severity:** high
**Tags:** `github-actions`, `yaml`, `syntax`, `workflow`, `ci`

## What this failure means

GitHub Actions rejected the workflow definition or could not resolve a referenced action, so the workflow could not validate cleanly before the intended job logic ran.

## Common log signals

```text
Invalid workflow file
You have an error in your yaml syntax
yaml: line
did not find expected key
mapping values are not allowed here
could not parse the workflow file
error reading
Unrecognized named-value
```

## Diagnosis

GitHub Actions failed while validating the workflow file or resolving a referenced action revision. Common causes include malformed YAML, invalid expressions, and action references that point to a missing tag, branch, or commit.

## Fix steps

1. Validate the workflow locally: `pip install yamllint && yamllint .github/workflows/*.yml`.
2. Use the GitHub Actions VS Code extension which provides schema validation in the editor.
3. Check the exact line number from the error message and inspect indentation carefully.
4. Ensure expressions like `${{ secrets.TOKEN }}` are quoted when they appear in string values.
5. Verify action references such as `uses: actions/checkout@v4` point to a real released ref and are not being truncated or templated into an invalid value.
6. Verify `on:` trigger syntax using the GitHub Actions documentation.

## Validation

- Re-run the local reproduction command after the fix.
- yamllint .github/workflows/
- actionlint

## Likely files to inspect

- `.github/workflows/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-syntax
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions workflow validation or action reference error
- Ci: github actions workflow validation or action reference error
- You have an error in your yaml syntax
- GitHub Actions github actions workflow validation or action reference error
- faultline explain github-actions-syntax


---

*Generated from [playbooks/bundled/log/ci/github-actions-syntax.yaml](../../../playbooks/bundled/log/ci/github-actions-syntax.yaml). Do not edit directly — run `make docs-generate`.*
