#!/bin/sh

set -eu

ROOT_DIR="${ROOT_DIR:-$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)}"
BINARY="${BINARY:-$ROOT_DIR/bin/faultline}"
PLAYBOOK_DIR="${FAULTLINE_PLAYBOOK_DIR:-$ROOT_DIR/playbooks/bundled}"
STARTER_PLAYBOOK_COUNT="${STARTER_PLAYBOOK_COUNT:-182}"
TMP_DIR="$(mktemp -d)"

cleanup() {
	rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

case "$BINARY" in
	/*) ;;
	*) BINARY="$ROOT_DIR/$BINARY" ;;
esac

if [ ! -x "$BINARY" ]; then
	printf '%s\n' "faultline binary is missing or not executable: $BINARY" >&2
	exit 1
fi

run_compare() {
	label="$1"
	expected="$2"
	shift 2
	got="$TMP_DIR/$label"
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$@" >"$got"
	if ! cmp -s "$got" "$expected"; then
		diff -u "$expected" "$got" >&2 || true
		return 1
	fi
}

compare_file() {
	got="$1"
	expected="$2"
	if ! cmp -s "$got" "$expected"; then
		diff -u "$expected" "$got" >&2 || true
		return 1
	fi
}

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze "$ROOT_DIR/examples/docker-auth.log" --no-history >"$TMP_DIR/analyze.txt"
grep -F "docker-auth" "$TMP_DIR/analyze.txt" >/dev/null

run_compare "docker-auth.expected.md" "$ROOT_DIR/examples/docker-auth.expected.md" \
	"$BINARY" analyze "$ROOT_DIR/examples/docker-auth.log" --format markdown --no-history --git=false
run_compare "lockfile-drift.expected.md" "$ROOT_DIR/examples/lockfile-drift.expected.md" \
	"$BINARY" analyze "$ROOT_DIR/examples/lockfile-drift.log" --format markdown --no-history --git=false
run_compare "missing-executable.expected.md" "$ROOT_DIR/examples/missing-executable.expected.md" \
	"$BINARY" analyze "$ROOT_DIR/examples/missing-executable.log" --format markdown --no-history --git=false
run_compare "runtime-mismatch.expected.md" "$ROOT_DIR/examples/runtime-mismatch.expected.md" \
	"$BINARY" analyze "$ROOT_DIR/examples/runtime-mismatch.log" --format markdown --no-history --git=false
(
	cd "$ROOT_DIR"
	run_compare "batch.expected.md" "$ROOT_DIR/examples/batch.expected.md" \
		"$BINARY" batch examples/missing-executable.log examples/runtime-mismatch.log --format markdown --no-history
)

cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze --json --no-history --git=false >"$TMP_DIR/missing.analysis.json"
jq -e --argjson expected_count "$STARTER_PLAYBOOK_COUNT" \
	'.pack_provenance | length == 1 and .[0].name == "starter" and .[0].playbook_count == $expected_count' \
	"$TMP_DIR/missing.analysis.json" >/dev/null
cat "$ROOT_DIR/examples/runtime-mismatch.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze --json --no-history --git=false >"$TMP_DIR/runtime.analysis.json"

cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" workflow --no-history --git=false >"$TMP_DIR/workflow.local.txt"
compare_file "$TMP_DIR/workflow.local.txt" "$ROOT_DIR/examples/missing-executable.workflow.local.txt"

cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" workflow --json --mode agent --no-history --git=false >"$TMP_DIR/workflow.agent.json"
compare_file "$TMP_DIR/workflow.agent.json" "$ROOT_DIR/examples/missing-executable.workflow.agent.json"

STORE_PATH="$TMP_DIR/faultline-store.db"
FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze "$ROOT_DIR/examples/missing-executable.log" --store "$STORE_PATH" --git=false >/dev/null
FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze "$ROOT_DIR/examples/missing-executable.log" --store "$STORE_PATH" --git=false >/dev/null
FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" report --store "$STORE_PATH" --json >"$TMP_DIR/report.json"
grep -F '"failure_id": "missing-executable"' "$TMP_DIR/report.json" >/dev/null
grep -F '"count": 2' "$TMP_DIR/report.json" >/dev/null

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" explain docker-auth >"$TMP_DIR/explain.txt"
grep -F "docker-auth" "$TMP_DIR/explain.txt" >/dev/null

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" list >"$TMP_DIR/list.txt"
grep -F "docker-auth" "$TMP_DIR/list.txt" >/dev/null

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" fix "$ROOT_DIR/examples/docker-auth.log" --format markdown --no-history >"$TMP_DIR/fix.md"
grep -F "## Fix" "$TMP_DIR/fix.md" >/dev/null

printf '%s\n' "example cache prime missing" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze --json --no-history --git=false --playbook-pack "$ROOT_DIR/examples/packs/minimal" >"$TMP_DIR/extra-pack.analysis.json"
grep -F '"failure_id":"example-cache-prime-missing"' "$TMP_DIR/extra-pack.analysis.json" >/dev/null

printf '%s\n' "cli smoke passed"
