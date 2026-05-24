# ADR-010: Provider auth and Google Calendar discovery

**Status:** Accepted
**Date:** 2026-05-24
**Applies to:** `internal/auth/`, `internal/google/`, `internal/caldav/`, `internal/config/`

## Context

DankCalendar currently authenticates CalDAV requests with Basic auth using a username and app-specific password. This works for providers such as iCloud and many self-hosted CalDAV servers, but Google Workspace CalDAV rejects Basic auth and requires OAuth 2.0 bearer tokens.

Earlier Google-related work fixed CalDAV discovery behavior and event metadata rendering for Google Calendar, but did not add OAuth. Supporting Google Workspace requires two separate capabilities:

1. Authenticate HTTP requests with OAuth bearer tokens instead of Basic auth.
2. Discover Google calendar IDs for a signed-in account.

Google's current CalDAV endpoint is calendar-ID based. Pure CalDAV discovery is therefore not enough for a good account setup flow. The Google Calendar REST API exposes `users/me/calendarList`, which can discover calendar IDs, display names, and access roles. A related implementation in `dms-qcal-calendar` uses that endpoint for discovery, then implements event operations through the Google REST API.

DankCalendar already has a working CalDAV event engine, ICS parser/builder, recurrence handling, and JSON command surface. Replacing event operations with provider-specific REST API code would duplicate behavior and risk diverging output semantics.

## Decision

Introduce a small provider authentication boundary and keep event operations in the CalDAV core.

- Add a generic auth package, likely `internal/auth`, with an interface that can authorize HTTP requests.
- Keep Basic auth as one implementation for existing CalDAV providers.
- Add a Google-specific package, likely `internal/google`, responsible for OAuth and Google calendar discovery.
- Use Google Calendar REST API only for account/calendar discovery via `users/me/calendarList`.
- Convert discovered Google calendars into CalDAV calendar entries using Google's CalDAV URL shape:
  `https://apidata.googleusercontent.com/caldav/v2/<calendar_id>/events`
- Use OAuth bearer auth for those Google CalDAV entries, but continue to execute list/add/edit/delete through `internal/caldav`.

The Google OAuth flow must use desktop-safe OAuth mechanics:

- Browser-based installed-app flow.
- Loopback localhost callback.
- PKCE.
- A CSRF `state` value.
- Refresh-token storage in the system keyring.
- Access-token refresh before making authenticated requests.

The config file may record non-secret provider metadata such as auth type, provider, account, calendar ID, display name, and read-only status. It must not store refresh tokens, access tokens, client secrets, or passwords.

## Alternatives Considered

- **Keep Basic auth and document Google Workspace as unsupported**: Rejected because the issue is legitimate and Google Workspace is a common calendar source.
- **Use Google REST API for all Google event operations**: Rejected for the first implementation. It would duplicate the existing CalDAV event engine, add provider-specific add/edit/delete behavior, and risk inconsistent JSON output.
- **Use CalDAV-only discovery for Google**: Rejected because Google's supported endpoint model is calendar-ID based, making REST `calendarList` a better setup mechanism.
- **Store Google OAuth client credentials and tokens in config**: Rejected. The config file is for non-secret settings. Secrets belong in the system keyring.
- **Add third-party OAuth libraries immediately**: Rejected for now to preserve the stdlib-only decision unless the OAuth implementation becomes large or security-sensitive enough to justify revisiting ADR-002.

## Consequences

- Existing Basic-auth CalDAV providers keep the same behavior.
- Google Workspace support becomes a provider-specific setup/auth path, not a rewrite of the calendar backend.
- Calendar discovery for Google will require Calendar REST API scope in addition to CalDAV bearer-token use.
- Google accounts can expose read-only calendars; the UI and command layer must respect discovered access roles where practical.
- The config schema needs a backward-compatible extension for auth/provider metadata.
- Tests should cover Basic auth preservation, bearer authorization, token refresh behavior, and Google discovery response mapping without requiring live Google credentials.
