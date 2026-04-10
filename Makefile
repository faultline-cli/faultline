SHELL := /bin/sh

GO ?= go
BINARY ?= bin/faultline
IMAGE ?= faultline
LOG ?=
VERSION ?= dev
RELEASE_OUTPUT ?= dist/releases/$(VERSION)

.PHONY: help build run test review docker-build docker-analyze release-snapshot clean-dist

help:
	@printf "%s\n" "Targets:" \
		"  build           Build the faultline CLI into $(BINARY)" \
		"  run             Run the CLI locally: make run LOG=build.log" \
		"  test            Run all Go tests" \
		"  review          Print bundled playbook pattern conflicts" \
		"  release-snapshot  Build release tarballs into $(RELEASE_OUTPUT)" \
		"  clean-dist      Remove generated release artifacts" \
		"  docker-build    Build the Docker image tagged $(IMAGE)" \
		"  docker-analyze  Analyze a mounted log in Docker: make docker-analyze LOG=build.log"

build:
	@mkdir -p "$$(dirname "$(BINARY)")"
	$(GO) build -o $(BINARY) ./cmd

run:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make run LOG=build.log"; exit 1; fi
	$(GO) run ./cmd analyze "$(LOG)"

test:
	$(GO) test ./...

review:
	$(GO) run ./cmd/playbook-review

release-snapshot:
	VERSION=$(VERSION) OUTPUT_DIR=$(RELEASE_OUTPUT) ./scripts/release-build.sh

clean-dist:
	rm -rf dist

docker-build:
	docker build -t $(IMAGE) .

docker-analyze:
	@if [ -z "$(LOG)" ]; then printf "%s\n" "LOG is required, for example: make docker-analyze LOG=build.log"; exit 1; fi
	docker run --rm -v "$$(pwd)":/workspace $(IMAGE) analyze "/workspace/$(LOG)"
