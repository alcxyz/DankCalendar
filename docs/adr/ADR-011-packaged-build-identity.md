# ADR-011: Package one development identity for the plugin and helper

**Status:** Accepted
**Date:** 2026-09-19

## Decision

Keep the tracked manifest's X.Y.Z release version and shared release workflow.
Use the copied dms-plugins build-identity templates to package a stamped manifest
and helper with X.Y.Z-dev.<commit> (plus .dirty for modified source). Install the
plugin from the package's share/dms-plugins/DankCalendar directory, not raw source.
The package's `packaging.json` selects its directory and helper. Nix callers
supply revision metadata; source-only imports use a public-source fingerprint.
Manual packaging uses Python's standard library, then compiles the Go helper.
Official releases remain explicit and retain the stable version. Manual release
packaging requires the exact manifest release tag and a clean Git checkout.
Nix release output must be selected from the published release source.

## Alternatives and consequences

Stamping tracked manifests would interfere with release tags; helper-only
version changes would leave DMS reporting a different build. Copied canonical
packaging keeps standalone forks usable while the aggregate drift checker
prevents divergence. Direct go build reports embedded VCS identity but does not
install a matching DMS manifest; use packaging for complete installations.
