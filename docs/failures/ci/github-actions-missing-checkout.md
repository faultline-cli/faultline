# GitHub Actions missing actions/checkout before accessing repo files

**Playbook ID:** `github-actions-missing-checkout`
**Category:** ci
**Severity:** high
**Tags:** `github-actions`, `ci`, `checkout`, `workflow`

## What this failure means

GitHub Actions job tried to access repository files (build, test, run scripts) but the repository was never checked out. The `actions/checkout` step is missing or runs after the file access step.

## Common log signals

```text
actions/checkout
checkout step is missing
Did you forget to run actions/checkout
file not found in current directory
The process.*cannot find the file
```

## Diagnosis

`actions/checkout@v4` (or similar) must run before any steps that read or execute files from the repository. Common mistakes:

- The `actions/checkout` step is missing entirely.
- The `actions/checkout` step runs after the step that accesses repo files.
- The job runs on a self-hosted runner without a clean workspace.
- A previous workflow step deletes the checked-out files.

The error typically appears as `file not found`, `no such file or directory`, or commands failing to execute because they cannot find repository files.

## Fix steps

1. Ensure `actions/checkout` is the first step (or very early) in the job:

   ```yaml
   jobs:
     build:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4  # Add this as the FIRST step
         - uses: actions/setup-node@v3
           with:
             node-version: '18'
         - run: npm install
         - run: npm test
   ```

2. If the checkout is present but in the wrong place, move it before any file access steps.

3. If working with a monorepo, ensure you check out the correct path or use `fetch-depth` correctly:

   ```yaml
   - uses: actions/checkout@v4
     with:
       fetch-depth: 0  # For full history if needed
   ```

4. Verify the workspace is clean at the start of each job. Self-hosted runners may retain files from previous runs.

## Validation

- The workflow file contains `uses: actions/checkout@v4` (or newer version) as an early step.
- `ls -la` in the next step shows repository files in the runner's working directory.
- Re-run the GitHub Actions workflow.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain github-actions-missing-checkout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitHub Actions missing actions/checkout before accessing repo files
- Ci: github actions missing actions/checkout before accessing repo files
- Did you forget to run actions/checkout
- GitHub Actions github actions missing actions/checkout before accessing repo files
- faultline explain github-actions-missing-checkout


---

*Generated from [playbooks/bundled/log/ci/github-actions-missing-checkout.yaml](../../../playbooks/bundled/log/ci/github-actions-missing-checkout.yaml). Do not edit directly — run `make docs-generate`.*
