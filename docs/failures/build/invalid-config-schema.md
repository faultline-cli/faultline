# Configuration file fails schema validation

**Playbook ID:** `invalid-config-schema`
**Category:** build
**Severity:** medium
**Tags:** `config`, `schema`, `validation`, `yaml`, `json`, `toml`

## What this failure means

A configuration file exists but fails validation because a required field is
missing, an unrecognized field is present, or a value does not match the
expected type or constraint. The tool or framework rejects the config and
the CI step fails immediately.

## Common log signals

```text
schema validation failed
invalid configuration
config.*invalid
unknown field
unexpected field
required field.*missing
field.*is required
invalid value
```

## Diagnosis

Config schema failures arise from:
1. A tool upgrade that added required fields or removed deprecated ones
2. A copy-paste error introducing a typo in a field name
3. A wrong type (string where integer expected, list where scalar expected)
4. Merging from a template without removing example-only placeholders
5. YAML indentation error that changes the structure

Identify the exact invalid portion from the error:

```bash
# Validate YAML syntax
python3 -c "import yaml, sys; yaml.safe_load(open(sys.argv[1]))" config.yaml

# Validate JSON
python3 -m json.tool config.json

# Validate against a schema if available
npx ajv validate -s schema.json -d config.json
```

## Fix steps

1. Read the full validation error, which usually names the invalid field
   and its location (e.g., `.database.port: expected integer, got string`).

2. Compare the failing config against the tool's current documentation or
   upgrade changelog to identify renamed, removed, or newly required fields.

3. Fix the specific field:
   - Remove unrecognized keys or rename them to the current field names
   - Correct the type (quote/unquote values as needed in YAML)
   - Add required fields with appropriate values

4. For YAML: indent with spaces (never tabs), validate with a linter:

   ```bash
   yamllint config.yaml
   ```

5. For tool-specific config formats, use the tool's own validation command:

   ```bash
   helm lint .          # Helm charts
   terraform validate   # Terraform
   docker compose config  # Compose files
   ansible-lint         # Ansible
   ```

6. Pin the tool version so config schema changes do not break unexpectedly:

   ```yaml
   # package.json
   "eslint": "8.57.0"  # not "^8" or "latest"
   ```

## Validation

- Run the tool's validation command (`--dry-run`, `validate`, `lint`) and
  confirm it exits 0.
- Re-run the failing CI step.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain invalid-config-schema
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Configuration file fails schema validation
- Build: configuration file fails schema validation
- additional properties not allowed
- GitHub Actions configuration file fails schema validation
- faultline explain invalid-config-schema


---

*Generated from [playbooks/bundled/log/build/invalid-config-schema.yaml](../../../playbooks/bundled/log/build/invalid-config-schema.yaml). Do not edit directly — run `make docs-generate`.*
