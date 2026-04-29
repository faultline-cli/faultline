# Cross-ownership boundary failure

**Playbook ID:** `topology-boundary-crossed`
**Category:** build
**Severity:** high
**Tags:** `topology`, `monorepo`, `codeowners`, `boundary`, `cross-team`

## What this failure means

A change in one ownership zone has broken code in a separate ownership zone,
suggesting the failure originates at a cross-team integration boundary rather
than within a single component.

## Common log signals

```text
undefined:
ImportError:
cannot find module
package .* is not in GOROOT
Module not found:
does not implement
has no exported field or method
cannot use .* as type
```

## Diagnosis

Topology analysis detected that the recently changed files belong to a
different CODEOWNERS zone than the files producing the observed errors. This
pattern is characteristic of a cross-boundary regression:

- A shared library or internal API was modified.
- Downstream consumers in another team's codebase depend on the changed
  interface.
- The change was not coordinated with, or reviewed by, the owner of the
  consuming code.

Common root causes:
- **Unannounced interface change** — a function signature, type, or exported
  constant was altered without a matching update in dependent packages.
- **Missing cross-owner LGTM** — the CODEOWNERS file required review from the
  impacted owner, but the review was bypassed or the owner was not notified.
- **Implicit dependency** — the consuming code assumes undocumented
  implementation details that changed.

## Fix steps

1. Identify the ownership boundary:
   ```
   cat CODEOWNERS | grep -E "(the-changed-dir|the-failing-dir)"
   ```
2. Check whether the changed API or module is still backward-compatible:
   ```bash
   git diff HEAD~1 -- <changed-path>
   ```
3. Update the downstream consumer to match the new interface, or revert the
   upstream change until the consuming team can coordinate.
4. If the upstream change is intentional, open a PR that includes both the
   change and the downstream fix, requesting LGTM from the relevant owners.
5. Add a test at the boundary (integration or contract test) to catch future
   regressions.

## Validation

- Build succeeds end-to-end across both ownership zones.
- The CODEOWNERS-required reviewers have approved the combined change.
- A boundary test exists and passes in CI.

## Likely files to inspect

- `CODEOWNERS`
- `go.mod`
- `package.json`
- `pyproject.toml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain topology-boundary-crossed
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Cross-ownership boundary failure
- Build: cross-ownership boundary failure
- has no exported field or method
- GitHub Actions cross-ownership boundary failure
- faultline explain topology-boundary-crossed


---

*Generated from [playbooks/bundled/log/build/topology-boundary-crossed.yaml](../../playbooks/bundled/log/build/topology-boundary-crossed.yaml). Do not edit directly — run `make docs-generate`.*
