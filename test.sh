#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0

pass() { PASS=$((PASS + 1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "  FAIL: $1 — $2"; }

assert_eq() {
    local desc="$1" expected="$2" actual="$3"
    if [ "$expected" = "$actual" ]; then
        pass "$desc"
    else
        fail "$desc" "expected '$expected', got '$actual'"
    fi
}

assert_contains() {
    local desc="$1" needle="$2" haystack="$3"
    if echo "$haystack" | grep -qF "$needle"; then
        pass "$desc"
    else
        fail "$desc" "expected to contain '$needle'"
    fi
}

# ── VERSION format ──────────────────────────────────────────────────

echo "VERSION"
VERSION=$(cat VERSION | tr -d '[:space:]')
if echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    pass "semver format ($VERSION)"
else
    fail "VERSION" "'$VERSION' is not valid semver"
fi

# ── Go build ────────────────────────────────────────────────────────

echo "build"
VERSION_LD="-X main.version=$VERSION"
if go build -ldflags "$VERSION_LD" -o dankcal ./cmd/dankcal 2>/dev/null; then
    pass "binary compiles"
else
    fail "build" "go build failed"
    echo ""
    echo "Results: $PASS passed, $FAIL failed"
    exit 1
fi

# ── CLI basics ──────────────────────────────────────────────────────

echo "CLI"
HELP=$(./dankcal --help 2>&1 || true)
assert_contains "help shows list command" "list" "$HELP"
assert_contains "help shows calendars command" "calendars" "$HELP"
assert_contains "help shows add command" "add" "$HELP"
assert_contains "help shows edit command" "edit" "$HELP"
assert_contains "help shows delete command" "delete" "$HELP"
assert_contains "help shows notify command" "notify" "$HELP"
assert_contains "help shows setup command" "setup" "$HELP"

VERSION_OUT=$(./dankcal --version 2>&1)
assert_eq "version output matches VERSION file" "$VERSION" "$VERSION_OUT"

# ── JSON error output ───────────────────────────────────────────────

echo "JSON output"
ERROR_OUT=$(./dankcal list 2>/dev/null || true)
assert_contains "error output is JSON" '"error"' "$ERROR_OUT"

# ── stdlib-only (no go.sum) ─────────────────────────────────────────

echo "dependencies"
if [ -f go.sum ]; then
    fail "stdlib-only" "go.sum exists — external dependencies detected"
else
    pass "stdlib-only (no go.sum)"
fi

# ── project structure ───────────────────────────────────────────────

echo "project structure"
for dir in cmd/dankcal internal/caldav internal/ical internal/keyring internal/config internal/notify internal/output docs/adr; do
    if [ -d "$dir" ]; then
        pass "directory exists: $dir"
    else
        fail "structure" "missing directory: $dir"
    fi
done

for file in VERSION LICENSE .gitignore go.mod docs/adr/README.md; do
    if [ -f "$file" ]; then
        pass "file exists: $file"
    else
        fail "structure" "missing file: $file"
    fi
done

# ── ADR count ───────────────────────────────────────────────────────

echo "ADRs"
ADR_COUNT=$(ls docs/adr/ADR-*.md 2>/dev/null | wc -l)
if [ "$ADR_COUNT" -ge 5 ]; then
    pass "$ADR_COUNT ADRs present"
else
    fail "ADRs" "expected at least 5, found $ADR_COUNT"
fi

# ── cleanup ─────────────────────────────────────────────────────────

rm -f dankcal

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
