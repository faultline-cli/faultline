# DNS resolution failure

**Playbook ID:** `dns-resolution`
**Category:** network
**Severity:** medium
**Tags:** `dns`, `network`, `host`, `resolution`

## What this failure means

The runner could not resolve a hostname to an IP address, so the network call failed before it could connect.

## Common log signals

```text
no such host
temporary failure in name resolution
name or service not known
getaddrinfo enotfound
dial tcp: lookup
could not resolve host
nodename nor servname provided
```

## Diagnosis

The runner could not resolve a hostname to an IP address, so the network call failed before it could connect.

## Fix steps

1. Verify the hostname in the failing command or configuration for typos.
2. Manually resolve the host from inside the CI environment to confirm it is
   reachable:

   ```bash
   nslookup <hostname>
   dig +short <hostname>
   getent hosts <hostname>
   ```

   If DNS resolution works from the runner but the job still fails, the
   hostname may be coming from a misconfigured secret or environment variable.

3. Check whether the hostname is internal-only and requires VPN, VPC DNS, or
   a private resolver. CI runners on cloud providers may not have access to
   private DNS zones by default.
4. Retry the job once to rule out a transient resolver outage.
5. If the hostname comes from configuration, print the resolved value in the
   same job or container and confirm it matches the expected environment or
   secret before retrying.

## Validation

- `dig +short <hostname>` or `nslookup <hostname>` — confirm successful
  resolution from the runner before re-running the job.
- Re-run the failing step and confirm the DNS lookup no longer fails.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `docker-compose*.yml`
- `.env.example`


## Run Faultline

```bash
faultline analyze build.log
faultline explain dns-resolution
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- DNS resolution failure
- Network: dns resolution failure
- temporary failure in name resolution
- faultline explain dns-resolution


---

*Generated from [playbooks/bundled/log/network/dns-resolution.yaml](../../../playbooks/bundled/log/network/dns-resolution.yaml). Do not edit directly — run `make docs-generate`.*
