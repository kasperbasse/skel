#!/usr/bin/env bash
set -euo pipefail

# End-to-end smoke test for skel commands using an isolated HOME.
#
# Usage:
#   scripts/smoke_commands.sh
#   SKEL_BIN=/path/to/skel scripts/smoke_commands.sh
#   KEEP_SMOKE_TMP=1 scripts/smoke_commands.sh

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_ROOT="$(mktemp -d)"
TMP_HOME="$TMP_ROOT/home"
TMP_WORK="$TMP_ROOT/work"
mkdir -p "$TMP_HOME" "$TMP_WORK"

BIN="${SKEL_BIN:-$TMP_ROOT/skel}"
LAST_OUTPUT=""

cleanup() {
  if [[ "${KEEP_SMOKE_TMP:-0}" == "1" ]]; then
    echo "Keeping smoke temp dir: $TMP_ROOT"
    return
  fi
  rm -rf "$TMP_ROOT"
}
trap cleanup EXIT

if [[ -z "${SKEL_BIN:-}" ]]; then
  (cd "$ROOT_DIR" && go build -o "$BIN" .)
fi

if [[ ! -x "$BIN" ]]; then
  echo "ERROR: skel binary is not executable: $BIN" >&2
  exit 1
fi

echo "Using skel binary: $BIN"
echo "Using isolated HOME: $TMP_HOME"
echo "Working dir: $TMP_WORK"

strip_ansi() {
  sed -E 's/\x1b\[[0-9;]*[A-Za-z]//g'
}

run_ok() {
  local label="$1"
  shift

  echo
  echo "==> $label"
  echo "+ $*"

  local output
  set +e
  output="$(HOME="$TMP_HOME" "$@" 2>&1)"
  local rc=$?
  set -e

  echo "$output"
  LAST_OUTPUT="$(printf '%s' "$output" | strip_ansi)"

  if [[ $rc -ne 0 ]]; then
    echo "FAILED: expected success, got exit $rc" >&2
    exit 1
  fi
}

run_fail() {
  local label="$1"
  shift

  echo
  echo "==> $label"
  echo "+ $*"

  local output
  set +e
  output="$(HOME="$TMP_HOME" "$@" 2>&1)"
  local rc=$?
  set -e

  echo "$output"
  LAST_OUTPUT="$(printf '%s' "$output" | strip_ansi)"

  if [[ $rc -eq 0 ]]; then
    echo "FAILED: expected failure, command succeeded" >&2
    exit 1
  fi
}

run_ok_with_input() {
  local label="$1"
  local input="$2"
  shift 2

  echo
  echo "==> $label"
  echo "+ (input piped) $*"

  local output
  set +e
  output="$(printf '%b' "$input" | HOME="$TMP_HOME" "$@" 2>&1)"
  local rc=$?
  set -e

  echo "$output"
  LAST_OUTPUT="$(printf '%s' "$output" | strip_ansi)"

  if [[ $rc -ne 0 ]]; then
    echo "FAILED: expected success, got exit $rc" >&2
    exit 1
  fi
}

run_ok_in_dir() {
  local label="$1"
  local dir="$2"
  shift 2

  echo
  echo "==> $label"
  echo "+ (cd $dir && $*)"

  local output
  set +e
  output="$(cd "$dir" && HOME="$TMP_HOME" "$@" 2>&1)"
  local rc=$?
  set -e

  echo "$output"
  LAST_OUTPUT="$(printf '%s' "$output" | strip_ansi)"

  if [[ $rc -ne 0 ]]; then
    echo "FAILED: expected success, got exit $rc" >&2
    exit 1
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local context="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    echo "ASSERT FAILED ($context): expected to find '$needle'" >&2
    exit 1
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local context="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    echo "ASSERT FAILED ($context): expected not to find '$needle'" >&2
    exit 1
  fi
}

assert_contains_any() {
  local haystack="$1"
  local context="$2"
  shift 2

  local needle
  for needle in "$@"; do
    if [[ "$haystack" == *"$needle"* ]]; then
      return
    fi
  done

  echo "ASSERT FAILED ($context): expected to find one of: $*" >&2
  exit 1
}

assert_file_exists() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    echo "ASSERT FAILED: expected file to exist: $file" >&2
    exit 1
  fi
}

PROFILE_DIR="$TMP_HOME/.skel/profiles"
FIXTURE_JSON="$TMP_WORK/fixture-profile.json"
BREWFILE_OUT="$TMP_WORK/Brewfile"
EXPORT_JSON="$TMP_WORK/fixture-skel.json"

