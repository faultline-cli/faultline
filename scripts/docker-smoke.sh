#!/bin/sh

set -eu

IMAGE="${IMAGE:-faultline-smoke}"
ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TMP_DIR="$(mktemp -d)"

cleanup() {
	rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

cat >"$TMP_DIR/docker-auth.log" <<'EOF'
## Step: smoke docker image
$ docker login ghcr.io
pull access denied
authentication required
EOF

chmod 755 "$TMP_DIR"
chmod 644 "$TMP_DIR/docker-auth.log"

docker build -t "$IMAGE" "$ROOT_DIR"
output="$(docker run --rm -v "$TMP_DIR":/workspace "$IMAGE" analyze /workspace/docker-auth.log)"
printf '%s\n' "$output"
printf '%s' "$output" | grep -F "docker-auth" >/dev/null
