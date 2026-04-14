# Faultline

Deterministic CI failure triage. No guessing. No rereading logs.

Give it a failing log or repository checkout and it returns:

- the likeliest known failure pattern
- the exact evidence behind that match
- concrete fix and validation steps

Built on explicit playbooks, not probabilistic summaries.

Works well for common failures like Docker auth issues, missing executables, runtime mismatches, and CI misconfigurations.

![Faultline terminal modes demo](docs/readme-assets/docker-auth-modes.gif)

- Analyze logs from a file or stdin.
- Inspect a repository for source-level failure risks.
- Return deterministic terminal, markdown, and JSON output.
- Run locally with explicit, reviewable rules.

## Why use it

Faultline is for repeatable CI failures, not open-ended debugging.

Use it when:

- a job failed and you want the shortest path to a likely cause
- the log is noisy enough that manual reading is slow
- you want evidence-backed fix steps instead of a vague summary
- repository layout or config drift may be part of the problem

Use raw log reading when the failure mode is genuinely new and there is no matching playbook yet.

## Install and try it

### Release archive

```bash
VERSION=v0.1.0
curl -L "https://github.com/faultline-cli/faultline/releases/download/${VERSION}/faultline_${VERSION}_linux_amd64.tar.gz" -o faultline.tar.gz
tar -xzf faultline.tar.gz
cd "faultline_${VERSION}_linux_amd64"
./faultline analyze examples/docker-auth.log
```

### Build from source

```bash
make build
./bin/faultline analyze examples/docker-auth.log
```

### Docker

```bash
docker build -t faultline .
docker run --rm -v "$(pwd)":/workspace faultline analyze /workspace/examples/docker-auth.log
```

### Minimal usage

```bash
# Analyze a log file
faultline analyze build.log

# Read from stdin
cat build.log | faultline analyze

# Emit stable JSON for automation
faultline analyze build.log --json

# Render markdown instead of terminal output
faultline analyze build.log --format markdown

# Print only the fix steps for the top diagnosis
faultline fix build.log --format markdown

# Inspect a repository for source-level findings
faultline inspect .
```

## Example

The demo above shows the same Docker auth failure first with the default terminal output, then with the more detailed terminal view.

```bash
# Default terminal output
faultline analyze examples/docker-auth.log

# Markdown output
faultline analyze examples/docker-auth.log --format markdown

# More detailed terminal output
faultline analyze examples/docker-auth.log --mode detailed
```

When you want the remediation path directly:

```bash
faultline fix examples/docker-auth.log --format markdown
```

When you want the full playbook behind the match:

```bash
faultline explain docker-auth
```

## Why Faultline is different

Faultline does a small number of things on purpose.

- Deterministic: the same input and playbook set produce the same result.
- Evidence-backed: diagnoses point to matched log lines, not hidden reasoning.
- Local-first: no hosted analysis service and no runtime network dependency.
- Reviewable: playbooks are structured rules with operator-facing guidance.
- Action-oriented: output includes fix and validation steps, not just labels.

That tradeoff is intentional. Faultline is not trying to guess every failure. It is trying to be reliable on the failures it knows.

## Commands

| Command | Purpose |
| --- | --- |
| `analyze [file]` | Diagnose a CI log from a file or stdin |
| `fix [file]` | Print the fix steps for the top diagnosis |
| `inspect [path]` | Inspect a repository for source-level failure risks |
| `explain <id>` | Show the full details for one playbook |
| `workflow [file]` | Generate a deterministic follow-up workflow |
| `list` | List available playbooks |
| `packs` | Install or inspect additional playbook coverage |
| `completion` | Generate shell completion scripts |

### Common flags

| Flag | Description |
| --- | --- |
| `--json` | Emit machine-readable JSON |
| `--format terminal\|markdown\|json` | Select the output format |
| `--mode quick\|detailed` | Control output detail for human-readable results |
| `--top N` | Show the top N ranked results |
| `--git` | Enrich analysis with recent local git context |
| `--repo <path>` | Choose the repository path used by `--git` |

## Playbooks and coverage

Faultline ships with a bundled catalog that is useful on first run.

- common CI failure patterns are included in the default release
- evidence, diagnosis, fix, and validation guidance ship with each playbook
- `inspect` includes baseline source-level coverage without requiring extra installs

### Example playbooks

A small sample of the bundled catalog:

**Auth and access**

- `docker-auth` - Docker registry authentication failure
- `aws-credentials` - AWS credentials missing or invalid

**Build and environment**

- `missing-executable` - Required executable or runtime binary missing
- `runtime-mismatch` - Python, Ruby, or Go runtime version mismatch
- `cache-corruption` - Corrupted or stale dependency cache

**Runtime and infrastructure**

- `permission-denied` - Permission denied
- `oom-killed` - Process killed by OOM killer
- `dns-resolution` - DNS resolution failure

See the full list with:

```bash
faultline list
```

Faultline also supports installed playbook packs for extra coverage. That is the upgrade path for narrower or deeper failure modes without forking the CLI or replacing the bundled catalog.

Right now that mainly means one optional premium playbook pack. It adds coverage in narrower areas such as provider-specific workflows and deeper ecosystem coverage.

Install an additional pack once and it will load automatically on future runs:

```bash
faultline packs install ./path/to/pack
faultline packs list
faultline list
```

The bundled catalog stands on its own. Packs are there if you need more coverage in a specific area.

## Validation and credibility

- `./bin/faultline fixtures stats` currently reports 112 accepted real fixtures.
- Current regression snapshot reports top-1 = 1.000, top-3 = 1.000, unmatched = 0.000, false_positive = 0.000.
- Bundled coverage currently ships with 67 playbook files under `playbooks/bundled/`.
- Release validation runs `make test`, `make review`, `make fixture-check`, archive smoke tests, and Docker smoke tests in CI.

These numbers describe the checked-in regression corpus, not the full space of CI failures.

## Feedback and coverage requests

The most useful feedback is a real failure Faultline should explain better.

- Open an issue with a sanitized log excerpt, the expected diagnosis, and the relevant context.
- Add or refine fixtures when a failure should stay covered in regression tests.
- If Faultline missed a recurring failure in your stack, send that case.
- If you need coverage beyond the bundled catalog, get in touch about the premium pack.

Raw ingestion artifacts belong in `fixtures/staging/` as a local review queue only. Sanitize them first, then promote accepted cases into `fixtures/real/`.

## Development

```bash
make build
make test
make review
make demo-assets
```

`make demo-assets` regenerates the README terminal GIFs and screenshots from source-controlled VHS tapes under `docs/readme-assets/tapes/`. The canonical entrypoint is `tools/render-demo-assets.sh`; it prefers repo-local tools under `.tools/bin`, otherwise uses local `vhs`, and falls back to the official VHS Docker image when Docker is installed.

Helpful references:

- `examples/README.md` for runnable sample logs and expected outputs
- `docs/architecture.md` for package boundaries and runtime flow
- `docs/playbooks.md` for playbook authoring rules
- `docs/distribution.md` for release and Docker packaging
- `docs/detectors.md` for detector expectations
- `docs/adr/README.md` for architectural decisions
- `CONTRIBUTING.md` for contribution and fixture-sanitization guidance
- `SECURITY.md` for vulnerability reporting expectations
- `CODE_OF_CONDUCT.md` for project participation expectations

## License

Faultline is licensed under MIT. See `LICENSE` for the full text.


