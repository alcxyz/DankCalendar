# ADR-006: Protected main branch with dev workflow

**Status:** Accepted
**Date:** 2026-04-23
**Applies to:** GitHub repository settings, `.github/workflows/ci.yml`

## Context

The DankMaterialShell plugin repositories use `main` for releases and `dev` for active development. Direct pushes to `main` risk bypassing the release review path and can trigger unintended release automation.

## Decision

Protect the `main` branch with:
- Pull requests for changes to `main`
- No required approvals by default, since these are solo-maintained repositories
- No force pushes or deletions

All development happens on `dev`. Releases go through a PR from a release branch into `main`. Version bumps happen on the release branch unless the release is documentation-only. Release automation reads `plugin.json.version`.

## Alternatives Considered

- **No protection (match existing plugins)**: Simpler but allows accidental pushes that trigger unintended releases.
- **Full PR review requirements**: Overkill for a solo-maintained project.

## Consequences

- Every merge to main goes through a PR before landing.
- Releases are deliberate: develop on `dev`, create a release branch, bump `plugin.json.version` for non-doc releases, then PR to `main`.
- Slightly more friction for small changes, but prevents broken releases.
