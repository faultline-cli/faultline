# Read the Docs site-build integration test failure

**Playbook ID:** `readthedocs-build-test-failure`
**Category:** test
**Severity:** medium
**Tags:** `readthedocs`, `django`, `sphinx`, `documentation`, `build`, `integration-test`

## What this failure means

A Read the Docs (readthedocs.io) Django integration test failed.
These tests exercise the documentation build pipeline end-to-end — building
demo projects, verifying sitemaps, checking generated links, testing URL routing,
and validating translated content. A failure here usually means a regression in
the build system, URL configuration, or template rendering logic.

## Common log signals

```text
DemoBuildTest
EmptyBuildTest
Testmaker.test_
test_integration.TranslatedBuildTest
test_integration.TestCheck
```

## Diagnosis

The test suite for the Read the Docs application includes integration tests that
actually invoke the Sphinx/MkDocs build pipeline and inspect the output. Failing
test classes include:

- `DemoBuildTest` — verifies the demo documentation site builds correctly and
  produces correct RSS feeds, sitemaps, and indexes.
- `EmptyBuildTest` — verifies that an empty project builds without error.
- `Testmaker` / `TestMaker` — tests the `readthedocs-build` CLI tool for creating
  new projects.
- `TranslatedBuildTest` / `TranslationsPatternTest` — tests multilingual docs.
- `TestCheck` / `TestCheckFailure` — link-checker integration.
- `FuturePostTest`, `MonthlyArchiveTest`, `DayArchiveTest` — blog-style archive tests.
- `PrivacyTests`, `RedirectTests`, `SubdomainUrlTests` — URL routing and privacy
  gate tests.

Common causes:

- **Build dependency missing** — a required Sphinx extension, theme, or Python
  package is not installed in the test environment.
- **Template or URL regression** — a change to view logic, URL patterns, or
  a Jinja/Django template broke expected output.
- **File system fixture mismatch** — the test expects a specific file to be
  generated at a specific path, but the path changed.
- **Sitemap or RSS schema change** — the structured output format changed
  and the test assertions haven't been updated.

## Fix steps

1. Run the failing tests locally with verbose output to see the actual vs expected
   diff:
   ```bash
   python -m pytest readthedocs/ -k "DemoBuildTest or Testmaker" -v
   ```
2. For build failures, check that all Sphinx extensions and themes required by
   the test fixtures are installed:
   ```bash
   pip install -r requirements/testing.txt
   ```
3. For sitemap or link assertion failures, inspect the generated output file and
   compare it with the expected fixture.
4. For URL routing failures, run the URL resolver against the failing path:
   ```bash
   python manage.py shell -c "from django.urls import reverse; print(reverse('docs_detail', ...))"
   ```
5. Check recent commits to `readthedocs/templates/`, `readthedocs/urls.py`, or
   `readthedocs/builds/` for the change that broke the test.

## Validation

- Re-run the full integration test suite: `python -m pytest readthedocs/integrations/ -v`
- Confirm that `DemoBuildTest`, `TranslatedBuildTest`, and `TestCheck` all pass.

## Likely files to inspect

- `readthedocs/builds/`
- `readthedocs/templates/`
- `readthedocs/urls.py`
- `requirements/testing.txt`


## Run Faultline

```bash
faultline analyze build.log
faultline explain readthedocs-build-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Read the Docs site-build integration test failure
- Test: read the docs site-build integration test failure
- test_integration.TranslatedBuildTest
- faultline explain readthedocs-build-test-failure


---

*Generated from [playbooks/bundled/log/test/readthedocs-build-test-failure.yaml](../../../playbooks/bundled/log/test/readthedocs-build-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
