#!/bin/sh

set -eu

VERSION="${VERSION:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-dist/releases/$VERSION}"
TARGET="${TARGET:-$(go env GOOS)/$(go env GOARCH)}"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TMP_DIR="$(mktemp -d)"

cleanup() {
	rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

goos="${TARGET%/*}"
goarch="${TARGET#*/}"
archive_path="$ROOT_DIR/$OUTPUT_DIR/faultline_${VERSION}_${goos}_${goarch}.tar.gz"

if [ ! -f "$archive_path" ]; then
	printf '%s\n' "release archive not found: $archive_path" >&2
	exit 1
fi

tar -C "$TMP_DIR" -xzf "$archive_path"
stage_dir="$TMP_DIR/faultline_${VERSION}_${goos}_${goarch}"
binary_path="$stage_dir/faultline"

if [ ! -x "$binary_path" ]; then
	printf '%s\n' "release binary is missing or not executable: $binary_path" >&2
	exit 1
fi

cat >"$TMP_DIR/docker-auth.log" <<'EOF'
## Step: smoke release archive
$ docker login ghcr.io
pull access denied
authentication required
EOF

output="$($binary_path analyze "$TMP_DIR/docker-auth.log")"
printf '%s\n' "$output"
printf '%s' "$output" | grep -F "docker-auth" >/dev/null
