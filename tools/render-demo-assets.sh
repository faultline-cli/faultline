#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
TAPE_DIR="$ROOT_DIR/docs/readme-assets/tapes"
OUTPUT_DIR="$ROOT_DIR/docs/readme-assets"
VHS_IMAGE=${VHS_IMAGE:-ghcr.io/charmbracelet/vhs:latest}
TOOLS_DIR="$ROOT_DIR/.tools/bin"

PATH="$TOOLS_DIR:$PATH"
export PATH

usage() {
	cat <<'EOF'
Usage: sh ./tools/render-demo-assets.sh [tape-name ...]

Renders README demo GIFs and screenshots from VHS tapes.

Dependencies:
- repo-local tools under `.tools/bin`
- or local `vhs` on PATH
- or `docker` on PATH for the official VHS container fallback

Examples:
  sh ./tools/render-demo-assets.sh
  sh ./tools/render-demo-assets.sh docker-auth-hero runtime-mismatch
EOF
}

have_command() {
	command -v "$1" >/dev/null 2>&1
}

run_vhs() {
	tape_path=$1
	if have_command vhs; then
		(
			cd "$ROOT_DIR"
			vhs "$tape_path"
		)
		return
	fi

	if have_command docker; then
		docker run --rm -v "$ROOT_DIR":/vhs "$VHS_IMAGE" "$tape_path"
		return
	fi

	printf "%s\n" "render-demo-assets requires either 'vhs' or 'docker' on PATH" >&2
	exit 1
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
	usage
	exit 0
fi

mkdir -p "$OUTPUT_DIR"

printf "%s\n" "building faultline for demo assets"
	(
		cd "$ROOT_DIR"
		make build
	)

if [ "$#" -eq 0 ]; then
	set -- docker-auth-hero docker-auth-fix docker-auth-modes missing-executable runtime-mismatch
fi

for tape_name in "$@"; do
	tape_file="$TAPE_DIR/$tape_name.tape"
	if [ ! -f "$tape_file" ]; then
		printf "%s\n" "unknown tape: $tape_name" >&2
		exit 1
	fi

	printf "%s\n" "rendering $tape_name"
	run_vhs "docs/readme-assets/tapes/$tape_name.tape"
done

printf "%s\n" "demo assets written to $OUTPUT_DIR"