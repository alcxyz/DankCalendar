# ADR-001: Single Go binary replaces Go+Python stack

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `cmd/dankcalendar/main.go`, entire project

## Context

The existing `dms-qcal-calendar` plugin uses a three-layer architecture: QML UI -> Python wrapper (`qcal-wrapper.py`, 945 lines) -> Go binary (`qcal` submodule from psic4t). The Python layer parses qcal's text output into JSON, handles CalDAV discovery, manages credentials, and generates ICS files. This creates two process hops per operation, requires Python 3 as a runtime dependency, and makes the qcal submodule a coupling point for upstream changes.

## Decision

Replace both the qcal Go submodule and the Python wrapper with a single Go binary (`dankcalendar`) that speaks CalDAV natively and outputs JSON directly. QML calls dankcalendar as a subprocess and parses its stdout via SplitParser — exactly one process hop.

## Alternatives Considered

- **Keep Python wrapper, replace only qcal**: Still requires Python runtime and two process hops. Doesn't simplify the dependency chain.
- **Rewrite in Python only**: Loses Go's static binary advantage and the easy cross-compilation for different architectures.
- **Use an existing CalDAV Go library**: Adds external dependencies; the CalDAV subset we need (REPORT, PROPFIND, PUT, DELETE) is small enough for stdlib HTTP.

## Consequences

- Single static binary with no runtime dependencies beyond `secret-tool` and `notify-send`.
- Plugin `requires` field in plugin.json changes from `["python3"]` to `[]` (or `["go"]` build-time only).
- All CalDAV logic, ICS generation, credential handling, and JSON serialization live in one codebase.
- Upstream qcal changes no longer affect this plugin.
- More Go code to maintain, but it is self-contained and testable.
