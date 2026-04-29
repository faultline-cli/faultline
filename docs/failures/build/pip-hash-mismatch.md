# pip hash-checking mode failure

**Playbook ID:** `pip-hash-mismatch`
**Category:** build
**Severity:** high
**Tags:** `python`, `pip`, `hash`, `security`, `supply-chain`, `requirements`

## What this failure means

pip rejected one or more downloaded packages because their hash did not match the expected value recorded in `requirements.txt`. The install was aborted to prevent supply-chain tampering.

## Common log signals

```text
THESE PACKAGES DO NOT MATCH THE HASHES FROM THE REQUIREMENTS FILE
Expected sha256
Hash mismatch (got:
does not match expected hash
RECORD mismatch
hash of the downloaded file
There are no versions that match the hash
```

## Diagnosis

pip's hash-checking mode is active when `requirements.txt` contains `--hash=sha256:...` entries. pip verifies every downloaded file against the recorded hash before installing it. If the hash is absent, wrong, or the package was updated at the source without a lockfile update, pip aborts the install.

Common causes:

- `requirements.txt` was updated (version bump or new package) but the hash entries were not regenerated.
- A package was re-released at the same version number on PyPI, changing the wheel hash.
- A custom or internal index is serving a different file than the one that was originally recorded.
- The `requirements.txt` was edited by hand without regenerating hashes via `pip-compile --generate-hashes`.

## Fix steps

1. Identify which package failed from the error output — the line will show the expected hash and the received hash.

2. If using `pip-tools`, regenerate the fully hashed requirements file:

   ```bash
   pip-compile --generate-hashes requirements.in -o requirements.txt
   ```

3. If maintaining hashes by hand, update the failing entry:

   ```bash
   pip download <package>==<version> -d /tmp/pkg
   sha256sum /tmp/pkg/<wheel-file>
   ```

   Replace the old `--hash=sha256:...` line with the new value.

4. If the package is from an internal registry, verify the registry is not caching a stale or modified artifact.

5. If hash checking is not intentional in this project, remove all `--hash=sha256:...` lines from `requirements.txt` (but doing so weakens supply-chain guarantees).

## Validation

- `pip install -r requirements.txt` completes successfully.
- No `Hash mismatch` or `THESE PACKAGES DO NOT MATCH` lines appear in the output.

## Likely files to inspect

- `requirements.txt`
- `requirements-dev.txt`
- `requirements.in`
- `pyproject.toml`


## Run Faultline

```bash
faultline analyze build.log
faultline explain pip-hash-mismatch
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- pip hash-checking mode failure
- Build: pip hash-checking mode failure
- THESE PACKAGES DO NOT MATCH THE HASHES FROM THE REQUIREMENTS FILE
- GitHub Actions pip hash-checking mode failure
- faultline explain pip-hash-mismatch
- Python pip hash-checking mode failure


---

*Generated from [playbooks/bundled/log/build/pip-hash-mismatch.yaml](../../playbooks/bundled/log/build/pip-hash-mismatch.yaml). Do not edit directly — run `make docs-generate`.*
