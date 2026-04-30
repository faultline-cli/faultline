# IPv6 vs IPv4 DNS resolution failure in CI container

**Playbook ID:** `ipv6-ipv4-resolution`
**Category:** network
**Severity:** low
**Tags:** `ipv6`, `ipv4`, `dns`, `network`, `container`, `aaaa`, `connectivity`, `dual-stack`

## What this failure means

CI container or test environment fails to connect because the system
resolves a hostname to an IPv6 address (AAAA record) but the container
network does not support IPv6, or a service binds on `::1` (IPv6 loopback)
when the client connects over IPv4.

## Common log signals

```text
Network unreachable
connect: Network is unreachable
ECONNREFUSED.*::1
getaddrinfo ENOTFOUND
curl.*Could not resolve host
prefer-ipv4
prefer-ipv6
ADDRESS_FAMILY_NOT_SUPPORTED
```

## Diagnosis

Docker containers have IPv6 disabled by default unless explicitly enabled
in `daemon.json`. When a hostname resolves to an AAAA record first and the
container network offers no IPv6 route, connections fail immediately.

Common scenarios:
1. An integration test service binds on `0.0.0.0` but the test opens a
   connection to `::1` or `localhost` resolves to `::1`
2. A container DNS resolver returns AAAA records, the client tries IPv6,
   and the kernel returns `EAFNOSUPPORT`
3. A service listens on `tcp6` and the client connects IPv4-only

Investigate:

```bash
# Check current address families
docker run --rm alpine cat /etc/hosts
# ::1 vs 127.0.0.1 for localhost

# Check what address a hostname resolves to inside the container
docker exec <container> getent ahosts registry.npmjs.org

# Check kernel IPv6 support
docker exec <container> cat /proc/sys/net/ipv6/conf/all/disable_ipv6
```

## Fix steps

1. **Force IPv4 resolution** for tools that try IPv6 first:

   ```bash
   # curl
   curl -4 https://registry.npmjs.org

   # Node.js 17+ prefers IPv6 by default; revert to v4-first
   export NODE_OPTIONS="--dns-result-order=ipv4first"

   # Python
   import socket
   socket.setdefaulttimeout(10)
   # Or use paramiko/requests with source_address bound to an IPv4 addr
   ```

2. **Enable IPv6 in Docker** if services genuinely need it:

   ```json
   // /etc/docker/daemon.json
   {
     "ipv6": true,
     "fixed-cidr-v6": "2001:db8:1::/64"
   }
   ```

   ```bash
   sudo systemctl reload docker
   ```

3. **Bind services to `0.0.0.0` and not `::1`** to avoid IPv6-only sockets:

   ```go
   // BAD: Go net.Listen — ":8080" binds both IPv4 and IPv6 on most systems
   // but "[::]" or "[::1]" is IPv6-only
   ln, _ := net.Listen("tcp", "[::1]:8080")  // IPv6 only

   // GOOD: explicit dual-stack or IPv4
   ln, _ := net.Listen("tcp", "0.0.0.0:8080")
   ```

4. **Disable IPv6 resolution in the container** if not needed:

   ```bash
   # In the CI job, before running services
   echo "net.ipv6.conf.all.disable_ipv6 = 1" | sudo tee -a /etc/sysctl.conf
   sudo sysctl -p

   # Or update /etc/hosts to remove ::1 for localhost
   sudo sed -i '/^::1.*localhost/d' /etc/hosts
   ```

5. **GitHub Actions**: hosted runners have IPv6 disabled; if relying on
   Docker services, specify `127.0.0.1` over `localhost`:

   ```yaml
   services:
     redis:
       image: redis:7
       ports:
         - "127.0.0.1:6379:6379"   # explicit IPv4 bind
   ```

## Validation

- Run `getent ahosts <hostname>` inside the container to confirm IPv4
  resolution.
- Run `nc -zv 127.0.0.1 6379` to confirm the service is reachable on IPv4.
- Re-run the failing test/job.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain ipv6-ipv4-resolution
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- IPv6 vs IPv4 DNS resolution failure in CI container
- Network: ipv6 vs ipv4 dns resolution failure in ci container
- bind: cannot assign requested address
- faultline explain ipv6-ipv4-resolution


---

*Generated from [playbooks/bundled/log/network/ipv6-ipv4-resolution.yaml](../../playbooks/bundled/log/network/ipv6-ipv4-resolution.yaml). Do not edit directly — run `make docs-generate`.*
