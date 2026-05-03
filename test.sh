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
VERSION=$(jq -r .version plugin.json)
if echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    pass "semver format ($VERSION)"
else
    fail "VERSION" "'$VERSION' is not valid semver"
fi

# ── Go build ────────────────────────────────────────────────────────

echo "build"
VERSION_LD="-X main.version=$VERSION"
if go build -ldflags "$VERSION_LD" -o dankcalendar ./cmd/dankcalendar 2>/dev/null; then
    pass "binary compiles"
else
    fail "build" "go build failed"
    echo ""
    echo "Results: $PASS passed, $FAIL failed"
    exit 1
fi

# ── CLI basics ──────────────────────────────────────────────────────

echo "CLI"
HELP=$(./dankcalendar --help 2>&1 || true)
assert_contains "help shows list command" "list" "$HELP"
assert_contains "help shows calendars command" "calendars" "$HELP"
assert_contains "help shows add command" "add" "$HELP"
assert_contains "help shows edit command" "edit" "$HELP"
assert_contains "help shows delete command" "delete" "$HELP"
assert_contains "help shows notify command" "notify" "$HELP"
assert_contains "help shows setup command" "setup" "$HELP"
assert_contains "help shows discover command" "discover" "$HELP"

VERSION_OUT=$(./dankcalendar --version 2>&1)
assert_eq "version output matches plugin.json" "$VERSION" "$VERSION_OUT"

# ── JSON error output ───────────────────────────────────────────────

echo "JSON output"
ERROR_OUT=$(./dankcalendar list 2>/dev/null || true)
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
for dir in cmd/dankcalendar internal/caldav internal/ical internal/keyring internal/config internal/output docs/adr; do
    if [ -d "$dir" ]; then
        pass "directory exists: $dir"
    else
        fail "structure" "missing directory: $dir"
    fi
done

for file in LICENSE .gitignore go.mod docs/adr/README.md plugin.json; do
    if [ -f "$file" ]; then
        pass "file exists: $file"
    else
        fail "structure" "missing file: $file"
    fi
done

# ── DMS plugin structure ───────────────────────────────────────────

echo "plugin"
for file in plugin.json CalendarWidget.qml CalendarSettings.qml; do
    if [ -f "$file" ]; then
        pass "plugin file exists: $file"
    else
        fail "plugin" "missing file: $file"
    fi
done

# Validate plugin.json fields
if [ -f plugin.json ]; then
    for field in id name type component settings; do
        if grep -q "\"$field\"" plugin.json; then
            pass "plugin.json has field: $field"
        else
            fail "plugin.json" "missing field: $field"
        fi
    done

    PLUGIN_ID=$(grep '"id"' plugin.json | sed 's/.*: *"\([^"]*\)".*/\1/')
    assert_eq "plugin.json id is dankCalendar" "dankCalendar" "$PLUGIN_ID"

    # pluginId consistency across QML files
    if grep -q "pluginId: \"$PLUGIN_ID\"" CalendarWidget.qml; then
        pass "CalendarWidget.qml pluginId matches plugin.json"
    else
        fail "pluginId" "CalendarWidget.qml pluginId does not match plugin.json id"
    fi

    if grep -q "pluginId: \"$PLUGIN_ID\"" CalendarSettings.qml; then
        pass "CalendarSettings.qml pluginId matches plugin.json"
    else
        fail "pluginId" "CalendarSettings.qml pluginId does not match plugin.json id"
    fi
fi

# ── ADR count ───────────────────────────────────────────────────────

echo "ADRs"
ADR_COUNT=$(ls docs/adr/ADR-*.md 2>/dev/null | wc -l)
if [ "$ADR_COUNT" -ge 5 ]; then
    pass "$ADR_COUNT ADRs present"
else
    fail "ADRs" "expected at least 5, found $ADR_COUNT"
fi

# ── Go unit tests ──────────────────────────────────────────────────

echo "unit tests"
if go test ./... > /dev/null 2>&1; then
    pass "go test ./... passes"
else
    fail "unit tests" "go test ./... failed"
fi

# ── cleanup ─────────────────────────────────────────────────────────

rm -f dankcalendar

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ] || exit 1
