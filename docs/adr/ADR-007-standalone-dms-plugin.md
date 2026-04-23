# ADR-007: DankCalendar bundles its own DMS plugin layer

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `plugin.json`, `CalendarWidget.qml`, `CalendarSettings.qml`

## Context

DankCalendar is both a CalDAV CLI tool and a DankMaterialShell widget plugin. The QML files that DMS loads (`plugin.json`, `CalendarWidget.qml`, `CalendarSettings.qml`) could live in a separate repository, requiring users to install and update two repos to get one working calendar widget.

## Decision

Bundle the DMS plugin files directly in this repository alongside the Go source. A single install gives users the binary and the UI. The plugin source is distributed through `dms-plugins` (like all other DankMaterialShell plugins), while the Go binary is provided as a Nix package from the same flake.

The QML layer calls `dankcalendar` subcommands directly — no interpreter, no wrapper scripts.

## Alternatives Considered

**Separate repository for QML files:** Decouples UI iterations from backend releases, but means two repos to update and version-sync for what is a single logical thing.

**Stdin-based credential passing for `discover`:** More secure (avoids password briefly in `ps` output), but requires a more complex QML `Process` setup. The discover command runs once at setup time; the exposure window is brief. Accepted as a known trade-off (see ADR-003).

## Consequences

- One repository, one release, one install.
- Plugin version tracks the binary version via the `VERSION` file.
- CI validates both the Go binary and the plugin manifest in the same test run.
- Credential setup requires `dankcalendar discover`, which briefly passes the password via CLI args before storing it in the keyring.
