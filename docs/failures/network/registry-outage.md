# Package registry or CDN outage

**Playbook ID:** `registry-outage`
**Category:** network
**Severity:** medium
**Tags:** `registry`, `npm`, `pypi`, `maven`, `rubygems`, `outage`, `availability`

## What this failure means

The dependency installation step failed because the upstream package registry
(npm, PyPI, Maven Central, RubyGems, etc.) is temporarily unavailable or
returning errors. This is a transient infrastructure failure unrelated to
the project's code.

## Common log signals

```text
registry is down
npm ERR! network
registry.*unavailable
pypi.*down
npmjs.*down
maven.*unavailable
central.sonatype.*unavailable
rubygems.*unavailable
```

## Diagnosis

Registry outages are typically short-lived (minutes to hours) and affect
all projects using the registry simultaneously. The failure will resolve
itself once the registry recovers.

Check registry status:
- npm: https://status.npmjs.org/
- PyPI: https://status.python.org/
- Maven Central: https://status.maven.org/
- RubyGems: https://status.rubygems.org/

Identify the outage symptom:
```
npm ERR! network This is a problem related to network connectivity
error: Registry returned 503 for "https://pypi.org/simple/requests/"
Could not transfer artifact from/to central: 503 Service Unavailable
```

## Fix steps

1. **Immediate:** Re-run the CI job. Registry outages are often brief and
   the retry will succeed.

2. Configure automatic retries in your package manager:

   ```bash
   # npm — retry on transient errors
   npm ci --fetch-retry-mintimeout 20000 --fetch-retry-maxtimeout 120000 --fetch-retries 3

   # pip — retry with backoff
   pip install --retries 5 --timeout 60 -r requirements.txt
   ```

3. For production pipelines, configure a mirror or proxy registry that caches
   packages locally:
   - npm: Verdaccio or GitHub Packages as a proxy
   - PyPI: devpi or Google Artifact Registry
   - Maven: Nexus or Artifactory

4. Enable dependency caching in CI so subsequent runs do not need to reach
   the registry at all:

   ```yaml
   # GitHub Actions
   - uses: actions/setup-node@v4
     with:
       cache: 'npm'
   ```

5. For time-critical pipelines, maintain a vendor directory with committed
   dependencies as a fallback.

## Validation

- Check the registry status page to confirm recovery.
- Re-run the failing CI job.
- Confirm dependency installation completes without network errors.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain registry-outage
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Package registry or CDN outage
- Network: package registry or cdn outage
- unexpected response from.*registry
- faultline explain registry-outage
- npm package registry or cdn outage


---

*Generated from [playbooks/bundled/log/network/registry-outage.yaml](../../playbooks/bundled/log/network/registry-outage.yaml). Do not edit directly — run `make docs-generate`.*
