SHELL := /bin/sh

GO ?= go
BINARY ?= bin/faultline
IMAGE ?= faultline
LOG ?=
VERSION ?= dev
RELEASE_OUTPUT ?= dist/releases/$(VERSION)
WITH_DOCKER ?= 0
PREMIUM_PACK_DIR ?=
PREMIUM_PACK_LINK ?= playbooks/packs/premium-local

.PHONY: help build run test bench review premium-path premium-link premium-check premium-review smoke-release docker-build docker-analyze docker-smoke release-snapshot release-check clean-dist

help:
	@printf "%s\n" "Targets:" \
		"  build           Build the faultline CLI into $(BINARY)" \
		"  run             Run the CLI locally: make run LOG=build.log" \
		"  test            Run all Go tests" \
		"  bench           Run bundled playbook load and analysis benchmarks" \
		"  review          Print bundled playbook pattern conflicts" \
		"  premium-path    Print the resolved premium pack path used locally" \
		"  premium-link    Create/update the ignored local premium-pack symlink" \
		"  premium-check   Compose starter with PREMIUM_PACK_DIR and fail on duplicate IDs or pack load errors" \
		"  premium-review  Compose starter with PREMIUM_PACK_DIR and print overlap conflicts across the combined catalog" \
		"  release-check   Run release-grade validation: tests, review, archive build, and smoke" \
		"  smoke-release   Verify a built release archive can run end to end" \
		"  release-snapshot  Build release tarballs into $(RELEASE_OUTPUT)" \
		"  clean-dist      Remove generated release artifacts" \
		"  docker-build    Build the Docker image tagged $(IMAGE)" \
		"  docker-analyze  Analyze a mounted log in Docker: make docker-analyze LOG=build.log" \
		"  docker-smoke    Build the Docker image and verify an auth fixture end to end" \
		"  WITH_DOCKER=1   Include docker-smoke when running release-check"

build:
	@mkdir -p "$$(dirname "$(BINARY)")"
	$(GO) build -o $(BINARY) ./cmd

run:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make run LOG=build.log"; exit 1; fi
	$(GO) run ./cmd analyze "$(LOG)"

test:
	$(GO) test ./...

bench:
	$(GO) test ./internal/engine -run '^$$' -bench 'Benchmark(LoadBundledPlaybooks|AnalyzeRepresentativeCorpus)' -benchmem

review:
	$(GO) run ./cmd/playbook-review

premium-path:
	@PREMIUM_PACK_DIR="$(PREMIUM_PACK_DIR)" sh ./scripts/resolve-premium-pack.sh

premium-link:
	@mkdir -p "$$(dirname "$(PREMIUM_PACK_LINK)")"
	@ln -sfn ../../../faultline-premium "$(PREMIUM_PACK_LINK)"
	@printf "%s\n" "linked $(PREMIUM_PACK_LINK) -> ../../../faultline-premium"

premium-check:
	@resolved="$$(PREMIUM_PACK_DIR="$(PREMIUM_PACK_DIR)" sh ./scripts/resolve-premium-pack.sh)" && \
	$(GO) run ./cmd/pack-compose-check --pack "$$resolved"

premium-review:
	@resolved="$$(PREMIUM_PACK_DIR="$(PREMIUM_PACK_DIR)" sh ./scripts/resolve-premium-pack.sh)" && \
	$(GO) run ./cmd/pack-compose-check --pack "$$resolved" --review

smoke-release:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) sh ./scripts/smoke-release.sh

release-snapshot:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) ./scripts/release-build.sh

release-check: test review release-snapshot smoke-release
	@if PREMIUM_PACK_DIR="$(PREMIUM_PACK_DIR)" sh ./scripts/resolve-premium-pack.sh >/dev/null 2>&1; then \
		$(MAKE) premium-check PREMIUM_PACK_DIR="$(PREMIUM_PACK_DIR)"; \
	else \
		printf "%s\n" "skipping premium-check (set PREMIUM_PACK_DIR, run make premium-link, or check out ../faultline-premium-pack)"; \
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
