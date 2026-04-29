# GitHub Actions matrix axis variable undefined

**Playbook ID:** `github-actions-matrix-axis-invalid`
**Category:** ci
**Severity:** high
**Tags:** `github-actions`, `matrix`, `workflow`, `ci`, `undefined-variable`

## What this failure means

A GitHub Actions workflow uses an undefined matrix axis variable in a step or action, causing the workflow to fail during variable resolution.

## Common log signals

```text
Unresolved variable
matrix.*not defined
matrix.*undefined
Unable to retrieve action metadata
```

## Diagnosis

GitHub Actions matrix jobs allow you to run the same workflow across multiple configurations. If a step references a matrix variable (e.g., `${{ matrix.node-version }}`) but that variable is not defined in the job's `matrix:` strategy, the workflow fails with an "Unresolved variable" error.

This is a workflow authoring error — the variable name was mistyped, the axis was never added to the matrix, or the reference syntax is incorrect.

Common causes:
- Typo in matrix axis name (e.g., `matrix.node` instead of `matrix.node-version`)
- Matrix axis removed from strategy but step still references it
- Copy-paste from different workflow with different matrix axes
- Incorrect template syntax (e.g., missing braces or dollar sign)

## Fix steps

1. Identify the undefined variable in the error message. It will say something like "Unresolved variable matrix.node-version".

2. Check your workflow's `matrix:` strategy to confirm the axis exists:

   ```yaml
   strategy:
     matrix:
       node-version: [16, 18, 20]
       os: [ubuntu-latest, macos-latest]
   ```

3. Fix the reference in the step to match exactly:

   ```yaml
   - name: Setup Node
     uses: actions/setup-node@v3
     with:
       node-version: ${{ matrix.node-version }}
   ```

4. Verify the axis name exists in the parent job's `matrix:` block and the syntax is correct.

5. Use `${{ matrix.<axis-name> }}` only for axes defined in the same job's matrix strategy. For variables from other sources, use `${{ env.VAR_NAME }}` or `${{ secrets.SECRET_NAME }}`.

## Validation

- Update the workflow file to use a valid matrix axis.
- Push to a feature branch and run the workflow again.
- Confirm the "Unresolved variables in workflow" error no longer appears.
- Verify the job runs successfully with a valid matrix axis.

## Likely files to inspect

- `.github/workflows/*.y*ml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-matrix-axis-invalid
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions matrix axis variable undefined
- Ci: github actions matrix axis variable undefined
- Unable to retrieve action metadata
- GitHub Actions github actions matrix axis variable undefined
- faultline explain github-actions-matrix-axis-invalid


---

*Generated from [playbooks/bundled/log/ci/github-actions-matrix-axis-invalid.yaml](../../playbooks/bundled/log/ci/github-actions-matrix-axis-invalid.yaml). Do not edit directly — run `make docs-generate`.*
