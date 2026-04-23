# ADR-001: Single Go binary as the CalDAV backend

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `cmd/dankcalendar/main.go`, entire project

## Context

DankMaterialShell plugins call external tools as subprocesses and parse their stdout. A CalDAV calendar widget therefore needs a backend that speaks CalDAV and outputs structured data. The straightforward implementation — a Python script wrapping a separate Go binary — creates two process hops per operation, requires a Python runtime, and introduces an external binary as a coupling point for upstream changes.

## Decision

Implement CalDAV natively in a single Go binary (`dankcalendar`) that outputs JSON directly. QML calls `dankcalendar` as a subprocess and parses its stdout via `SplitParser` — exactly one process hop, no interpreter required.

## Alternatives Considered

- **Python wrapper around a Go CalDAV binary**: Still requires a Python runtime and two process hops. Doesn't simplify the dependency chain.
- **Python only**: Loses Go's static binary advantage and straightforward cross-compilation.
- **Use an existing CalDAV Go library**: Adds external dependencies; the CalDAV subset needed (REPORT, PROPFIND, PUT, DELETE) is small enough to implement with stdlib HTTP.

## Consequences

- Single static binary with no runtime dependencies beyond `secret-tool` and `notify-send`.
- Plugin `requires` in `plugin.json` lists only runtime tools, no language runtimes.
- All CalDAV logic, ICS generation, credential handling, and JSON serialisation live in one self-contained, testable codebase.
