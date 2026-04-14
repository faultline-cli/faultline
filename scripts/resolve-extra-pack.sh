#!/bin/sh
set -eu

resolve_dir() {
	cd "$1"
	pwd -P
}

if [ -n "${EXTRA_PACK_DIR:-}" ]; then
	if [ -d "$EXTRA_PACK_DIR" ]; then
		resolve_dir "$EXTRA_PACK_DIR"
		exit 0
	fi
	printf '%s\n' "extra pack directory not found: $EXTRA_PACK_DIR" >&2
	exit 1
fi

for candidate in \
	"playbooks/packs/extra-local" \
	"faultline-extra-pack" \
	"extra-pack" \
	"../faultline-extra" \
	"../faultline-extra-pack"
do
	if [ -d "$candidate" ]; then
		resolve_dir "$candidate"
		exit 0
	fi
done

printf '%s\n' "extra pack directory not found; set EXTRA_PACK_DIR, run 'make extra-pack-link', or set an explicit extra pack path" >&2
	exit 1