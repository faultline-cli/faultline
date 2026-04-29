# Jenkins required plugin missing or incompatible

**Playbook ID:** `jenkins-plugin-missing`
**Category:** ci
**Severity:** high
**Tags:** `jenkins`, `plugin`, `dependency`, `ci`, `missing`

## What this failure means

A Jenkins job or pipeline step failed because a required plugin is not installed, is disabled, or has an incompatible version. Jenkins cannot execute the affected step.

## Common log signals

```text
No such DSL method
No such step
is not installed
Plugin .* is not installed
Required plugin
unable to find plugin
ClassNotFoundException.*plugin
Cannot resolve class
```

## Diagnosis

Jenkins pipelines use plugins to provide steps, node labels, and integration features. When a required plugin is absent or incompatible:

- The Jenkinsfile stage fails with `No such DSL method` or `No such step`.
- Jenkins logs `Plugin [name] is not installed` or `Required plugin [name] [version] is not satisfied`.
- A step that previously worked starts failing after a Jenkins upgrade changed plugin compatibility.
- The job configuration page shows a warning banner about a missing or inactive plugin.

## Fix steps

1. Identify the missing plugin from the error message, then install it via Jenkins → Manage Jenkins → Plugins → Available:

   ```groovy
   // Jenkinsfile step referencing a missing plugin will fail like:
   // ERROR: No such DSL method 'withCredentials' found
   // → Install the 'credentials-binding' plugin
   ```

2. After installing, restart Jenkins if prompted.

3. For version incompatibilities: check Manage Jenkins → Plugins → Installed for the affected plugin, and update it or pin it to a known-good version using the Plugin Manager's advanced settings.

4. For Configuration-as-Code or JCasC deployments, add the plugin to your `plugins.txt` or `installPlugins:` section in the CasC YAML.

5. Confirm plugin dependencies are satisfied — some plugins require other plugins at minimum version thresholds.

## Validation

- Manage Jenkins → Plugins → Installed shows the plugin as active.
- The previously failing Jenkinsfile step runs without a DSL or missing-plugin error.

## Likely files to inspect

- `Jenkinsfile`
- `plugins.txt`


## Run Faultline

```bash
faultline analyze build.log
faultline explain jenkins-plugin-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Jenkins required plugin missing or incompatible
- Ci: jenkins required plugin missing or incompatible
- ClassNotFoundException.*plugin
- GitHub Actions jenkins required plugin missing or incompatible
- faultline explain jenkins-plugin-missing


---

*Generated from [playbooks/bundled/log/ci/jenkins-plugin-missing.yaml](../../playbooks/bundled/log/ci/jenkins-plugin-missing.yaml). Do not edit directly — run `make docs-generate`.*
