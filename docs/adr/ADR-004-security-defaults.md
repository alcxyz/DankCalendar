# ADR-004: Security by default

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** entire project

## Context

CalDAV clients handle personal calendar data over the network and generate ICS content that is uploaded to remote servers. Several attack surfaces exist: URL manipulation, ICS injection, path traversal in filenames, and unencrypted transport.

## Decision

Apply defense-in-depth across all layers:

- **HTTPS-only**: Reject non-TLS CalDAV URLs. No `--insecure` flag.
- **ICS escaping**: All user-supplied text (summary, description, location) is escaped per RFC 5545 before embedding in ICS templates.
- **Filename validation**: Calendar and event filenames are validated against path traversal (`../`, null bytes).
- **URL construction**: Use `url.ResolveReference()` for all URL building — never string concatenation. This applies to `resolveHref` in `internal/caldav/discover.go` (fixed in v0.3.3; prior to v0.3.2 the function used string concatenation, producing malformed paths when `effectiveBase` contained a non-root path).
- **Config permissions**: Config files are created with `0600` permissions.
- **No shell expansion**: All subprocess calls use `exec.Command` with explicit argument lists, never shell interpolation.

## Alternatives Considered

- **Allow HTTP with a warning**: Some self-hosted CalDAV servers use plain HTTP behind a reverse proxy. Rejected because the connection between client and proxy is still unencrypted on the local network.
- **Trust user input in ICS fields**: Simpler code but opens injection vectors in shared calendars.

## Consequences

- Users with HTTP-only CalDAV servers must set up TLS (even self-signed with a local CA).
- Slightly more code in ICS generation for proper escaping.
- Config file permissions may conflict with backup tools that don't preserve modes — documented in README.
