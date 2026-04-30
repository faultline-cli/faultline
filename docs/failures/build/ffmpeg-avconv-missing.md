# ffmpeg or avconv not available in the CI environment

**Playbook ID:** `ffmpeg-avconv-missing`
**Category:** build
**Severity:** medium
**Tags:** `ffmpeg`, `avconv`, `media`, `toolchain`, `binary`, `path`

## What this failure means

The job requires ffmpeg (or its libav counterpart avconv) for media processing
but neither tool is installed or reachable on `$PATH` in the CI environment.

## Common log signals

```text
ffmpeg or avconv could not be found
avconv could not be found
ffmpeg could not be found
please install ffmpeg
please install one
ffmpeg is not installed
Unable to locate ffmpeg
```

## Diagnosis

A media-aware test or tool checked for ffmpeg or avconv and could not locate
either binary. This appears most often in projects that process video, audio,
or streaming formats (HLS/m3u8, DASH, MP4, WebM).

Common causes:

- **Missing system dependency** — the CI base image does not include ffmpeg.
- **Minimal runner image** — container jobs that use a stripped-down Linux
  image (Alpine, debian-slim) often omit ffmpeg.
- **Changed image tag** — an update to the base container silently dropped
  the ffmpeg package.
- **PATH mismatch** — ffmpeg is installed to a non-standard prefix that is not
  included in the job's `$PATH`.

## Fix steps

### Ubuntu / Debian CI (GitHub Actions, GitLab CI, CircleCI)

Add an installation step before any test steps that require ffmpeg:

```yaml
- name: Install ffmpeg
  run: sudo apt-get update && sudo apt-get install -y ffmpeg
```

### Alpine (Docker-based runner)

```dockerfile
RUN apk add --no-cache ffmpeg
```

### macOS runner (Homebrew)

```yaml
- name: Install ffmpeg
  run: brew install ffmpeg
```

### Verify on the runner

```bash
ffmpeg -version
which ffmpeg
```

### Pin to a specific version

If different test suites require different ffmpeg capabilities, pin the version
in your runner image or install step to avoid silent regressions when the
package is updated.

## Validation

- `which ffmpeg` exits zero and prints the binary path.
- `ffmpeg -version` prints a version string without errors.
- The failing test suite runs to completion without "ffmpeg could not be found"
  in the output.

## Likely files to inspect

- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `Dockerfile`
- `Makefile`


## Run Faultline

```bash
faultline analyze build.log
faultline explain ffmpeg-avconv-missing
faultline workflow build.log --json --mode agent
```

## Search phrases this page answers

- ffmpeg or avconv not available in the CI environment
- Build: ffmpeg or avconv not available in the ci environment
- ffmpeg or avconv could not be found
- GitHub Actions ffmpeg or avconv not available in the ci environment
- faultline explain ffmpeg-avconv-missing


---

*Generated from [playbooks/bundled/log/build/ffmpeg-avconv-missing.yaml](../../../playbooks/bundled/log/build/ffmpeg-avconv-missing.yaml). Do not edit directly — run `make docs-generate`.*
