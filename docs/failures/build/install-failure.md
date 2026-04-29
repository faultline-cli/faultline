# Package installation failure

**Playbook ID:** `install-failure`
**Category:** build
**Severity:** high
**Tags:** `install`, `dependencies`, `npm`, `pip`, `apt`

## What this failure means

The dependency installation step failed because the requested package or
version could not be found in the configured registry, package index, or OS
repository.

## Common log signals

```text
npm ERR! code E404
npm ERR! 404 Not Found
could not find package
package not found
apt-get install failed
dpkg: error
E: Unable to locate package
```

## Diagnosis

The dependency installation step failed because the requested package or
version could not be found in the configured registry, package index, or OS
repository.

This usually means the package name is wrong, the version constraint points
at an unpublished release, or CI is configured to use a repository mirror
that does not contain the requested artifact.

## Fix steps

1. Re-run the exact install command locally with verbose logging from the
   same package manager, so you can see which package name, version, or
   index URL failed.
2. Check that the package name and requested version actually exist in the
   configured registry or package index using that package manager's own
   metadata lookup command, such as `npm view`, `pip index versions`, or
   `apt-cache policy`.
3. If the package is private, verify the registry configuration files are
   present before the install step runs, for example `.npmrc`, `pip.conf`,
   or the apt sources list used by the job.
4. If a package was removed or renamed, pin to a known-good published version
   or replace it with the maintained package.
5. If CI uses a mirror or proxy, confirm the missing version has been synced
   there before retrying the build.

## Validation

- Run a registry metadata check for the missing package or version.
- Re-run the failing install command.
- Confirm CI proceeds past dependency installation without the original
  package-not-found signature.

## Likely files to inspect

- `package.json`
- `package-lock.json`
- `requirements.txt`
- `pyproject.toml`
- `.npmrc`
- `pip.conf`


## Run Faultline

```bash
faultline analyze build.log
faultline explain install-failure
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- Package installation failure
- Build: package installation failure
- E: Unable to locate package
- GitHub Actions package installation failure
- faultline explain install-failure
- npm package installation failure


---

*Generated from [playbooks/bundled/log/build/install-failure.yaml](../../playbooks/bundled/log/build/install-failure.yaml). Do not edit directly — run `make docs-generate`.*
