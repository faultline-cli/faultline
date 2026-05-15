# Faultline Failure Pages

These pages turn recurring CI failures into searchable, problem-first docs.

Each page is designed to win searches for exact CI error strings, explain the failure quickly, and show how Faultline maps the log to a deterministic diagnosis.

## Full failure catalog

**[docs/failures/catalog/README.md](catalog/README.md)** — all 182 bundled playbooks, generated directly from the playbook YAML. Every failure category is indexed with links to its dedicated diagnosis and fix page.

## Generating this catalog

The pages under `docs/failures/<category>/` and the files `catalog/README.md` and `llms.txt` are generated automatically from `playbooks/bundled/**/*.yaml`.

```bash
# Regenerate all pages after adding or modifying a playbook.
make docs-generate

# Verify pages are up to date (used in CI).
make docs-check
```

The generator lives at `tools/gen-failure-docs/main.go`. It reads every YAML file under `--src` (`playbooks/bundled/` by default), renders one Markdown page per playbook, writes the catalog index, and writes an `llms.txt` entry-point — all to `--dst` (`docs/failures/` by default).

Do not edit generated pages directly. Make changes in the corresponding YAML file under `playbooks/bundled/`, then run `make docs-generate` and commit both.

## Recommended content model

Keep each page narrowly scoped to one dominant search intent:

1. Put the exact error string in the title and code block.
2. Keep the log snippet minimal and real.
3. Explain the failure in one short paragraph.
4. Give the smallest fix sequence that gets CI green again.
5. Show the Faultline playbook ID and primary matching signals.
6. Cross-link to adjacent failures instead of bloating the page.

## Authoring template

Use [docs/failures/_template.md](_template.md) for new pages.
