# Network request timeout

**Playbook ID:** `network-timeout`
**Category:** network
**Severity:** medium
**Tags:** `network`, `timeout`, `http`, `connection`

## What this failure means

A network request exceeded its timeout limit. The remote host did not respond within the expected window, causing the CI step to fail.

## Common log signals

```text
connection timed out
read timeout
dial timeout
Get ".*": context deadline exceeded
context deadline exceeded while fetching
operation timed out
request timeout
ETIMEDOUT
```

## Diagnosis

A network request exceeded its timeout limit. The remote host did not respond within the expected window, causing the CI step to fail.

## Fix steps

1. Identify the exact host and command that timed out — registry, API, artifact store, or internal service — so you know where to look.
2. Test connectivity directly from the runner with a short timeout:

   ```bash
   curl -v --max-time 10 <URL>
   nc -zv <host> <port>
   ```

3. Retry the failed job once to separate a transient blip from a persistent routing or service problem.
4. Check the service status page and confirm the endpoint is healthy before adjusting timeout settings.
5. Verify the CI runner can reach the required host through any proxy, firewall, or egress policy applied in that environment.
6. Increase client-side timeout settings only when the endpoint is confirmed healthy and the operation is legitimately large or slow.

## Validation

- `curl -v --max-time 10 <URL>` completes within 10 seconds from the runner.
- Re-run the failing step and confirm it completes without a timeout error.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain network-timeout
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Network request timeout
- Network: network request timeout
- context deadline exceeded while fetching
- faultline explain network-timeout


---

*Generated from [playbooks/bundled/log/network/network-timeout.yaml](../../playbooks/bundled/log/network/network-timeout.yaml). Do not edit directly — run `make docs-generate`.*
