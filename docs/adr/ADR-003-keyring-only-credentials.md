# ADR-003: Keyring-only credentials via secret-tool

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** `internal/keyring/`

## Context

The existing Python wrapper supports both GNOME Keyring (via `secret-tool`) and plaintext passwords in `~/.config/qcal/config.json`. Plaintext credential storage is a security risk — config files can be accidentally committed, backed up unencrypted, or read by other processes.

## Decision

Credentials are stored and retrieved exclusively through `secret-tool` (freedesktop Secret Service API). There is no plaintext password field in the config file. The config file stores only the CalDAV URL, username, and non-sensitive display preferences.

## Alternatives Considered

- **Support both keyring and plaintext fallback**: Matches the old behavior but perpetuates the insecure path. Users who "just want it to work" will choose plaintext and stay there.
- **Encrypt passwords in the config file**: Adds complexity (key management, encryption scheme) for marginal benefit over the OS keyring.
- **Use libsecret directly via CGo**: Avoids the `secret-tool` subprocess but adds CGo build complexity and breaks cross-compilation.

## Consequences

- `secret-tool` must be installed — this is standard on GNOME/KDE desktops where DankMaterialShell runs.
- First-run setup requires `dankcalendar setup` which prompts for the password and stores it via `secret-tool store`.
- No passwords ever appear in config files, environment variables, or CLI arguments.
- Headless/containerized environments without a keyring daemon cannot use dankcalendar (acceptable — this is a desktop widget).
