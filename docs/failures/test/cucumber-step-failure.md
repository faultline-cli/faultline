# Cucumber/Gherkin scenario step failed

**Playbook ID:** `cucumber-step-failure`
**Category:** test
**Severity:** medium
**Tags:** `cucumber`, `gherkin`, `bdd`, `ruby`, `acceptance-test`, `capybara`, `rspec`

## What this failure means

One or more Cucumber scenarios failed. The failing step is identified by its
feature file path and line number. Failures commonly indicate a broken
application flow, a missing environment dependency (browser driver, database,
external service), or a timing issue in a browser-based test.

## Common log signals

```text
.feature
Failing Scenarios:
cucumber features
capybara_features/
```

## Diagnosis

Cucumber runs acceptance tests written in Gherkin (`.feature` files). Each
scenario is a sequence of Given/When/Then steps; a single failing step fails
the whole scenario.

Common causes:

- **Application regression** — the feature behaviour changed and the scenario
  now fails because the expected UI element, response, or side-effect is absent.
- **Missing environment dependency** — a browser driver (chromedriver, geckodriver),
  database seed, external service mock, or required gem is absent or misconfigured
  in the CI environment.
- **Timing/flakiness** — a Capybara or Selenium wait timeout was hit because
  the page did not render fast enough on CI hardware.
- **Database state** — the database was not migrated, seeded, or reset between
  scenarios, producing unexpected results.
- **Parallel execution conflict** — in parallel Cucumber runs (`-p travis`),
  shared state caused scenario interference.

## Fix steps

1. Find the failing feature file and line number in the log output.
2. Run the failing scenario in isolation locally:

   ```bash
   bundle exec cucumber features/path/to/scenario.feature:42
   ```

3. If the failure is environment-related, check that required services are
   running and properly seeded:

   ```bash
   bundle exec rake db:test:prepare
   bundle exec rails server -e test &
   ```

4. For browser-based failures, confirm the expected browser driver is installed
   and matches the configured browser version:

   ```bash
   which chromedriver && chromedriver --version
   ```

5. For timeout failures, increase the Capybara wait time or check for slow
   CI-specific rendering.
6. If the failure is in the application logic, review the diff for changes
   to the affected feature area.

## Validation

- Run the failing scenario in isolation: `bundle exec cucumber <feature>:<line>`
- Run the full feature file: `bundle exec cucumber features/...feature`
- Confirm the integration test suite passes: `bundle exec rake cucumber`

## Likely files to inspect

- `features/**/*.feature`
- `features/**/*_steps.rb`
- `features/support/env.rb`
- `spec/support/**`
- `Gemfile`
- `.ruby-version`


## Run Faultline

```bash
faultline analyze build.log
faultline explain cucumber-step-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Cucumber/Gherkin scenario step failed
- Test: cucumber/gherkin scenario step failed
- Failing Scenarios:
- faultline explain cucumber-step-failure


---

*Generated from [playbooks/bundled/log/test/cucumber-step-failure.yaml](../../../playbooks/bundled/log/test/cucumber-step-failure.yaml). Do not edit directly — run `make docs-generate`.*
