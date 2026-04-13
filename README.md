# Faultline

Faultline is a deterministic CLI for CI failure diagnosis.

It reads a raw CI log from a file or stdin, matches it against a reviewed playbook catalog, and returns the most likely failure with evidence and concrete fix steps. The same engine also supports repository inspection through source-aware detector playbooks.

## Starter and Premium

Faultline ships with a bundled starter catalog under `playbooks/bundled/`.

Premium playbooks are designed to live in a separate private repository or private release archive. They are not baked into the public starter release. Instead, you install them once with the CLI and Faultline loads them automatically on future runs.

## Install

The simplest v1 install path is the release tarball from GitHub Releases:

```bash
curl -L <release-tarball-url> -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd faultline_<version>_<os>_<arch>
./faultline analyze build.log
```

Each archive contains:

```text
faultline_<version>_<os>_<arch>/
  faultline
  playbooks/
    bundled/
  README.md
```

If you move the binary elsewhere, keep `playbooks/bundled/` alongside it or set `FAULTLINE_PLAYBOOK_DIR` to the bundled catalog directory.

## Quick Start

```bash
# Analyze a log file
faultline analyze build.log

# Pipe from stdin
cat build.log | faultline analyze

# Detailed human-readable output
faultline analyze --mode detailed build.log

# Emit markdown instead of terminal-oriented raw text
faultline analyze --format markdown build.log

# JSON for automation
faultline analyze --json build.log

# Add local git context
faultline analyze build.log --git --since 30d --repo .

# Print only the top fix steps
faultline fix build.log

# Inspect a repository for source-risk patterns
faultline inspect .

# Produce a deterministic local or agent workflow
faultline workflow build.log
faultline workflow build.log --mode agent --git --repo .

# Browse the catalog
faultline list
faultline list --category auth
faultline explain docker-auth
```

## Premium Upgrade Path

Install premium playbooks into `~/.faultline/packs/` once and Faultline will compose them with the bundled starter catalog automatically.

```bash
# Clone or unpack your private premium pack
git clone <private-premium-pack-repo> ../faultline-premium-pack

# Install it into the local Faultline directory
faultline packs install ../faultline-premium-pack

# Inspect installed packs
faultline packs list

# Verify the premium rules are active
faultline list
```

When you pull a newer premium pack version, reinstall it in place:

```bash
cd ../faultline-premium-pack && git pull
faultline packs install --force ../faultline-premium-pack
```

For one-off testing without installing, compose an extra pack directly:

```bash
faultline analyze --playbook-pack ../faultline-premium-pack build.log
```

Use `--playbooks <dir>` only when you want a full catalog override. It replaces the starter catalog instead of composing with it.

## Commands

| Command | Description |
|---------|-------------|
| `analyze [file]` | Ranked diagnosis from a file or stdin |
| `inspect [path]` | Ranked source-risk findings from a repository tree |
| `fix [file]` | Print only the fix steps for the top diagnosis |
| `workflow [file]` | Build a deterministic local or agent-ready follow-up plan |
| `list` | List available playbooks from the active catalog |
| `explain <id>` | Show full detail for one playbook |
| `packs install <dir>` | Install an extra playbook pack into `~/.faultline/packs/` |
| `packs list` | Show installed extra playbook packs |

## Common Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--top N` | 1 | Show top N ranked results |
| `--mode quick\|detailed` | `quick` | Human output verbosity |
| `--format raw\|markdown` | `raw` | Human output shape for analyze, inspect, fix, and explain |
| `--json` | false | Emit stable JSON |
| `--ci-annotations` | false | Emit GitHub Actions `::warning::` lines |
| `--playbooks <dir>` | auto | Replace the active catalog with one directory |
| `--playbook-pack <dir>` | none | Add one or more extra pack roots on top of starter |
| `--no-history` | false | Skip reading and writing local history |
| `--git` | false | Enrich diagnosis output with recent local git context |
| `--since <window>` | `30d` | History window for `--git` |
| `--repo <path>` | `.` | Repository path used by `--git` |

Environment variables:

| Variable | Description |
|----------|-------------|
| `FAULTLINE_PLAYBOOK_DIR` | Full catalog override, equivalent to `--playbooks` |
| `FAULTLINE_PLAYBOOK_PACKS` | Additional pack roots, separated by the platform path separator |
| `FAULTLINE_PLAIN=1` | Force plain text output even on interactive terminals |
| `NO_COLOR` | Disable ANSI colors |

## Output Modes

Faultline auto-detects terminal output. Styled rendering is used only for interactive terminals. Redirected output, CI, `NO_COLOR`, and `FAULTLINE_PLAIN=1` fall back to readable plain text. `--json` always bypasses terminal styling.

`--format raw` keeps the current terminal-oriented output shape, with styling when a TTY is available and plain text everywhere else. `--format markdown` emits the same ranked diagnosis content as markdown source, which is useful for redirecting into files, issue templates, or assistant handoff notes.

Quick mode keeps the response short and action-first. Detailed mode includes fuller evidence and any available repository context. JSON keeps the same deterministic ranking in a stable schema for automation.

## Workflow Output

`faultline workflow` turns the top diagnosis into a deterministic next-step plan. `--mode local` focuses on practical triage. `--mode agent` adds a structured handoff prompt for a coding assistant without changing Faultline's deterministic runtime.

## Docker

Build and run the starter image:

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/build.log
docker run --rm -v "$(pwd)":/workspace faultline analyze --json /workspace/build.log
```

The image bundles the starter catalog at `/playbooks/bundled` and also prepares `/home/faultline/.faultline/packs` for installed premium packs.

To use the same premium packs you installed locally, mount your Faultline home directory into the container:

```bash
docker run --rm \
  -v "$(pwd)":/workspace \
  -v "$HOME/.faultline":/home/faultline/.faultline \
  faultline list

docker run --rm \
  -v "$(pwd)":/workspace \
  -v "$HOME/.faultline":/home/faultline/.faultline \
  faultline analyze /workspace/build.log
```

If you need a custom image with premium rules baked in, build a thin derived image that copies the premium pack into `/home/faultline/.faultline/packs/<pack-name>`.

## CI Usage

Examples:

```bash
# Emit machine-readable output for a later step
faultline analyze --json build.log

# Emit GitHub Actions annotations
faultline analyze --ci-annotations build.log

# Build an agent-ready workflow handoff
faultline workflow build.log --mode agent --json
```

## Validation and Release Checks

Core validation:

```bash
make test
make review
make bench
```

Release validation:

```bash
make release-check VERSION=v0.1.0
make release-check VERSION=v0.1.0 WITH_DOCKER=1 IMAGE=faultline-smoke
```

When the private premium repository is checked out locally, include it in release validation so cross-pack duplicate IDs or load failures fail before publish:

```bash
make premium-check PREMIUM_PACK_DIR=../faultline-premium-pack
make premium-review PREMIUM_PACK_DIR=../faultline-premium-pack
make release-check VERSION=v0.1.0 PREMIUM_PACK_DIR=../faultline-premium-pack
```

`make premium-link` creates the ignored local symlink used by the Makefile convenience flow, and `make premium-path` shows which premium directory those targets resolve.

## Playbooks

Playbooks are deterministic YAML files: structured fields drive matching and ranking, while markdown fields explain the diagnosis and fix. See [docs/playbooks.md](docs/playbooks.md) for authoring guidance and [docs/architecture.md](docs/architecture.md) for catalog resolution details.

## Development

```bash
make build
make test
make run LOG=build.log
make docker-smoke IMAGE=faultline-smoke
```
