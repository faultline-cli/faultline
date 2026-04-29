# Required configuration file not found

**Playbook ID:** `config-file-missing`
**Category:** build
**Severity:** high
**Tags:** `config`, `missing`, `file`, `not-found`, `setup`

## What this failure means

A required configuration file is absent from the CI workspace. The file
either was not committed to the repository, was excluded by `.gitignore`,
or was expected to be generated or templated at an earlier step that did not
run correctly.

## Common log signals

```text
config file not found
configuration file.*not found
No such file or directory.*config
Cannot find.*config
Could not find config
Missing configuration
config.*does not exist
.env.*not found
```

## Diagnosis

Configuration files are frequently absent in CI for one of these reasons:
1. The file is in `.gitignore` (common for `.env`, `*.local`, `secrets.yaml`)
2. The file is environment-specific and was not generated in a prior CI step
3. A required copy or template step was skipped or executed with the wrong
   working directory
4. The file exists locally but was never committed

Identify which file is missing:

```bash
# Check gitignore status
git check-ignore -v path/to/config.file

# List all tracked config files
git ls-files | grep -E '\.ya?ml$|\.json$|\.env'
```

## Fix steps

1. If the file contains secrets (API keys, passwords), do not commit it.
   Instead:
   - Create a template `.env.example` or `config.template.yaml` with
     placeholder values and commit that.
   - Generate the real file in CI from secrets:

     ```yaml
     # GitHub Actions
     - name: Create config
       run: |
         cat > config/app.yaml << EOF
         database_url: ${{ secrets.DATABASE_URL }}
         api_key: ${{ secrets.API_KEY }}
         EOF
     ```

2. If the file does not contain secrets, commit it:

   ```bash
   git add config/app.yaml
   git commit -m "chore: add application configuration"
   git push
   ```

3. For generated configs, ensure the generation step runs before the step
   that depends on it, and verify the step's working directory is correct.

4. For `.env` files, use a CI-native approach:
   - Copy from a template: `cp .env.example .env`
   - Populate from CI environment variables via a script

5. If using a config framework (Spring, Ruby on Rails, Django settings),
   verify the environment-specific config file is present for `CI` or
   `test` environments.

## Validation

- Re-run the failing job.
- Confirm `ls -la <config-path>` shows the file exists before the step
  that requires it.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain config-file-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Required configuration file not found
- Build: required configuration file not found
- No such file or directory.*config
- GitHub Actions required configuration file not found
- faultline explain config-file-missing


---

*Generated from [playbooks/bundled/log/build/config-file-missing.yaml](../../playbooks/bundled/log/build/config-file-missing.yaml). Do not edit directly — run `make docs-generate`.*
