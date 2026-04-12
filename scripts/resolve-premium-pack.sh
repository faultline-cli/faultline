#!/bin/sh
set -eu

resolve_dir() {
	cd "$1"
	pwd -P
}

if [ -n "${PREMIUM_PACK_DIR:-}" ]; then
	if [ -d "$PREMIUM_PACK_DIR" ]; then
		resolve_dir "$PREMIUM_PACK_DIR"
		exit 0
	fi
	printf '%s\n' "premium pack directory not found: $PREMIUM_PACK_DIR" >&2
	exit 1
fi

for candidate in \
	"playbooks/packs/premium-local" \
	"premium-pack" \
	"../faultline-premium-pack"
do
	if [ -d "$candidate" ]; then
		resolve_dir "$candidate"
		exit 0
	fi
done

printf '%s\n' "premium pack directory not found; set PREMIUM_PACK_DIR, run 'make premium-link', or check out ../faultline-premium-pack" >&2
	exit 1