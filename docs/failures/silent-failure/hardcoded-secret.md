# Hardcoded secret, token, or password in source code

**Playbook ID:** `hardcoded-secret`
**Category:** silent_failure
**Severity:** high
**Tags:** `source`, `security`, `secret`, `credentials`, `owasp`

## What this failure means

A secret-named variable is assigned a literal string value in source code, exposing credentials, API keys, or tokens to anyone with repository access.

## Common log signals

*(This playbook uses source-code pattern matching rather than log signals.)*

## Diagnosis

A variable whose name suggests it holds a secret (API key, token, password,
or credential) is assigned a literal string directly in source code.

Hardcoded secrets are committed to version control, visible in PR diffs,
and persist in git history even after removal. This affects the security of
all environments where the secret is valid.

Common patterns:
- `const APIKey = "sk_live_..."` in Go config or init files
- `API_KEY = "..."` in Python settings modules
- `token: "..."` in YAML configuration committed to the repo

## Fix steps

1. Remove the literal secret value from source code immediately.
2. Replace it with an environment variable lookup: `os.Getenv("API_KEY")` in
   Go, `process.env.API_KEY` in Node.js, `ENV["API_KEY"]` in Ruby, etc.
3. Store the actual secret in CI secrets, a secrets manager (Vault, AWS SSM,
   GCP Secret Manager), or an `.env` file that is in `.gitignore`.
4. If the secret has already been committed, rotate it — git history is
   permanent and the exposed value should be treated as compromised.
5. Add a `git-secrets` or `trufflehog` pre-commit hook to catch future leaks.

## Validation

- Re-run `faultline inspect .` or `faultline guard .`.
- Confirm the finding is resolved and the secret is loaded from the environment.
- Run `git log -p | grep -i api_key` to confirm the literal is not present in
  committed history.

## Likely files to inspect

- `config.go`
- `config.py`
- `settings.py`
- `config.js`
- `config.ts`
- `.env`
- `docker-compose.yml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain hardcoded-secret
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Hardcoded secret, token, or password in source code
- Silent Failure: hardcoded secret, token, or password in source code
- faultline explain hardcoded-secret


---

*Generated from [playbooks/bundled/source/hardcoded-secret.yaml](../../../playbooks/bundled/source/hardcoded-secret.yaml). Do not edit directly — run `make docs-generate`.*
