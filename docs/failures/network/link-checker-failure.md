# Link checker found broken or redirecting URLs

**Playbook ID:** `link-checker-failure`
**Category:** network
**Severity:** medium
**Tags:** `link-checker`, `html-proofer`, `htmltest`, `documentation`, `urls`

## What this failure means

A link-checking tool (html-proofer, htmltest, or similar) found broken, missing,
or permanently redirecting URLs in the project's documentation or web output.

## Common log signals

```text
HTML-Proofer found
htmlproofer
htmltest
linkcheck
> Links
```

## Diagnosis

A CI step that validates documentation links found one or more failures. Common
output formats:

**HTML-Proofer:**
```
htmlproofer 3.x.x | Starting the crawl
- https://example.com/old-path - 301
HTML-Proofer found 3 failures!
```

**htmltest:**
```
> Links
  1. [L0239] 522 http://example.com/image.jpg
  2. [L1581] 404 https://example.com/missing-page
> Dupes
  None ✓
```

Common causes:

- **Moved content** — an external resource was permanently redirected (301) or
  removed (404/410). These were valid links when written but have since changed.
- **Rate-limited or unreachable** — external hosts return 5xx or 522 during CI
  runs due to rate limiting or transient unavailability.
- **Internal broken link** — an anchor, page path, or section heading was
  renamed without updating all references.
- **Image asset missing** — an image referenced in documentation was moved,
  renamed, or never committed.

## Fix steps

1. Review the list of failing URLs in the checker output.

2. For **301 redirects** — update the link in documentation to point to the
   canonical (redirected) destination URL.

3. For **404 / gone** — find an equivalent replacement page or remove the link.
   Web archives (archive.org) can help recover vanished content.

4. For **transient failures (5xx, timeout)** — consider adding the domain to
   the checker's ignore list if it rate-limits CI requests, or retry the job.

5. For **image assets** — verify the asset is committed and the path in the
   documentation matches the actual file location.

6. To run locally:
   ```bash
   # html-proofer (Ruby)
   htmlproofer ./_site --checks Links,Images,Scripts

   # htmltest (Go)
   htmltest

   # markdown-link-check
   find . -name "*.md" -exec markdown-link-check {} \;
   ```

## Validation

- The link checker exits zero with no failures reported.
- Run `htmlproofer ./_site` or `htmltest` locally to confirm all links resolve.

## Likely files to inspect

- `README.md`
- `docs/**/*.md`
- `_config.yml`
- `.htmltest.yml`
- `.htmlprooferrc`


## Run Faultline

```bash
faultline analyze build.log
faultline explain link-checker-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Link checker found broken or redirecting URLs
- Network: link checker found broken or redirecting urls
- HTML-Proofer found
- faultline explain link-checker-failure


---

*Generated from [playbooks/bundled/log/network/link-checker-failure.yaml](../../../playbooks/bundled/log/network/link-checker-failure.yaml). Do not edit directly — run `make docs-generate`.*
