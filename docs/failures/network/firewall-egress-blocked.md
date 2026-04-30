# Outbound network traffic blocked by firewall or security group

**Playbook ID:** `firewall-egress-blocked`
**Category:** network
**Severity:** high
**Tags:** `firewall`, `security-group`, `egress`, `network`, `blocked`

## What this failure means

An outbound network connection was blocked before reaching its destination. A firewall rule, VPC security group, or network policy denied the egress traffic from the CI runner.

## Common log signals

```text
Network is unreachable
ENETUNREACH
EHOSTUNREACH
Host is unreachable
getaddrinfo ENETUNREACH
no route to host
connect: network is unreachable
blocked by firewall
```

## Diagnosis

An outbound network connection was blocked before reaching its destination. A firewall rule, VPC security group, or network policy denied the egress traffic from the CI runner.

## Fix steps

1. Verify connectivity from the runner: `curl -v https://<host>` or `nc -zv <host> <port>`.
2. Add the required destination host and port to the egress rules of the security group, firewall policy, or network ACL.
3. If the runner is in a private VPC subnet: ensure a NAT Gateway or VPC endpoint (e.g., `com.amazonaws.*.s3`) is configured for the target service.
4. For self-hosted runners: check the host-level firewall (`iptables -L`, `ufw status`) for DROP rules that apply to the runner process user.

## Validation

- Re-run the local reproduction command after the fix.
- curl -v https://<destination-host>
- nc -zv <destination-host> <port>

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `terraform/`
- `infra/`


## Run Faultline

```bash
faultline analyze build.log
faultline explain firewall-egress-blocked
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Outbound network traffic blocked by firewall or security group
- Network: outbound network traffic blocked by firewall or security group
- Outbound connections are blocked
- faultline explain firewall-egress-blocked


---

*Generated from [playbooks/bundled/log/network/firewall-egress-blocked.yaml](../../../playbooks/bundled/log/network/firewall-egress-blocked.yaml). Do not edit directly — run `make docs-generate`.*
