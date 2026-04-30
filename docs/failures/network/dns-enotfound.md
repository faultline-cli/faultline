# DNS resolution failed ENOTFOUND

**Playbook ID:** `dns-enotfound`
**Category:** network
**Severity:** high
**Tags:** `dns`, `network`, `resolution`, `enotfound`

## What this failure means

DNS resolution failed for a hostname. The domain name could not be resolved to an IP address, or there is no network connectivity.

## Common log signals

```text
ENOTFOUND
getaddrinfo ENOTFOUND
Name or service not known
nodename nor servname provided
DNS resolution
cannot resolve hostname
```

## Diagnosis

ENOTFOUND errors occur when:

- The hostname is mis-spelled or not resolvable in the current network.
- DNS servers are unreachable or not configured correctly on the CI runner.
- The private hostname is not registered in the local DNS or service discovery system (e.g., Kubernetes DNS).
- Network connectivity is missing or blocked by a firewall.
- The CI runner has no internet access (for public hostnames).

The error typically appears as `getaddrinfo ENOTFOUND <hostname>` or similar DNS resolution failures.

## Fix steps

1. Verify the hostname is spelled correctly and resolvable:

   ```bash
   nslookup <hostname>
   dig <hostname>
   curl -I https://<hostname>
   ```

2. Check if DNS is working at all:

   ```bash
   nslookup 8.8.8.8
   curl -I https://google.com
   ```

3. If the hostname is a private service (database, cache, etc.), verify:
   - The service is running and accessible.
   - The service name matches the configured hostname exactly.
   - On Kubernetes: the service is in the same namespace and reachable via DNS.
   - The CI environment has network access to the service.

4. Check the application's configuration for the correct hostname:

   ```bash
   grep -r "<hostname>" . --include="*.env" --include="*.yml" --include="*.yaml"
   ```

5. If DNS is misconfigured, update the application configuration to the correct hostname or service name.

## Validation

- `nslookup <hostname>` or `dig <hostname>` returns a valid IP address.
- `curl -I https://<hostname>` completes without DNS errors.
- Re-run the failing application or test.

## Likely files to inspect

- `.env`
- `.env.example`
- `.github/workflows/*.yml`
- `docker-compose.yml`
- `config.yaml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain dns-enotfound
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- DNS resolution failed ENOTFOUND
- Network: dns resolution failed enotfound
- nodename nor servname provided
- faultline explain dns-enotfound


---

*Generated from [playbooks/bundled/log/network/dns-enotfound.yaml](../../../playbooks/bundled/log/network/dns-enotfound.yaml). Do not edit directly — run `make docs-generate`.*
