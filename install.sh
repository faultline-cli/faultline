#!/bin/sh

set -eu

REPO="${REPO:-faultline-cli/faultline}"
API_ROOT="${API_ROOT:-https://api.github.com/repos/$REPO}"
RELEASE_ROOT="${RELEASE_ROOT:-https://github.com/$REPO/releases/download}"
VERSION="${VERSION:-}"
PREFIX="${PREFIX:-/usr/local}"
BIN_DIR="${BIN_DIR:-$PREFIX/bin}"
LIB_DIR="${LIB_DIR:-$PREFIX/lib/faultline}"
TMP_DIR="$(mktemp -d)"

cleanup() {
	rm -rf "$TMP_DIR"
}

trap cleanup EXIT INT TERM

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		printf '%s\n' "missing required command: $1" >&2
		exit 1
	fi
}

detect_os() {
	case "$(uname -s)" in
		Linux)
			printf '%s' "linux"
			;;
		Darwin)
			printf '%s' "darwin"
			;;
		*)
			printf '%s\n' "unsupported operating system: $(uname -s)" >&2
			exit 1
			;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64)
			printf '%s' "amd64"
			;;
		arm64|aarch64)
			printf '%s' "arm64"
			;;
		*)
			printf '%s\n' "unsupported architecture: $(uname -m)" >&2
			exit 1
			;;
	esac
}

resolve_version() {
	if [ -n "$VERSION" ]; then
		printf '%s' "$VERSION"
		return
	fi

	json_path="$TMP_DIR/latest-release.json"
	if ! curl -fsSL "$API_ROOT/releases/latest" -o "$json_path"; then
		printf '%s\n' "could not resolve the latest Faultline release. Publish a tagged release first or set VERSION=v0.1.0 explicitly." >&2
		exit 1
	fi

	resolved_version="$(sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$json_path" | head -n 1)"
	if [ -z "$resolved_version" ]; then
		printf '%s\n' "latest release metadata did not include a tag_name." >&2
		exit 1
	fi
	printf '%s' "$resolved_version"
}

install_path() {
	parent="$1"
	probe_parent="$parent"
	if [ ! -d "$probe_parent" ]; then
		probe_parent="$(dirname "$probe_parent")"
	fi
	if mkdir -p "$parent" 2>/dev/null; then
		probe="$parent/.faultline-write-test"
		if : >"$probe" 2>/dev/null; then
			rm -f "$probe"
			printf '%s' ""
			return
		fi
	fi
	if [ -w "$probe_parent" ] && [ -w "$parent" ]; then
		printf '%s' ""
		return
	fi
	if command -v sudo >/dev/null 2>&1; then
		printf '%s' "sudo"
		return
	fi
	printf '%s\n' "cannot write to $parent and sudo is unavailable" >&2
	exit 1
}

need_cmd curl
need_cmd tar
need_cmd uname
need_cmd mktemp
need_cmd sed
need_cmd dirname

os="$(detect_os)"
arch="$(detect_arch)"

case "$os/$arch" in
	linux/amd64|darwin/amd64|darwin/arm64)
		;;
	*)
		printf '%s\n' "no published Faultline release is available for $os/$arch" >&2
		exit 1
		;;
esac

version="$(resolve_version)"
archive_name="faultline_${version}_${os}_${arch}.tar.gz"
archive_url="$RELEASE_ROOT/$version/$archive_name"
archive_path="$TMP_DIR/$archive_name"
stage_dir="$TMP_DIR/faultline_${version}_${os}_${arch}"
sudo_prefix="$(install_path "$PREFIX")"

printf '%s\n' "Downloading $archive_url"
curl -fL "$archive_url" -o "$archive_path"
tar -C "$TMP_DIR" -xzf "$archive_path"

if [ ! -x "$stage_dir/faultline" ]; then
	printf '%s\n' "downloaded archive is missing the faultline binary" >&2
	exit 1
fi

if [ -z "$sudo_prefix" ]; then
	mkdir -p "$BIN_DIR" "$LIB_DIR"
	rm -rf "$LIB_DIR/playbooks"
	cp "$stage_dir/faultline" "$LIB_DIR/faultline"
	cp -R "$stage_dir/playbooks" "$LIB_DIR/playbooks"
	cat >"$BIN_DIR/faultline" <<EOF
#!/bin/sh
export FAULTLINE_PLAYBOOK_DIR="$LIB_DIR/playbooks/bundled"
exec "$LIB_DIR/faultline" "\$@"
EOF
	chmod 755 "$LIB_DIR/faultline" "$BIN_DIR/faultline"
else
	$sudo_prefix mkdir -p "$BIN_DIR" "$LIB_DIR"
	$sudo_prefix rm -rf "$LIB_DIR/playbooks"
	$sudo_prefix cp "$stage_dir/faultline" "$LIB_DIR/faultline"
	$sudo_prefix cp -R "$stage_dir/playbooks" "$LIB_DIR/playbooks"
	cat >"$TMP_DIR/faultline-wrapper" <<EOF
#!/bin/sh
export FAULTLINE_PLAYBOOK_DIR="$LIB_DIR/playbooks/bundled"
exec "$LIB_DIR/faultline" "\$@"
EOF
	$sudo_prefix cp "$TMP_DIR/faultline-wrapper" "$BIN_DIR/faultline"
	$sudo_prefix chmod 755 "$LIB_DIR/faultline" "$BIN_DIR/faultline"
fi

printf '%s\n' "Installed faultline $version to $BIN_DIR/faultline"
printf '%s\n' "Run: faultline analyze ci.log"