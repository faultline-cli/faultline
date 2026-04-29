# GitLab CI job log size limit exceeded

**Playbook ID:** `gitlab-job-log-limit`
**Category:** ci
**Severity:** medium
**Tags:** `gitlab`, `ci`, `log`, `limit`, `truncated`, `verbose`

## What this failure means

The GitLab CI job produced more output than the runner's configured log size limit. GitLab truncates the log and marks the job as failed.

## Common log signals

```text
Job's log exceeded limit
job log size exceeded
Archiving trace
log limit exceeded
truncated log
CI_JOB_LOG_LIMIT
trace size
job trace exceeded
```

## Diagnosis

GitLab enforces a maximum log size per job (default 100 MB for self-managed instances; GitLab.com uses a lower limit). When the job output reaches that ceiling, the runner stops archiving the trace, truncates the log, and fails the job.

Common causes:

- A verbose build tool printing every dependency download (Maven, Gradle, npm with `--verbose`).
- A test suite printing full diff output or stack traces for a large number of failures.
- A debug flag (`set -x`, `CI_DEBUG_TRACE=true`, or `--log-level debug`) left enabled in production workflows.
- An infinite loop or very long-running command that continuously writes to stdout.
- Unnecessary progress bars or animated output captured as static log lines.

## Fix steps

1. Identify which step is generating volume: look at the truncated log for the start of the last heavy section.

2. Reduce verbosity:
   - Maven: add `-q` (quiet) or remove `-X` (debug) from the `mvn` invocation.
   - Gradle: remove `--info` or `--debug`; add `--quiet` for builds that pass.
   - npm: remove `--verbose`; use `npm ci --prefer-offline 2>&1 | grep -v "^npm http"` to suppress download noise.

3. For test output, configure the test reporter to print only failures:

   ```yaml
   # pytest: show only failures
   pytest -q --tb=short
   # Jest: suppress pass summaries
   jest --silent
   ```

4. Remove `CI_DEBUG_TRACE: "true"` or `set -x` from the job configuration.

5. For self-managed GitLab, raise the log size limit in the runner configuration (`max_log_size`) or in the GitLab admin panel under Settings → CI/CD → Maximum job log file size.

6. Split the noisy job into smaller stages so each stage's output fits within the limit.

## Validation

- Re-run the job and confirm it completes without the log-limit message.
- `wc -c` on a captured log file should be well below the runner's configured limit.

## Likely files to inspect

- `.gitlab-ci.yml`
- `.gitlab/ci/*.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain gitlab-job-log-limit
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- GitLab CI job log size limit exceeded
- Ci: gitlab ci job log size limit exceeded
- Job's log exceeded limit
- GitHub Actions gitlab ci job log size limit exceeded
- faultline explain gitlab-job-log-limit


---

*Generated from [playbooks/bundled/log/ci/gitlab-job-log-limit.yaml](../../playbooks/bundled/log/ci/gitlab-job-log-limit.yaml). Do not edit directly — run `make docs-generate`.*
