SHELL := /bin/sh

GO ?= go
BINARY ?= bin/faultline
IMAGE ?= faultline
LOG ?=
VERSION ?= dev
RELEASE_OUTPUT ?= dist/releases/$(VERSION)
WITH_DOCKER ?= 0
.PHONY: help build run test fixture-check bayes-check bench review review-verbose review-update cli-smoke demo-assets smoke-release docker-build docker-analyze docker-smoke release-snapshot release-check release-verify clean-dist docs-generate docs-check stats-check

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
		"  stats-check     Verify hardcoded playbook counts in README and llms.txt match the actual bundled set"

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
	printf 'stats-check: playbook count %s matches README.md and llms.txt\n' "$$actual"

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