cat >"$FIXTURE_JSON" <<'JSON'
{
  "name": "fixture",
  "created_at": "2026-04-01T00:00:00Z",
  "machine": "smoke-machine",
  "homebrew": {
    "formulas": ["ripgrep"],
    "casks": [],
    "mas_apps": []
  },
  "shell": {
    "shell": "zsh",
    "zshrc_content": "export TEST_SMOKE=1",
    "oh_my_zsh": false,
    "starship": false
  },
  "editor": {
    "vscode": false,
    "cursor": false,
    "neovim": false
  },
  "git": {
    "user_name": "",
    "user_email": ""
  },
  "languages": {
    "npm_globals": ["typescript"]
  },
  "system": {
    "hostname": "smoke-host",
    "macos_version": "14.0",
    "chip_arch": "arm64"
  }
}
JSON

# Help coverage for top-level and subcommands.
run_ok "root help" "$BIN" --help
for cmd_help in \
  "scan --help" \
  "restore --help" \
  "list --help" \
  "show --help" \
  "update --help" \
  "delete --help" \
  "status --help" \
  "drift --help" \
  "diff --help" \
  "doctor --help" \
  "export --help" \
  "import --help" \
  "clone --help" \
  "publish --help" \
  "brewfile --help" \
  "brewfile export --help" \
  "brewfile import --help"; do
  # shellcheck disable=SC2086
  run_ok "$cmd_help" "$BIN" $cmd_help
done

# Import fixture and validate profile file existence.
run_ok "import fixture profile" "$BIN" import "$FIXTURE_JSON"
assert_file_exists "$PROFILE_DIR/fixture.json"

run_ok "list profiles" "$BIN" list
assert_contains "$LAST_OUTPUT" "fixture" "list includes imported fixture"

run_ok "show fixture" "$BIN" show fixture
assert_contains "$LAST_OUTPUT" "Profile: 'fixture'" "show renders fixture header"

run_ok "status fixture" "$BIN" status fixture
assert_contains "$LAST_OUTPUT" "Status: 'fixture'" "status renders fixture"

run_ok "doctor fixture" "$BIN" doctor fixture
assert_contains "$LAST_OUTPUT" "Checking 'fixture'" "doctor runs against fixture"

run_ok "drift fixture" "$BIN" drift fixture
assert_contains "$LAST_OUTPUT" "Checking for drift against 'fixture'" "drift runs"

run_ok "diff fixture fixture" "$BIN" diff fixture fixture
assert_contains "$LAST_OUTPUT" "identical" "diff identical path"

# Export + import round-trip.
run_ok_in_dir "export fixture" "$TMP_WORK" "$BIN" export fixture
assert_file_exists "$EXPORT_JSON"

run_ok "import exported fixture" "$BIN" import "$EXPORT_JSON"
assert_file_exists "$PROFILE_DIR/fixture.json"

# Brewfile flow.
run_ok "brewfile export fixture" "$BIN" brewfile export fixture --output "$BREWFILE_OUT"
assert_file_exists "$BREWFILE_OUT"

run_ok "brewfile import as brewfixture" "$BIN" brewfile import "$BREWFILE_OUT" --name brewfixture
assert_file_exists "$PROFILE_DIR/brewfixture.json"

# Restore dry-run must stay non-destructive.
run_ok "restore fixture dry-run" "$BIN" restore fixture --dry-run --only shell
assert_contains "$LAST_OUTPUT" "Dry run" "restore dry-run banner"

# Scan create + overwrite confirmation with y.
run_ok "scan smoke --force" "$BIN" scan smoke --force
assert_file_exists "$PROFILE_DIR/smoke.json"

before_mtime="$(stat -f %m "$PROFILE_DIR/smoke.json")"
sleep 1
run_ok_with_input "scan smoke overwrite using y" "y\n" "$BIN" scan smoke
assert_contains "$LAST_OUTPUT" "already exists" "scan overwrite prompt shown"
assert_not_contains "$LAST_OUTPUT" "Canceled." "scan overwrite should not cancel on y"
after_mtime="$(stat -f %m "$PROFILE_DIR/smoke.json")"
if [[ "$after_mtime" -le "$before_mtime" ]]; then
  echo "ASSERT FAILED: smoke profile timestamp did not increase after overwrite" >&2
  exit 1
fi

run_ok "update smoke" "$BIN" update smoke
assert_contains "$LAST_OUTPUT" "updated" "update success"

# delete uses interactive TUI confirmation, so validate safe failure path in non-interactive smoke.
run_fail "delete non-existing profile" "$BIN" delete does-not-exist
assert_contains "$LAST_OUTPUT" "not found" "delete failure path"

# Network commands: validate safe failure paths without external dependencies.
run_fail "clone invalid source" "$BIN" clone not-a-gist
assert_contains "$LAST_OUTPUT" "unrecognized source" "clone parse validation"

run_fail "publish without auth" env -u GITHUB_TOKEN -u GH_TOKEN PATH="/usr/bin:/bin" "$BIN" publish fixture
assert_contains_any "$LAST_OUTPUT" "publish auth validation" "GITHUB_TOKEN" "gh auth login" "not installed"

rm -f "$EXPORT_JSON"

echo
echo "Smoke test passed: command matrix and file outcomes validated."

