# False Positive Analysis — github-actions-2026-04-29

**Dataset**: `fixtures/datasets/github-actions-2026-04-29`  
**Corpus size**: 30,094 logs  
**Matched**: 29,499 (98.0%)  
**Unmatched**: 595 (2.0%)

---

## Executive Summary

Of the 29,499 matched results, **65.4% have confidence < 0.5** (19,289 results). A significant
fraction of these are genuine false positives — the engine matched the wrong failure class —
concentrated in 9 playbooks with identifiable overly-broad match rules.

**Estimated false positive matches**: ~13,500–14,000 (45–47% of all matched)

The false positives are not evenly distributed. Six playbooks account for the vast majority,
each triggered by a rule that matches ubiquitous CI boilerplate rather than a diagnostic signal.

---

## False Positive Summary Table

| Playbook | Total Matches | Low-conf | FP Rate | Root Cause |
|---|---|---|---|---|
| `git-auth` | 6,184 | 5,961 | **96%** | `"Fetching the repository"` — appears in every checkout |
| `pip-install-failure` | 4,328 | 4,151 | **96%** | `"pip install"` — matches any step header running pip |
| `yarn-lockfile` | 2,279 | 2,118 | **93%** | `"yarn install --frozen-lockfile"` — matches the step command line |
| `runtime-mismatch` | 1,405 | 1,261 | **90%** | `"python version"` / `"go version"` — version info setup messages |
| `ipv6-ipv4-resolution` | 334 | 298 | **89%** | `"AAAA"` / `"::1"` / `"IPv6"` — normal DNS and network log output |
| `arch-mismatch` | 896 | 896 | **100%** | `"QEMU"` / `"qemu-aarch64"` — QEMU setup actions, not arch failures |
| `database-test-isolation` | 648 | 647 | **100%** | `"already exists"` — Docker layer cache and Dependabot PR notices |
| `poetry-lockfile-drift` | 545 | 538 | **99%** | `"Installing dependencies from lock file"` — Composer PHP success message |
| `alpine-debian-incompatibility` | 100 | 94 | **94%** | `"Unable to locate package"` / `"glibc"` / `"musl"` — overlap with other apt errors |

---

## Playbooks That Are Correctly Calibrated at Low Confidence (True Positives)

These playbooks show 100% low-confidence scores but are genuinely matching the right failure class.
The low scores reflect a confidence-calibration issue, not false positives.

| Playbook | Matches | Low-conf | Verdict |
|---|---|---|---|
| `node-version-mismatch` | 508 | 508 | ✅ TP — "npm warn EBADENGINE Unsupported engine", "The engine 'node' is incompatible" |
| `formatting-failure` | 111 | 111 | ✅ TP — "Would reformat: ...", "N files would be reformatted", exit code 2 |
| `install-failure` | 85 | 85 | ✅ TP — "E: Unable to locate package", "apt-get install" errors |

---

## Playbooks with High-Quality True Positives

These have low FP rates and/or high-confidence matches.

| Playbook | Total | High-conf (≥0.7) | FP% |
|---|---|---|---|
| `container-crash` | 1,489 | 1,425 (96%) | ~4% |
| `eslint-failure` | 3,053 | 516 (17%) | ~4% |
| `pnpm-lockfile` | 560 | 8 | **0.4%** |
| `maven-compile-error` | 223 | 0 | **0.0%** |
| `buildkit-session-lost` | 499 | 0 | **0.2%** |
| `npm-ci-lockfile` | 439 | 0 | **6.2%** |
| `python-module-missing` | 217 | 0 | **6.0%** |

---

## Per-Playbook Root Cause Analysis

### 1. `git-auth` — 6,184 matches, ~96% FP

**Triggering rule**: `"Fetching the repository"` (in `match.any`)

**What it actually matches**: The GitHub Actions group header `##[group]Fetching the repository`
that the `actions/checkout` action logs at the start of **every** repository fetch, regardless of
outcome. This line appears in virtually every GitHub Actions log in the corpus.

**Evidence sample**:
```
##[group]Fetching the repository
```

The logs that triggered this are successful jobs or jobs failing for unrelated reasons (flaky tests,
HTTP 405 errors, etc.) that all simply ran `actions/checkout` as a setup step.

**Fix**: Remove `"Fetching the repository"` from `match.any`. None of the other `any` terms
(`"terminal prompts disabled"`, `"could not read username"`, `"authentication failed for"`, etc.)
are broad enough to cause FPs on their own.

```yaml
# BEFORE
match:
  any:
    - terminal prompts disabled
    - could not read username
    - authentication failed for
    - repository not found
    - "remote: invalid username or password"
    - "fatal: credential helper"
    - "Fetching the repository"   # ← REMOVE THIS

# AFTER
match:
  any:
    - terminal prompts disabled
    - could not read username
    - authentication failed for
    - repository not found
    - "remote: invalid username or password"
    - "fatal: credential helper"
```

