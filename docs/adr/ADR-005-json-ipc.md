# ADR-005: JSON-over-stdout IPC with QML

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `internal/output/`, `cmd/dankcal/`

## Context

The QML plugin consumes subprocess output via SplitParser, which reads stdout line-by-line. The existing Python wrapper outputs JSON objects that QML parses with `JSON.parse()`. This pattern works well — it is simple, debuggable (pipe to `jq`), and requires no IPC framework.

## Decision

All dankcal commands output a single JSON object to stdout. Errors produce `{"error": "message"}` with a non-zero exit code. Diagnostic/progress messages go to stderr only. No interactive prompts — all input comes via CLI flags or stdin (for `setup` password entry).

Command output contracts:
- `list`: `{"events": [...], "count": N}`
- `calendars`: `{"calendars": [...]}`
- `add/edit/delete`: `{"success": true}` or `{"error": "..."}`
- `notify`: `{"notified": N}`

## Alternatives Considered

- **D-Bus service**: More "desktop native" but far more complex, requires a running daemon, and doesn't match the existing SplitParser pattern.
- **Unix socket / gRPC**: Overkill for a widget that refreshes every few minutes.
- **Structured text (key=value)**: Harder to parse reliably than JSON, no nested data support.

## Consequences

- QML integration requires zero changes to the SplitParser pattern — just point it at `dankcal` instead of `qcal-wrapper.py`.
- Every command is independently testable: `dankcal list | jq .`
- Output must be a single line of JSON (no pretty-printing by default) to work with SplitParser's line-based reading.
- A `--pretty` flag can be added for human debugging without affecting QML.
