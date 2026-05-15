# TLS certificate verification disabled in HTTP client

**Playbook ID:** `insecure-tls-skip-verify`
**Category:** security
**Severity:** high
**Tags:** `source`, `security`, `tls`, `certificate`, `owasp`

## What this failure means

TLS certificate verification is explicitly disabled, allowing the client to
connect to servers with invalid, expired, or self-signed certificates without
raising an error.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A TLS configuration or HTTP client has certificate verification disabled.
This removes the protection TLS is designed to provide: without certificate
verification, any server can impersonate the intended target and all traffic
is vulnerable to interception.

Common patterns:
- Go: `&tls.Config{InsecureSkipVerify: true}` used in an `http.Transport`
- Python: `requests.get(url, verify=False)` or `session.verify = False`
- Node.js: `new https.Agent({ rejectUnauthorized: false })`

This often starts as a quick workaround for a self-signed certificate in a
development environment and then gets committed and promoted to production.

## Fix steps

1. **Go**: Remove `InsecureSkipVerify` and configure a CA pool if needed:
   ```go
   certPool, err := x509.SystemCertPool()
   if err != nil { return nil, err }
   tlsConfig := &tls.Config{
       RootCAs:    certPool,
       MinVersion: tls.VersionTLS12,
   }
   transport := &http.Transport{TLSClientConfig: tlsConfig}
   ```
2. **Python**: Remove `verify=False` or point to a CA bundle:
   `verify="/path/to/ca-bundle.pem"`.
3. **Node.js**: Remove `rejectUnauthorized: false` or set `ca` to the correct
   certificate PEM string.
4. For internal services, use a private CA and distribute the root certificate
   via environment configuration rather than disabling verification.
5. For development, use `mkcert` to generate locally-trusted certificates
   rather than disabling TLS validation.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Confirm the client connects successfully with certificate verification enabled.
- Run `go vet ./...` or your linter to catch any remaining misconfigurations.

## Likely files to inspect

- `**/*.go`
- `**/*.py`
- `**/*.ts`
- `**/*.js`


## Run Faultline

```bash
faultline analyze build.log
faultline explain insecure-tls-skip-verify
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- TLS certificate verification disabled in HTTP client
- Security: tls certificate verification disabled in http client
- faultline explain insecure-tls-skip-verify


---

*Generated from [playbooks/bundled/source/insecure-tls-skip-verify.yaml](../../../playbooks/bundled/source/insecure-tls-skip-verify.yaml). Do not edit directly — run `make docs-generate`.*