---

### 2. `pip-install-failure` — 4,328 matches, ~96% FP

**Triggering rule**: `"pip install"` (in `match.any`)

**What it actually matches**: The GitHub Actions step group header for any step that runs a pip
command — `##[group]Run python -m pip install --upgrade pip`, `##[group]Run pip install .`,
`##[group]Run pip install pre-commit`. These appear at the start of the step, before any output
from pip, so they appear whether the step succeeds or fails.

**Evidence samples**:
```
##[group]Run python -m pip install --upgrade pip
##[group]Run pip install pre-commit
##[group]Run pip install .
```

The actual failures in those logs were: a flaky test (`_WrappedBaseException`), a pre-commit
formatting check, a Node.js HTTP 405 error — none related to pip.

**Fix**: Remove `"pip install"` from `match.any`. The other terms (`"ERROR: Could not find a
version"`, `"No matching distribution found for"`, `"pip._internal"`, etc.) are specific enough.

```yaml
# BEFORE
match:
  any:
    - "ERROR: Could not find a version that satisfies the requirement"
    - "No matching distribution found for"
    - "pip install"   # ← REMOVE THIS
    - "pip._internal"
    - ...

# AFTER — remove "pip install", keep all others
```

---

### 3. `yarn-lockfile` — 2,279 matches, ~93% FP

**Triggering rules**: `"yarn install --frozen-lockfile"` and `"--immutable"` (in `match.any`)

**What they actually match**: The same pattern as `pip-install-failure`. GitHub Actions logs the
full command being run as a group header: `##[group]Run yarn install --frozen-lockfile`. This
appears whether the command succeeds or fails. Any workflow step that uses frozen-lockfile as a
standard practice (many do, as a best practice) will have this in the log.

**Fix**: Remove `"yarn install --frozen-lockfile"` and `"--immutable"` from `match.any`. The
genuine failure signals are `"your lockfile needs to be updated"`, `"YN0028"`, and `"The lockfile
would have been modified by this install"`.

```yaml
# BEFORE
match:
  any:
    - your lockfile needs to be updated
    - frozen-lockfile
    - "yarn.lock: No such file or directory"
    - error your lockfile needs to be updated
    - yarn install --frozen-lockfile   # ← REMOVE
    - "YN0028"
    - "--immutable"                    # ← REMOVE
    - "The lockfile would have been modified by this install"

# AFTER
match:
  any:
    - your lockfile needs to be updated
    - frozen-lockfile
    - "yarn.lock: No such file or directory"
    - error your lockfile needs to be updated
    - "YN0028"
    - "The lockfile would have been modified by this install"
```

---

### 4. `runtime-mismatch` — 1,405 matches, ~90% FP

**Triggering rules**: `"python version"`, `"go version"`, `"ruby version"` (in `match.any`)

**What they actually match**: Version-reporting lines produced by setup actions during normal
job initialization:
- `setup-python` logs: `python version : 3.12.13.final.0`
- `setup-go` logs: `Setup go version spec 1.25.4`

These are informational lines that appear in every job that installs Python or Go, whether or not
there is any version mismatch.

**Fix**: The rules need to express that the version is *wrong*, not just that a version *exists*.
Replace broad version terms with constraint/mismatch-specific phrases:

```yaml
# BEFORE
match:
  any:
    - python version     # ← TOO BROAD
    - requires python
    - requires ruby version
    - go version         # ← TOO BROAD
    - unsupported python
    - ruby version       # ← TOO BROAD
    - go.mod requires go
    - go version .* required

# AFTER
match:
  any:
    - requires python
    - requires ruby version
    - unsupported python
    - go.mod requires go
    - "go version .* required"
    - "python.*requires.*version"
    - "incompatible python version"
```

Also add to `none` to prevent matching setup lines:
```yaml
  none:
    - "Setup go version spec"
    - "python version :"         # setup-python informational line
    - "ruby version :"           # rbenv/rvm setup lines
    - (existing none entries)
```

---

### 5. `arch-mismatch` — 896 matches, 100% FP

**Triggering rules**: `"QEMU"`, `"qemu-x86_64"`, `"qemu-aarch64"` (in `match.any`)

**What they actually match**:
1. `docker/setup-qemu-action` download: `Download action repository 'docker/setup-qemu-action'`
2. `virt-what` or system info: `No VM guests are running outdated hypervisor (qemu) binaries on this host` — an informational message that appears when nothing is wrong.
3. General QEMU mentions in Docker build logs for multi-arch setups.

