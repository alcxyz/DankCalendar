# Changelog

All notable changes to this plugin are documented here. The format follows
Keep a Changelog, and the release workflow publishes each version's section
as its GitHub release notes.

## [Unreleased]

## [0.7.1] - 2026-09-20

- Reorganized the README to focus on what the plugin is, requirements, install, and account setup; detailed Google OAuth setup and the full CLI command table moved into dedicated `docs/google-oauth.md` and `docs/cli.md` guides, with the design rationale folded into `CONTRIBUTING.md`, so the top-level docs are easier to scan.
- Development builds now stamp an identifiable `X.Y.Z-dev.<commit>` version (`.dirty` for uncommitted changes) via `scripts/package.py` or the Nix `default.nix`, so a checkout install can be told apart from a tagged release. Release packaging still requires a clean checkout at the exact `vX.Y.Z` tag.

## [0.7.0] - 2026-07-03

### Security hardening

- CalDAV discovery passwords are now fed over stdin instead of a command-line argument, so they no longer show up in process listings.
- The plaintext CalDAV password entered in plugin settings is cleared automatically once it has been synced into the system keyring.
- The CalDAV password field in settings is now marked as a password input.
- CalDAV redirects that would downgrade from HTTPS or change host are rejected, preventing credentials from being replayed to an unsafe destination.
- Only HTTPS conference links are exposed as clickable Meet links.

### Documentation

- Clarified that Google OAuth setup uses the user's own Google Cloud OAuth client, not a shared one.

## [0.6.0] - 2026-07-02

### Event details

- Added a plugin setting to show event details/descriptions in the calendar UI.
- Event descriptions can now be added or edited directly when creating or editing events.
- The CLI supports descriptions via `add --description` and `edit --description`.

### Calendar import fixes

- Fixed parsing of top-level `VEVENT` descriptions so nested alarm descriptions are no longer picked up by mistake.
- HTML-ish event descriptions from external calendars are now normalized into readable plain text.
- The QML component cache for the plugin's entry points is busted on load, so UI updates are actually picked up after an upgrade instead of showing stale widgets.

## [0.5.0] - 2026-06-02

- Added Google Workspace support: a new `google-discover` command runs an OAuth 2.0 browser authorization flow (using the user's own Google Cloud OAuth client), stores the refresh token in the system keyring, and writes discovered Google calendars into the normal config alongside CalDAV accounts. This is required because Google's current CalDAV endpoint rejects basic auth and app-specific passwords.
- Added a "hide past events" toggle so timed events that have already started can be hidden from the list automatically as time passes.
- Fixed the Nix install example in the README to use the correct plugin id.

## [0.4.1] - 2026-04-28

- Added a scrollable event list in the popout panel, so long lists of events are no longer clipped.
- Added clickable "Join Meet" links for Google Calendar events that include a conference link, plus new "show Meet link" and "show RSVP count" settings.
- Added an accepted/total attendee (RSVP) count on events that have attendees.
- Fixed a bug where the password field in settings used an unsupported `password` property, which caused a QML load error that silently prevented the entire settings panel from rendering.
- Added a screenshot, `CONTRIBUTING.md`, and Nix install instructions to the docs.

## [0.4.0] - 2026-04-23

### Bug fixes

- Fixed cross-calendar event sorting: events from different CalDAV calendars (UTC, TZID, or bare local time) are now normalized to the configured timezone before sorting, producing correct chronological order. Previously a 14:00 UTC event could sort before a 10:00 Lisbon event despite actually being later.
- Fixed recurring events showing their original creation date instead of the next occurrence. The CalDAV query now requests server-side recurrence expansion, with a client-side fallback (yearly/monthly/weekly/daily) for subscribed calendars that don't support it.

### Enhancements

- The calendar name is now shown by default in both the event list and the bar pill summary, so it's clear at a glance which calendar an event belongs to.
- The default look-ahead window was increased from 7 to 14 days for better visibility of upcoming events; it remains configurable from 1 to 30 days.

### Documentation

- Added an ADR documenting the timezone normalization and server-side recurrence expansion approach.
- Updated the README to describe timezone-aware sorting and recurring event support.

## [0.3.3] - 2026-04-23

- Fixed a URL construction bug in CalDAV discovery where relative hrefs were resolved by string concatenation instead of proper URL reference resolution, producing malformed paths when the discovered base URL contained a non-root path.

## [0.3.2] - 2026-04-23

- Fixed CalDAV discovery to look up `.well-known/caldav` at the server root per RFC 5785 instead of appending it to the full base path.
- Fixed the effective base URL used during discovery so it no longer retains a full path, which previously caused doubled paths when resolving absolute hrefs.
- Discovery now falls back to treating the provided URL as the principal when the server does not report a `current-user-principal` (as some servers do), instead of failing outright.
- Removed references to the plugin's prior, non-public predecessor from the docs and ADRs.

## [0.3.1] - 2026-04-23

- Converted the project into a fully self-contained DankMaterialShell widget plugin: the QML widget now calls the `dankcalendar` binary directly for all CalDAV operations, with no external scripts or dependencies.
- Added a non-interactive `discover` command for QML integration, and Nix flake packaging that produces a `dankcalendar` binary build.
- Added an `--append` flag to `discover` so setting up a second CalDAV account (for example Google Calendar alongside iCloud) no longer overwrites previously configured accounts.
- Renamed the project to DankCalendar.

## [0.2.0] - 2026-04-23

- Implemented the core CalDAV client: calendar discovery, event listing, and create/edit/delete over CalDAV, enforcing HTTPS.
- Added an iCalendar (ICS) parser and builder with RFC 5545-compliant escaping.
- Credentials are stored exclusively in the system keyring via `secret-tool`; no plaintext passwords are written to disk.
- Added desktop notifications for upcoming events, with deduplication so the same event doesn't notify twice.
- Wired up the full CLI: `list`, `calendars`, `add`, `edit`, `delete`, `notify`, and `setup`.

## [0.1.0] - 2026-04-23

- Initial release: project scaffold with the Go module, CI, and test harness in place, laying the groundwork for the CalDAV calendar plugin.
