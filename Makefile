SHELL := /bin/sh

GO ?= go
BINARY ?= bin/faultline
IMAGE ?= faultline
LOG ?=
VERSION ?= dev
RELEASE_OUTPUT ?= dist/releases/$(VERSION)
WITH_DOCKER ?= 0

.PHONY: help build run test bench review smoke-release docker-build docker-analyze docker-smoke release-snapshot release-check clean-dist

help:
	@printf "%s\n" "Targets:" \
		"  build           Build the faultline CLI into $(BINARY)" \
		"  run             Run the CLI locally: make run LOG=build.log" \
		"  test            Run all Go tests" \
		"  bench           Run bundled playbook load and analysis benchmarks" \
		"  review          Print bundled playbook pattern conflicts" \
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

smoke-release:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) sh ./scripts/smoke-release.sh

release-snapshot:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) ./scripts/release-build.sh

release-check: test review release-snapshot smoke-release
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
