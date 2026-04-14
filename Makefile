SHELL := /bin/sh

GO ?= go
BINARY ?= bin/faultline
IMAGE ?= faultline
LOG ?=
VERSION ?= dev
RELEASE_OUTPUT ?= dist/releases/$(VERSION)
WITH_DOCKER ?= 0
EXTRA_PACK_DIR ?=
EXTRA_PACK_LINK ?= playbooks/packs/extra-local

.PHONY: help build run test fixture-check bench review extra-pack-path extra-pack-link extra-pack-check extra-pack-review smoke-release docker-build docker-analyze docker-smoke release-snapshot release-check clean-dist

help:
	@printf "%s\n" "Targets:" \
		"  build           Build the faultline CLI into $(BINARY)" \
		"  run             Run the CLI locally: make run LOG=build.log" \
		"  test            Run all Go tests" \
		"  fixture-check   Run the accepted real-fixture regression gate" \
		"  bench           Run bundled playbook load and analysis benchmarks" \
		"  review          Print bundled playbook pattern conflicts" \
		"  release-check   Run release-grade validation: tests, review, archive build, and smoke" \
		"  smoke-release   Verify a built release archive can run end to end" \
		"  release-snapshot  Build release tarballs into $(RELEASE_OUTPUT)" \
		"  clean-dist      Remove generated release artifacts" \
		"  docker-build    Build the Docker image tagged $(IMAGE)" \
		"  docker-analyze  Analyze a mounted log in Docker: make docker-analyze LOG=build.log" \
		"  docker-smoke    Build the Docker image and verify an auth fixture end to end" \
		"  WITH_DOCKER=1   Include docker-smoke when running release-check" \
		"" \
		"Internal pack-composition targets remain available but are intentionally omitted from the public help summary."

build:
	@mkdir -p "$$(dirname "$(BINARY)")"
	$(GO) build -o $(BINARY) ./cmd

run:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make run LOG=build.log"; exit 1; fi
	$(GO) run ./cmd analyze "$(LOG)"

test:
	$(GO) test ./...

fixture-check:
	$(GO) run ./cmd fixtures stats --class real --check-baseline

bench:
	$(GO) test ./internal/engine -run '^$$' -bench 'Benchmark(LoadBundledPlaybooks|AnalyzeRepresentativeCorpus)' -benchmem

review:
	$(GO) run ./cmd/playbook-review

extra-pack-path:
	@EXTRA_PACK_DIR="$(EXTRA_PACK_DIR)" sh ./scripts/resolve-extra-pack.sh

extra-pack-link:
	@mkdir -p "$$(dirname "$(EXTRA_PACK_LINK)")"
	@ln -sfn ../../../faultline-extra-pack "$(EXTRA_PACK_LINK)"
	@printf "%s\n" "linked $(EXTRA_PACK_LINK) -> ../../../faultline-extra-pack"

extra-pack-check:
	@resolved="$$(EXTRA_PACK_DIR="$(EXTRA_PACK_DIR)" sh ./scripts/resolve-extra-pack.sh)" && \
	$(GO) run ./cmd/pack-compose-check --pack "$$resolved"

extra-pack-review:
	@resolved="$$(EXTRA_PACK_DIR="$(EXTRA_PACK_DIR)" sh ./scripts/resolve-extra-pack.sh)" && \
	$(GO) run ./cmd/pack-compose-check --pack "$$resolved" --review

smoke-release:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) sh ./scripts/smoke-release.sh

release-snapshot:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) ./scripts/release-build.sh

release-check: test fixture-check review release-snapshot smoke-release
	@if EXTRA_PACK_DIR="$(EXTRA_PACK_DIR)" sh ./scripts/resolve-extra-pack.sh >/dev/null 2>&1; then \
		$(MAKE) extra-pack-check EXTRA_PACK_DIR="$(EXTRA_PACK_DIR)"; \
	else \
		printf "%s\n" "skipping extra-pack-check (set EXTRA_PACK_DIR, run make extra-pack-link, or set an explicit extra pack path)"; \
	fi
	@if [ "$(WITH_DOCKER)" = "1" ]; then \
		$(MAKE) docker-smoke IMAGE=$(IMAGE); \
	else \
		printf "%s\n" "skipping docker-smoke (set WITH_DOCKER=1 to include it)"; \
	fi

clean-dist:
	rm -rf dist

docker-build:
	docker build -t $(IMAGE) .

docker-analyze:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make docker-analyze LOG=build.log"; exit 1; fi
	docker run --rm -v "$$(pwd)":/workspace $(IMAGE) analyze "/workspace/$(LOG)"

docker-smoke:
	IMAGE=$(IMAGE) sh ./scripts/docker-smoke.sh
