# Corporate or CI proxy misconfiguration blocking outbound traffic

**Playbook ID:** `proxy-configuration`
**Category:** network
**Severity:** medium
**Tags:** `proxy`, `http-proxy`, `https-proxy`, `no-proxy`, `corporate`, `firewall`, `mitm`, `ssl-intercept`

## What this failure means

Outbound HTTP(S) traffic from CI fails because the runner is behind a
corporate proxy that is not configured in the job environment, or because
the proxy intercepts TLS and the custom CA certificate is not trusted.

## Common log signals

```text
CONNECT tunnel failed
407 Proxy Authentication Required
407 Proxy Auth
tunneling socket
HTTP_PROXY
HTTPS_PROXY
NO_PROXY
SSL_CERT_FILE
```

## Diagnosis

Corporate proxies and network appliances often intercept TLS traffic by
re-signing it with a private CA. Tools fail if:

1. The proxy environment variables (`HTTP_PROXY`, `HTTPS_PROXY`) are not
   set, and the proxy requires explicit configuration (no transparent proxy).
2. The proxy CA certificate is not in the trust store used by the tool.
3. `NO_PROXY` is missing internal hosts, causing all traffic (including
   internal services) to route through the proxy.

Inspect the environment:

```bash
env | grep -i proxy
curl -v https://registry.npmjs.org 2>&1 | grep -i 407
```

## Fix steps

1. **Set proxy environment variables**:

   ```yaml
   # GitHub Actions / GitLab CI
   env:
     HTTP_PROXY: "http://proxy.corp.example.com:8080"
     HTTPS_PROXY: "http://proxy.corp.example.com:8080"
     NO_PROXY: "localhost,127.0.0.1,.corp.example.com,169.254.169.254"
   ```

   Some tools use lowercase variants; set both:

   ```bash
   export http_proxy="$HTTP_PROXY"
   export https_proxy="$HTTPS_PROXY"
   export no_proxy="$NO_PROXY"
   ```

2. **Trust the proxy CA certificate for TLS interception**:

   ```bash
   # Ubuntu/Debian — add the proxy CA to the system trust store
   sudo cp corp-proxy-ca.crt /usr/local/share/ca-certificates/
   sudo update-ca-certificates

   # Python (requests/pip)
   export REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

   # Node.js
   export NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/corp-proxy-ca.crt

   # Go (uses system trust store)
   # Ruby
   export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
   ```

3. **Configure tool-specific proxy settings** when env vars are ignored:

   ```bash
   # npm
   npm config set proxy http://proxy.corp.example.com:8080
   npm config set https-proxy http://proxy.corp.example.com:8080
   npm config set cafile /path/to/corp-proxy-ca.crt

   # git
   git config --global http.proxy http://proxy.corp.example.com:8080
   git config --global http.sslCAInfo /path/to/corp-proxy-ca.crt

   # Docker daemon (edit /etc/docker/daemon.json)
   {
     "proxies": {
       "http-proxy": "http://proxy.corp.example.com:8080",
       "https-proxy": "http://proxy.corp.example.com:8080",
       "no-proxy": "localhost,127.0.0.1"
     }
   }
   ```

4. **Store proxy settings as CI secrets** to avoid exposing credentials in
   logs:

   ```yaml
   env:
     HTTPS_PROXY: ${{ secrets.CORP_PROXY_URL }}
   ```

## Validation

- Run `curl -v https://registry.npmjs.org` in the CI job and confirm 200 OK.
- Re-run the failing build step.
- Verify `NO_PROXY` excludes internal registries to prevent routing them
  through the external proxy.

## Likely files to inspect

*(Not specified.)*


## Run Faultline

```bash
faultline analyze build.log
faultline explain proxy-configuration
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Corporate or CI proxy misconfiguration blocking outbound traffic
- Network: corporate or ci proxy misconfiguration blocking outbound traffic
- self signed certificate in certificate chain
- faultline explain proxy-configuration


---

*Generated from [playbooks/bundled/log/network/proxy-configuration.yaml](../../playbooks/bundled/log/network/proxy-configuration.yaml). Do not edit directly — run `make docs-generate`.*
