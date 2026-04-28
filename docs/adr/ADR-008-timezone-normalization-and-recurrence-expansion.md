# ADR-008: Timezone normalization and server-side recurrence expansion

**Status:** Accepted
**Date:** 2026-04-28
**Applies to:** `internal/ical/parser.go`, `internal/ical/recurrence.go`, `internal/caldav/list.go`

## Context

Events from different CalDAV calendars use different timezone representations: UTC (`Z` suffix), explicit `TZID=` parameters, or bare local times. The original parser stripped timezone info and stored raw strings, causing incorrect chronological ordering when sorting events across calendars.

Additionally, recurring events (birthdays, yearly reminders) were returned by the CalDAV server with their original DTSTART from years ago, since the query used `<c:calendar-query>` without requesting recurrence expansion.

## Decision

1. **Timezone normalization**: `parseDateTime` now parses the source timezone (UTC `Z`, `TZID=` param) and converts all timed events to the configured target timezone before producing the ISO string. This makes string-based sorting correct across calendars.

2. **Server-side expansion**: The CalDAV REPORT query now includes `<c:expand start="..." end="..."/>` inside `<c:calendar-data>`, requesting the server to expand recurring events into individual occurrences with correct dates (RFC 4791).

3. **Client-side RRULE fallback**: Some servers (particularly subscribed/public calendars) don't honour `<c:expand>`. For events returned with a start date before the query range, `AdjustRecurrence` parses the RRULE and computes the next occurrence within the range. Supports FREQ=YEARLY/MONTHLY/WEEKLY/DAILY with INTERVAL and UNTIL. Events with no RRULE and a stale date are dropped.

## Alternatives Considered

- **Full client-side RRULE engine**: Would handle all edge cases (BYDAY, BYSETPOS, EXDATE, COUNT) but adds significant complexity for a stdlib-only project. The hybrid approach covers the common cases.
- **Always trust server expansion**: Simpler, but breaks with subscribed calendars that don't support `<c:expand>`.
- **Sort by parsed `time.Time` instead of strings**: Would fix ordering without changing the output format, but doesn't solve the display problem (UTC times shown as-is to the user).

## Consequences

- All timed events display in the user's configured timezone regardless of source calendar.
- Recurring events show their actual next occurrence date, not the original creation date.
- The RRULE fallback handles common patterns (YEARLY, MONTHLY, WEEKLY, DAILY) but not exotic rules (BYDAY=2MO, BYSETPOS, EXDATE). These are handled correctly when the server supports `<c:expand>`.
- The config `timezone` field (already present but previously unused for reads) is now load-bearing.
