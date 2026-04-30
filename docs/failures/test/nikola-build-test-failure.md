# Nikola static site generator integration test failure

**Playbook ID:** `nikola-build-test-failure`
**Category:** test
**Severity:** high
**Tags:** `nikola`, `python`, `static-site`, `integration-test`, `sitemap`, `html`

## What this failure means

An integration test in the Nikola static site generator test suite failed.
Nikola's tests build a full site fixture and assert properties of the output
HTML and sitemap. Failures indicate a template rendering regression, a broken
URL scheme, or a missing output file.

## Common log signals

```text
tests.test_integration.TestCheckFailure
tests.test_integration.TranslatedBuildTest
correct path is in sitemap
links in output/index.html are correct
future post is not present in the index and sitemap
```

## Diagnosis

Nikola's integration tests in `tests/test_integration.py` build a staging site
and inspect the resulting files:

- `TestCheckFailure` — exercises Nikola's `check` command which validates
  internal links and reports broken references. Fails when the generated HTML
  contains a link whose target does not exist in the output directory.
- `TranslatedBuildTest` — builds a multilingual site and verifies that
  translated posts appear in the correct output paths. Fails when the i18n
  pipeline changes URL generation or when a translation fixture is stale.
- Sitemap assertions — tests verify that `output/sitemap.xml` contains the
  expected canonical paths and that `output/index.html` contains correct
  relative links. Failures indicate a base URL or slug change in the text
  fixtures.

Common root causes:

- **Template change** — a Jinja2/Mako template produced a different link
  structure, breaking the exact-string assertions in the tests.
- **Python version incompatibility** — Nikola's i18n layer behaves differently
  between Python 2 and 3 for non-ASCII slugs.
- **Dependency version skew** — an upgrade to `lxml`, `docutils`, or `Pygments`
  changed the rendered HTML enough to break the expected-string checks.
- **Stale test fixture** — the demo site content in `tests/` was not updated
  after a template or plugin change.

## Fix steps

1. Install the development dependencies:
   ```bash
   pip install -e ".[extras,tests]"
   ```
2. Run the failing integration test class in isolation:
   ```bash
   python -m pytest tests/test_integration.py::TestCheckFailure -v
   python -m pytest tests/test_integration.py::TranslatedBuildTest -v
   ```
3. Inspect the generated output to see what changed:
   ```bash
   find tests/integration/output -name "*.html" | head -5
   cat tests/integration/output/index.html | grep href
   ```
4. Diff the generated sitemap against the expected content:
   ```bash
   diff tests/integration/output/sitemap.xml tests/integration/expected-sitemap.xml
   ```
5. Update the test fixtures if the change is intentional:
   ```bash
   python -m pytest tests/test_integration.py --update-fixtures
   ```

## Validation

- `python -m pytest tests/test_integration.py -v` all pass.
- `nikola check -l` reports no broken links on the test site.

## Likely files to inspect

- `nikola/nikola.py`
- `nikola/post.py`
- `nikola/templates/`
- `tests/test_integration.py`
- `tests/integration/conf.py`


## Run Faultline

```bash
faultline analyze build.log
faultline explain nikola-build-test-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Nikola static site generator integration test failure
- Test: nikola static site generator integration test failure
- future post is not present in the index and sitemap
- faultline explain nikola-build-test-failure
- Python nikola static site generator integration test failure


---

*Generated from [playbooks/bundled/log/test/nikola-build-test-failure.yaml](../../../playbooks/bundled/log/test/nikola-build-test-failure.yaml). Do not edit directly — run `make docs-generate`.*
