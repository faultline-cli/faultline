SHELL := /bin/sh

GO ?= go
BINARY ?= bin/faultline
IMAGE ?= faultline
LOG ?=
VERSION ?= dev
RELEASE_OUTPUT ?= dist/releases/$(VERSION)
WITH_DOCKER ?= 0
.PHONY: help build run test fixture-check bayes-check bench review review-verbose review-update cli-smoke demo-assets smoke-release docker-build docker-analyze docker-smoke release-snapshot release-check release-verify clean-dist docs-generate docs-check stats-check dev-start dev-stop dev-login dev-auth-status dev-sync

# Local Teams dev — points at the adjacent faultline-teams repo.
# Override DEV_API_URL to target a different host (e.g. a remote staging API).
DEV_API_URL     ?= http://localhost:8787
DEV_API_PORT    ?= 8787
TEAMS_DIR       ?= ../faultline-teams
DEV_API_PIDFILE ?= /tmp/faultline-teams-dev.pid
DEV_API_LOG     ?= /tmp/faultline-teams-dev.log

help:
	@printf "%s\n" "Targets:" \
		"  build           Build the faultline CLI into $(BINARY)" \
		"  run             Run the CLI locally: make run LOG=build.log" \
		"  test            Run all Go tests" \
		"  demo-assets     Rebuild README GIFs and screenshots from VHS tapes" \
		"  fixture-check   Run the accepted real-fixture regression gate" \
		"  bayes-check     Compare baseline vs Bayes ranking across the real fixture corpus (pre-promotion gate)" \
		"  cli-smoke       Build the CLI and validate shipped examples plus companion commands" \
		"  bench           Run bundled playbook load and analysis benchmarks" \
		"  review          Verify bundled playbook pattern conflicts against the baseline" \
		"  review-verbose  Print the full bundled playbook pattern conflict report" \
		"  release-check   Run release-grade validation: tests, fixtures, Bayes, docs, review, archive build, and smoke" \
		"  smoke-release   Verify a built release archive can run end to end" \
		"  release-snapshot  Build release tarballs into $(RELEASE_OUTPUT)" \
		"  clean-dist      Remove generated release artifacts" \
		"  docker-build    Build the Docker image tagged $(IMAGE)" \
		"  docker-analyze  Analyze a mounted log in Docker: make docker-analyze LOG=build.log" \
		"  docker-smoke    Build the Docker image and verify an auth fixture end to end" \
		"  WITH_DOCKER=1   Include docker-smoke when running release-check" \
		"  docs-generate   Generate failure catalog docs from bundled playbooks" \
		"  docs-check      Verify generated failure catalog docs are up to date" \
		"  stats-check     Verify hardcoded playbook counts in README and llms.txt match the actual bundled set" \
		"" \
		"Local Teams dev (DEV_API_URL defaults to http://localhost:8787; override as needed):" \
		"  dev-start       Start the faultline-teams API in the background (no-op if already running)" \
		"  dev-stop        Stop the background faultline-teams API" \
		"  dev-login       Sign in (starts API automatically)" \
		"  dev-auth-status Check auth status (starts API automatically)" \
		"  dev-sync        Sync a failure artifact (starts API automatically): make dev-sync LOG=result.json PROJECT=<id>"

build:
	@mkdir -p "$$(dirname "$(BINARY)")"
	$(GO) build -o $(BINARY) ./cmd

demo-assets: build
	sh ./tools/render-demo-assets.sh

run:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make run LOG=build.log"; exit 1; fi
	$(GO) run ./cmd analyze "$(LOG)"

test:
	$(GO) test ./...

fixture-check:
	$(GO) run ./cmd fixtures stats --class real --check-baseline

bayes-check: build
	$(BINARY) fixtures compare-modes --class real

cli-smoke: build
	sh ./scripts/cli-smoke.sh

bench:
	$(GO) test ./internal/engine -run '^$$' -bench 'Benchmark(LoadBundledPlaybooks|AnalyzeRepresentativeCorpus)' -benchmem

docs-generate:
	$(GO) run ./tools/gen-failure-docs --src playbooks/bundled --dst docs/failures

docs-check: stats-check
	$(GO) run ./tools/gen-failure-docs --src playbooks/bundled --dst docs/failures --check

