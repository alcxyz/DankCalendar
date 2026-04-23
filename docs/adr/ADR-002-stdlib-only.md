# ADR-002: Stdlib-only Go, no external dependencies

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `go.mod`

## Context

Go's standard library provides `net/http`, `encoding/json`, `encoding/xml`, `net/url`, `os/exec`, `crypto/tls`, and `text/template` — everything needed for a CalDAV client that outputs JSON. Adding third-party dependencies increases supply-chain risk, complicates vendoring, and introduces version-management overhead for a focused CLI tool.

## Decision

Use only the Go standard library. No `go.sum` file, no third-party imports.

## Alternatives Considered

- **Use a CalDAV library (e.g. emersion/go-webdav)**: Provides higher-level abstractions but pulls in transitive dependencies and may not match our exact CalDAV subset (REPORT with calendar-data, PROPFIND for discovery).
- **Use a CLI framework (e.g. cobra, urfave/cli)**: Convenient for complex CLIs, but DankCalendar's command surface is small (~7 subcommands) and `os.Args` parsing suffices.

## Consequences

- Zero supply-chain risk — no dependency CVEs to track.
- `go build` works offline after initial module init.
- CalDAV XML request/response bodies are hand-crafted, which requires understanding the protocol but keeps the code explicit.
- If future requirements demand complex XML namespace handling, this decision may be revisited.
