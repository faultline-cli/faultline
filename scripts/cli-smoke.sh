#!/bin/sh

set -eu

ROOT_DIR="${ROOT_DIR:-$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)}"
BINARY="${BINARY:-$ROOT_DIR/bin/faultline}"
PLAYBOOK_DIR="${FAULTLINE_PLAYBOOK_DIR:-$ROOT_DIR/playbooks/bundled}"
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
grep -F '"pack_provenance":[{"name":"starter","playbook_count":173}]' "$TMP_DIR/missing.analysis.json" >/dev/null
cat "$ROOT_DIR/examples/runtime-mismatch.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze --json --no-history --git=false >"$TMP_DIR/runtime.analysis.json"

run_compare "missing-executable.replay.expected.md" "$ROOT_DIR/examples/missing-executable.replay.expected.md" \
	"$BINARY" replay --format markdown --mode detailed "$TMP_DIR/missing.analysis.json"
run_compare "missing-vs-runtime.compare.expected.md" "$ROOT_DIR/examples/missing-vs-runtime.compare.expected.md" \
	"$BINARY" compare --format markdown "$TMP_DIR/missing.analysis.json" "$TMP_DIR/runtime.analysis.json"

cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" trace --format markdown --playbook missing-executable --no-history --git=false >"$TMP_DIR/missing.trace.md"
compare_file "$TMP_DIR/missing.trace.md" "$ROOT_DIR/examples/missing-executable.trace.expected.md"

cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" workflow --no-history --git=false >"$TMP_DIR/workflow.local.txt"
compare_file "$TMP_DIR/workflow.local.txt" "$ROOT_DIR/examples/missing-executable.workflow.local.txt"

cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" workflow --json --mode agent --no-history --git=false >"$TMP_DIR/workflow.agent.json"
compare_file "$TMP_DIR/workflow.agent.json" "$ROOT_DIR/examples/missing-executable.workflow.agent.json"

cat >"$TMP_DIR/metrics-history.jsonl" <<'EOF'
{"matched":true,"failure_id":"missing-executable"}
{"matched":true,"failure_id":"missing-executable"}
{"matched":true,"failure_id":"missing-executable"}
{"matched":true,"failure_id":"missing-executable"}
{"matched":false}
EOF
cat "$ROOT_DIR/examples/missing-executable.log" | \
	FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" workflow --json --mode agent --no-history --git=false --metrics-history "$TMP_DIR/metrics-history.jsonl" >"$TMP_DIR/workflow.metrics.agent.json"
grep -F '"metrics_hints":' "$TMP_DIR/workflow.metrics.agent.json" >/dev/null
grep -F '"policy_hints":' "$TMP_DIR/workflow.metrics.agent.json" >/dev/null

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

SMOKE_REPO="$TMP_DIR/guard-repo"
mkdir -p "$SMOKE_REPO/api"
(
	cd "$SMOKE_REPO"
	git init -q
	git config user.name "Faultline Smoke"
	git config user.email "faultline@example.com"
	cat >"api/handler.go" <<'EOF'
package api

func UserHandler() string { return "ok" }
EOF
	git add .
	GIT_AUTHOR_DATE='2026-04-10 10:00:00 +0000' GIT_COMMITTER_DATE='2026-04-10 10:00:00 +0000' git commit --quiet -m "baseline: add handler"
	cat >"api/handler.go" <<'EOF'
package api

func UserHandler() string {
	panic("boom")
}
EOF
)

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" inspect --format markdown "$SMOKE_REPO" >"$TMP_DIR/inspect.md"
grep -F "panic-in-http-handler" "$TMP_DIR/inspect.md" >/dev/null

if FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" guard "$SMOKE_REPO" >"$TMP_DIR/guard.txt"; then
	printf '%s\n' "guard smoke expected findings with non-zero exit" >&2
	exit 1
fi
grep -F "panic-in-http-handler" "$TMP_DIR/guard.txt" >/dev/null

HOME="$TMP_DIR/home" FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" packs list >"$TMP_DIR/packs.txt"
grep -F "No installed playbook packs." "$TMP_DIR/packs.txt" >/dev/null

PACK_HOME="$TMP_DIR/pack-home"
HOME="$PACK_HOME" FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" packs install "$ROOT_DIR/examples/packs/minimal" --name example-pack >"$TMP_DIR/packs-install.txt"
printf '%s\n' "example cache prime missing" | \
	HOME="$PACK_HOME" FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" analyze --json --no-history --git=false >"$TMP_DIR/extra-pack.analysis.json"
grep -F '"failure_id":"example-cache-prime-missing"' "$TMP_DIR/extra-pack.analysis.json" >/dev/null
grep -F '"pack_provenance":[{"name":"starter","playbook_count":173},{"name":"example-pack","version":"0.0.0+local"' "$TMP_DIR/extra-pack.analysis.json" >/dev/null

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" coverage >"$TMP_DIR/coverage.txt"
grep -F "Playbook coverage report" "$TMP_DIR/coverage.txt" >/dev/null
grep -F "Total playbooks" "$TMP_DIR/coverage.txt" >/dev/null

FAULTLINE_PLAYBOOK_DIR="$PLAYBOOK_DIR" "$BINARY" coverage --json >"$TMP_DIR/coverage.json"
grep -F '"total_playbooks"' "$TMP_DIR/coverage.json" >/dev/null

printf '%s\n' "cli smoke passed"
