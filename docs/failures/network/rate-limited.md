# Request rate-limited by external service

**Playbook ID:** `rate-limited`
**Category:** network
**Severity:** medium
**Tags:** `network`, `rate-limit`, `429`, `npm`, `docker`, `github`, `api`

## What this failure means

A request to an external service was rejected with HTTP 429 Too Many Requests
or an equivalent rate-limit error. Package registries, container registries,
and APIs all enforce rate limits that CI can exceed.

## Common log signals

```text
429 Too Many Requests
rate limit exceeded
rate-limited
You have reached your pull rate limit
toomanyrequests:
Request rate limit exceeded
API rate limit exceeded
npm ERR! 429
```

## Diagnosis

The CI job made too many requests to an external service in a short window.
Common sources:

- npm or yarn pulling many packages in parallel without caching
- Docker Hub pull limits (anonymous: 100 pulls/6 h, authenticated: 200
  pulls/6 h)
- GitHub API calls from Actions steps or third-party actions
- PyPI or other package indexes under heavy concurrent load

## Fix steps

1. Identify which service is rate-limiting from the error message — the
   response usually includes a `Retry-After` or `X-RateLimit-Reset` header.

2. Add or restore a dependency cache so packages are not re-downloaded on
   every run:

   ```yaml
   # GitHub Actions example
   - uses: actions/cache@v4
     with:
       path: ~/.npm
       key: npm-${{ hashFiles('package-lock.json') }}
   ```

3. For Docker Hub: authenticate in CI to raise the pull limit from 100 to
   200 requests per 6 hours:

   ```bash
   docker login -u "$DOCKERHUB_USER" --password-stdin <<< "$DOCKERHUB_TOKEN"
   ```

   Better: use a registry mirror (GitHub Container Registry, AWS ECR Public)
   to avoid Docker Hub limits entirely.

4. For GitHub API: check that third-party actions are not making excessive
   API calls. Pass `GITHUB_TOKEN` explicitly so requests are authenticated
   and share the per-installation rate limit.

5. Retry transient rate-limit failures with exponential back-off rather than
   failing the job immediately — most registries recover within minutes.

## Validation

- Re-run the job after adding caching; confirm the download step skips
  previously cached packages.
- Confirm no HTTP 429 response appears in the job output.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain rate-limited
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Request rate-limited by external service
- Network: request rate-limited by external service
- npm warn registry Replying with an error
- faultline explain rate-limited
- npm request rate-limited by external service


---

*Generated from [playbooks/bundled/log/network/rate-limited.yaml](../../playbooks/bundled/log/network/rate-limited.yaml). Do not edit directly — run `make docs-generate`.*
