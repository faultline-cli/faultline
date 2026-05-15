#!/bin/sh

set -eu

VERSION="${VERSION:-dev}"
WITH_DOCKER="${WITH_DOCKER:-0}"

printf '%s\n' "release verification: go test ./..."
go test ./...

printf '%s\n' "release verification: make fixture-check"
make fixture-check

printf '%s\n' "release verification: make bayes-check"
make bayes-check

printf '%s\n' "release verification: make review"
make review

printf '%s\n' "release verification: make docs-check"
make docs-check

printf '%s\n' "release verification: make cli-smoke"
make cli-smoke

if [ "$VERSION" != "dev" ]; then
	printf '%s\n' "release verification: make release-snapshot smoke-release"
	make release-snapshot VERSION="$VERSION"
	make smoke-release VERSION="$VERSION"
else
	printf '%s\n' "skipping release archive smoke (set VERSION=<tag> to include it)"
fi

if [ "$WITH_DOCKER" = "1" ]; then
	printf '%s\n' "release verification: make docker-smoke"
	make docker-smoke
else
	printf '%s\n' "skipping docker-smoke (set WITH_DOCKER=1 to include it)"
fi