stats-check:
	@actual=$$(find playbooks/bundled -name '*.yaml' | wc -l | tr -d ' '); \
	for file in README.md llms.txt; do \
		matches=$$(grep -oE '[0-9]+ bundled playbooks?' "$$file" | grep -oE '^[0-9]+' | sort -u); \
		for n in $$matches; do \
			if [ "$$n" != "$$actual" ]; then \
				printf 'stats-check: %s contains playbook count %s but actual is %s\n' "$$file" "$$n" "$$actual" >&2; \
				exit 1; \
			fi; \
		done; \
	done; \
	if ! grep -F -- "- Bundled playbooks: $$actual" docs/fixture-corpus.md >/dev/null; then \
		printf 'stats-check: docs/fixture-corpus.md does not contain current bundled playbook count %s\n' "$$actual" >&2; \
		exit 1; \
	fi; \
	coverage=$$(sed -n 's/.*| Coverage | \*\*\([0-9.]*\)%\*\*.*/\1/p' docs/fixture-corpus.md | head -n 1); \
	if [ -n "$$coverage" ] && ! grep -F "coverage-$$coverage%25" README.md >/dev/null; then \
		printf 'stats-check: README.md coverage badge does not match docs/fixture-corpus.md coverage %s%%\n' "$$coverage" >&2; \
		exit 1; \
	fi; \
	printf 'stats-check: playbook count %s and coverage badge match published docs\n' "$$actual"

review:
	$(GO) run ./cmd fixtures patterns

review-verbose:
	$(GO) run ./cmd fixtures patterns --verbose

review-update:
	$(GO) run ./cmd fixtures patterns --update-baseline

smoke-release:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) sh ./scripts/smoke-release.sh

release-snapshot:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) ./scripts/release-build.sh

release-check: test fixture-check bayes-check review docs-check cli-smoke release-snapshot smoke-release
	@if [ "$(WITH_DOCKER)" = "1" ]; then \
		$(MAKE) docker-smoke IMAGE=$(IMAGE); \
	else \
		printf "%s\n" "skipping docker-smoke (set WITH_DOCKER=1 to include it)"; \
	fi

release-verify:
	VERSION=$(VERSION) WITH_DOCKER=$(WITH_DOCKER) sh ./tools/release-verify.sh

clean-dist:
	rm -rf dist

docker-build:
	docker build -t $(IMAGE) .

docker-analyze:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make docker-analyze LOG=build.log"; exit 1; fi
	docker run --rm -v "$$(pwd)":/workspace $(IMAGE) analyze "/workspace/$(LOG)"

docker-smoke:
	IMAGE=$(IMAGE) sh ./scripts/docker-smoke.sh

## ─── Local Teams dev ──────────────────────────────────────────────────────────
# DEV_API_URL defaults to http://localhost:8787 (wrangler dev default port).
# Override to target a different host: make dev-login DEV_API_URL=http://localhost:9000

dev-start: ## Start the faultline-teams API in the background; no-op if already running
	@if curl -sf $(DEV_API_URL)/health >/dev/null 2>&1; then \
		printf '%s\n' "Teams API already running at $(DEV_API_URL)"; \
	else \
		printf '%s\n' "Starting Teams API at $(DEV_API_URL) (log: $(DEV_API_LOG))..."; \
		cd $(TEAMS_DIR) && pnpm run dev:api </dev/null >$(DEV_API_LOG) 2>&1 & \
		printf '%d\n' "$$!" >$(DEV_API_PIDFILE); \
		i=0; until curl -sf $(DEV_API_URL)/health >/dev/null 2>&1; do \
			i=$$((i+1)); if [ $$i -ge 30 ]; then \
				printf '%s\n' "Timed out waiting for Teams API. Check $(DEV_API_LOG)"; exit 1; \
			fi; sleep 1; \
		done; \
		printf '%s\n' "Teams API ready."; \
	fi

dev-stop: ## Stop the background faultline-teams API started by dev-start
	@if [ -f $(DEV_API_PIDFILE) ]; then \
		pid="$$(cat $(DEV_API_PIDFILE))"; \
		if kill "$$pid" 2>/dev/null; then \
			printf '%s\n' "Teams API (pid $$pid) stopped."; \
		else \
			printf '%s\n' "Process $$pid already gone."; \
		fi; \
		rm -f $(DEV_API_PIDFILE); \
	else \
		printf '%s\n' "No managed Teams API process found ($(DEV_API_PIDFILE) missing)."; \
	fi

dev-login: build dev-start ## Sign in to the local faultline-teams dev server
	FAULTLINE_API_URL=$(DEV_API_URL) $(BINARY) auth login

dev-auth-status: build dev-start ## Check auth status against the local dev server
	FAULTLINE_API_URL=$(DEV_API_URL) $(BINARY) auth status

dev-sync: build dev-start ## Sync a failure artifact to local dev (LOG=artifact.json PROJECT=proj-id)
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, e.g.: make dev-sync LOG=result.json PROJECT=<id>"; exit 1; fi
	@if [ -z "$(PROJECT)" ]; then printf "%s\n" "PROJECT is required, e.g.: make dev-sync LOG=result.json PROJECT=<id>"; exit 1; fi
	FAULTLINE_API_URL=$(DEV_API_URL) $(BINARY) sync --project $(PROJECT) $(LOG)
