#!/bin/sh

set -eu

VERSION="${VERSION:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-dist/releases/$VERSION}"
TARGETS="${TARGETS:-darwin/amd64 linux/amd64}"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TMP_DIR="$ROOT_DIR/dist/tmp/$VERSION"

rm -rf "$TMP_DIR"
mkdir -p "$TMP_DIR" "$ROOT_DIR/$OUTPUT_DIR"

for target in $TARGETS; do
	goos="${target%/*}"
	goarch="${target#*/}"
	stage_dir="$TMP_DIR/faultline_${VERSION}_${goos}_${goarch}"
	archive_path="$ROOT_DIR/$OUTPUT_DIR/faultline_${VERSION}_${goos}_${goarch}.tar.gz"

	mkdir -p "$stage_dir"
	CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
		go build \
			-trimpath \
			-ldflags "-s -w -X main.version=$VERSION" \
			-o "$stage_dir/faultline" \
			./cmd

	cp -R "$ROOT_DIR/playbooks" "$stage_dir/playbooks"
	cp "$ROOT_DIR/README.md" "$stage_dir/README.md"

	tar -C "$TMP_DIR" -czf "$archive_path" "$(basename "$stage_dir")"
	printf '%s\n' "built $archive_path"
done
