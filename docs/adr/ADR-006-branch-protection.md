# ADR-006: Protected main branch with dev workflow

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** GitHub repository settings, `.github/workflows/ci.yml`

## Context

The other DankMaterialShell plugins (DankQuickSearch, DankVault) use `main` + `dev` branches but without branch protection. Direct pushes to main risk breaking the release pipeline, which auto-tags and creates GitHub releases from the VERSION file on every main push.

## Decision

Protect the `main` branch with:
- Required status checks (CI "Test" job must pass, branch must be up to date)
- Required pull request (no direct pushes, 0 approvals needed since this is a solo project)
- No force pushes or deletions

All development happens on `dev`. Merges to `main` go through PRs, which triggers a release when VERSION is bumped.

## Alternatives Considered

- **No protection (match existing plugins)**: Simpler but allows accidental pushes that trigger unintended releases.
- **Full PR review requirements**: Overkill for a solo-maintained project.

## Consequences

- Every merge to main goes through CI before landing.
- Releases are deliberate — bump VERSION on dev, PR to main, CI passes, auto-release.
- Slightly more friction for small changes, but prevents broken releases.
