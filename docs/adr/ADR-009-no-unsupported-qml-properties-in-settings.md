# ADR-009: No unsupported QML properties in settings components

**Status:** Accepted
**Date:** 2026-04-28
**Applies to:** `CalendarSettings.qml`

## Context

The DMS plugin settings panel silently failed to render for DankCalendar. Clicking the expand arrow in Settings > Plugins showed nothing.

Root cause: `CalendarSettings.qml` set `password: true` on a `StringSetting` component. DMS's `StringSetting` (in `qs.Modules.Plugins`) does not declare a `password` property. In QML, setting an undeclared property on a loaded component causes a load-time error. Because DMS uses a `Loader` to dynamically instantiate the settings QML, the error is swallowed silently and the entire settings panel fails to render.

This was discovered by comparing DankCalculator (which has working settings) with DankCalendar and tracing through the DMS `PluginListItem.qml` Loader mechanism.

## Decision

Never use properties on DMS-provided setting components (`StringSetting`, `SliderSetting`, `ToggleSetting`, `SelectionSetting`) that are not declared in the upstream component definitions. If a feature is needed (e.g., password masking), request it upstream rather than assuming the property exists.

## Alternatives Considered

- **Wrap StringSetting with a custom component that adds password masking**: Rejected because DMS setting components use internal `findSettings()` parent traversal, and wrapping them would break that mechanism.
- **Fork the DMS StringSetting locally**: Rejected as too fragile; would break on DMS updates.

## Consequences

- The `caldavPassword` field currently renders as plaintext in the settings panel. This is acceptable because the settings UI is only visible to the local user, and the actual credential is stored in the system keyring.
- A feature request should be filed upstream on DMS to add `echoMode`/`password` support to `StringSetting`.
- Any future settings additions must be tested against the actual DMS component API before release.