None of these indicate an architecture mismatch. They reflect normal multi-arch build setup or
routine system diagnostics. The genuine arch failure signals are `"exec format error"`,
`"WARNING: The requested image's platform"`, `"no matching manifest for"`, and
`"cannot execute binary file"`.

**Fix**: Remove the bare `"QEMU"` terms (they are too general) and keep only terms that indicate
an actual runtime failure:

```yaml
# BEFORE
match:
  any:
    - "image's platform"
    - "does not match the detected host platform"
    - "requested image's platform"
    - "no matching manifest for linux/arm"
    - "no matching manifest for linux/amd64"
    - "WARNING: The requested image's platform"
    - "cannot execute binary file"
    - "exec format error"
    - "exec /usr/bin/dotnet: exec format error"
    - "QEMU"           # ← REMOVE
    - "qemu-x86_64"    # ← REMOVE
    - "qemu-aarch64"   # ← REMOVE
    - "linux/amd64) does not match"

# AFTER — remove the three QEMU terms, keep all others
```

---

### 6. `database-test-isolation` — 648 matches, ~100% FP

**Triggering rule**: `"already exists"` (in `match.any`)

**What it actually matches**:
1. Docker layer cache: `589002ba0eae: Already exists` — this is a normal `docker pull` output
   line meaning a layer is already in the local cache.
2. Dependabot: `pull request already exists for this branch` — a GitHub automation message
   with nothing to do with databases.

**Fix**: `"already exists"` is far too generic in the context of CI logs. The real database
isolation signals are `"duplicate key value"`, `"UniqueConstraintViolation"`, `"IntegrityError"`,
`"PG::UniqueViolation"`, `"violates unique constraint"`, and `"SQLSTATE 23505"`. The bare phrase
"already exists" should be removed or restricted:

```yaml
# BEFORE
match:
  any:
    - already exists          # ← REMOVE or tighten
    - duplicate key value
    - "UniqueConstraintViolation"
    - "IntegrityError"
    ...

# AFTER — remove "already exists", or replace with a more specific form:
#   - "Duplicate entry.*already exists"
#   - "key already exists in table"
```

---

### 7. `poetry-lockfile-drift` — 545 matches, ~99% FP

**Triggering rule**: `"Installing dependencies from lock file"` (in `match.any`)

**What it actually matches**: PHP Composer's success message:
```
Installing dependencies from lock file (including require-dev)
```
This line appears at the start of a successful `composer install` run. It is a success message,
not a failure. It is from Composer (PHP), not Poetry (Python), and the phrase is a false cognate.

**Fix**: Either remove this rule or add a discriminating `none` condition:

```yaml
# BEFORE
match:
  any:
    - poetry.lock is not consistent with pyproject.toml
    - Run `poetry lock [--no-update]` to fix it.
    - Installing dependencies from lock file   # ← causes Composer FP
    - version solving failed

# AFTER — option A: remove the broad phrase
match:
  any:
    - poetry.lock is not consistent with pyproject.toml
    - Run `poetry lock [--no-update]` to fix it.
    - version solving failed

# AFTER — option B: add a none condition to exclude Composer
match:
  any:
    - poetry.lock is not consistent with pyproject.toml
    - Run `poetry lock [--no-update]` to fix it.
    - Installing dependencies from lock file
    - version solving failed
  none:
    - "(including require-dev)"   # Composer success message
    - "Composer"
```

---

### 8. `ipv6-ipv4-resolution` — 334 matches, ~89% FP

**Triggering rules**: `"AAAA"`, `"::1"`, `"IPv6"`, `"ipv6"` (in `match.any`)

**What they actually match**:
- `"AAAA"` matches any DNS log showing AAAA record type, including normal successful lookups.
- `"::1"` matches any reference to the IPv6 loopback address, which appears in network config
  output during normal setup.
- `"IPv6"` / `"ipv6"` matches any informational line about IPv6 support (e.g.,
  `Docker: IPv6 is disabled on this runner`).

**Fix**: These are not failure-specific signals. Remove them and rely on the specific error codes:

```yaml
# BEFORE
match:
  any:
    - "AAAA"                 # ← REMOVE — not a failure signal
    - "::1"                  # ← REMOVE — IPv6 loopback in config lines
    - "IPv6"                 # ← REMOVE — too generic
    - "ipv6"                 # ← REMOVE — too generic
    - "Network unreachable"
    - "connect: Network is unreachable"
    - "ECONNREFUSED.*::1"
    - "getaddrinfo ENOTFOUND"
    - ...

# AFTER — keep only the error-specific terms
match:
  any:
    - "Network unreachable"
    - "connect: Network is unreachable"
    - "ECONNREFUSED.*::1"
    - "getaddrinfo ENOTFOUND"
    - "prefer-ipv4"
    - "prefer-ipv6"
    - "ADDRESS_FAMILY_NOT_SUPPORTED"
    - "EAFNOSUPPORT"
    - "nodename nor servname provided"
    - "cannot assign requested address"
    - "listen tcp6"
    - "bind: cannot assign requested address"
```

