# ADR-007: Standalone DMS plugin replaces dms-qcal-calendar

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `plugin.json`, `CalendarWidget.qml`, `CalendarSettings.qml`

## Context

DankCalendar was built as a standalone CLI binary (ADR-001) with JSON-over-stdout IPC (ADR-005) specifically to replace the Python wrapper layer in `dms-qcal-calendar`. However, the QML plugin layer — the files that DankMaterialShell actually loads — still lived in the separate `dms-qcal-calendar` repository. Users had to install both repos: `dms-qcal-calendar` for QML + `DankCalendar` for the binary.

This meant `dms-qcal-calendar` was still required as a dependency even though its Python wrapper and qcal submodule were no longer used.

## Decision

Bundle the DMS plugin files (`plugin.json`, `CalendarWidget.qml`, `CalendarSettings.qml`) directly in the DankCalendar repository, making it a fully self-contained DMS plugin. The `dms-qcal-calendar` repository is no longer needed.

The QML layer calls `dankcalendar` subcommands directly (no Python, no wrapper scripts). Credential sync uses `dankcalendar discover --url --username --password` which stores the password in the system keyring via `secret-tool` (ADR-003).

## Alternatives Considered

**Keep QML in dms-qcal-calendar, just swap the binary:** Would avoid code duplication but perpetuates the two-repo install requirement and keeps an otherwise empty repo alive.

**Stdin-based credential passing instead of CLI args for discover:** More secure (avoids password in `ps` output), but would require changes to the Go CLI and QML Process interaction. The discover command is a one-time setup operation, and the password exposure window is brief. Accepted as a known trade-off; can be revisited in a future ADR if needed.

## Consequences

- **DankCalendar is now a complete, installable DMS plugin** — copy the directory to `~/.config/DankMaterialShell/plugins/DankCalendar/` with `dankcalendar` in PATH.
- **dms-qcal-calendar can be archived** — it is fully superseded.
- **No Python runtime dependency** — `requires` in plugin.json lists only `dankcalendar`, `secret-tool`, `notify-send`.
- **Password briefly visible in process listing** during initial credential setup via `dankcalendar discover`. Once stored in keyring, subsequent commands never handle the password.