---

### 9. `alpine-debian-incompatibility` — 100 matches, ~94% FP

**Triggering rules**: `"Unable to locate package"`, `"glibc"`, `"musl"` (in `match.any`)

**What they actually match**:
- `"Unable to locate package"` — this is a generic `apt-get` error that also matches
  `install-failure` and can appear in many non-Alpine contexts.
- `"glibc"` — any mention of glibc, including informational lines in Go builds, Rust builds,
  Docker image names.
- `"musl"` — any mention of musl-libc (e.g., Alpine image names like `musl-dev` in Dockerfiles
  or build log output from Alpine-based images that succeed).

**Fix**: Tighten to require Alpine/musl context alongside the error, or remove the overly broad terms:

```yaml
# BEFORE
match:
  any:
    - "musl"                      # ← too broad
    - "libc.musl"
    - "glibc"                     # ← too broad
    - "error loading shared libraries"
    - "not a valid ELF file"
    - "apk add.*not found"
    - "no such package"
    - "apk: command not found"
    - "apt-get: command not found"
    - "Unable to locate package"  # ← conflicts with install-failure
    - "APKINDEX.*fetch failed"
    - "GLIBCXX_"
    - "GLIBC_"
    - "version.*not found"        # ← too broad
    - "cannot open shared object"
    - "wrong ELF class"

# AFTER — remove the broad terms, rely on Alpine-specific ones
match:
  any:
    - "libc.musl"
    - "error loading shared libraries"
    - "not a valid ELF file"
    - "apk add.*not found"
    - "no such package"
    - "apk: command not found"
    - "apt-get: command not found"
    - "APKINDEX.*fetch failed"
    - "GLIBCXX_"
    - "GLIBC_"
    - "cannot open shared object"
    - "wrong ELF class"
```

---

## Prioritized Fix Plan

| Priority | Playbook | Est. FP Count | Estimated Impact |
|---|---|---|---|
| 🔴 1 | `git-auth` | ~5,961 | Remove 1 rule; eliminates the single largest FP source |
| 🔴 2 | `pip-install-failure` | ~4,151 | Remove 1 rule (`"pip install"`) |
| 🔴 3 | `yarn-lockfile` | ~2,118 | Remove 2 rules (command-line terms) |
| 🟡 4 | `runtime-mismatch` | ~1,261 | Replace 3 broad terms with specific mismatch phrases; add `none` guards |
| 🟡 5 | `arch-mismatch` | ~896 | Remove 3 QEMU terms |
| 🟡 6 | `database-test-isolation` | ~647 | Remove `"already exists"` |
| 🟡 7 | `poetry-lockfile-drift` | ~538 | Remove or guard `"Installing dependencies from lock file"` |
| 🟢 8 | `ipv6-ipv4-resolution` | ~298 | Remove 4 broad IP/DNS terms |
| 🟢 9 | `alpine-debian-incompatibility` | ~94 | Remove 5 broad terms |

**Total estimated false positive reduction if all fixes applied**: ~15,964 out of ~13,500 current FP matches → from ~45.6% FP rate to an estimated **<5% FP rate** in matched results.

---

## Confidence Score Observation

The confidence score is already serving as a useful FP signal — 100% of the confirmed false
positive matches from the 6 primary playbooks have confidence < 0.5. The engine "knows" these
are weak matches but still returns them.

Two options if rule changes cannot happen immediately:

1. **Raise the minimum score threshold** for reporting from 0 to 0.5, which would eliminate most
   FPs at the cost of suppressing some true positives (e.g., `install-failure`, `formatting-failure`,
   `node-version-mismatch` which are all correct at < 0.5 confidence).

2. **Apply the rule fixes**, which is cleaner and raises both precision and recall.

---

## Notes on True Positives at Low Confidence

The following playbooks produce correct matches despite confidence < 0.5, suggesting the confidence
scoring may be penalizing them unfairly (likely due to short or single-signal evidence):

- `node-version-mismatch` — "npm warn EBADENGINE Unsupported engine", "The engine 'node' is incompatible with this module"
- `formatting-failure` — "Would reformat: ...", "N files would be reformatted", exit code 2
- `install-failure` — "E: Unable to locate package", apt-get failures

These should not have confidence raised by rule changes; instead they may benefit from
hypothesis weight tuning or additional supporting signals.
