---
applyTo: ".github/workflows/**"
---

# GitHub Actions Workflows - Review Instructions

## Two-workflow release model

1. `auto-release.yml` - validates version, creates Git tag, creates GitHub release,
   then **dispatches** `release.yml`. Only runs from `main`.
2. `release.yml` - checks out an existing tag, runs GoReleaser, patches the
   prerelease flag. Can also be triggered manually for re-publish or recovery.

`release.yml` must NEVER create tags. `auto-release.yml` must NEVER run GoReleaser.

## Required secrets

- `RELEASE_TOKEN` - PAT or GitHub App token for creating protected tags. If absent,
  `auto-release.yml` fails explicitly with a clear message. Do not fall back to
  `GITHUB_TOKEN` for tag creation.
- `HOMEBREW_TAP_TOKEN` - used by GoReleaser to push to the Homebrew tap. If missing,
  the build succeeds but the tap silently does not update.
- `GITHUB_TOKEN` - automatic; used by GoReleaser for the release assets and by the
  prerelease patch step.

## Prerelease convention

Tags starting with `v0.` are pre-releases. The `Enforce prerelease flag` step in
`release.yml` patches the release after GoReleaser creates it, because GoReleaser
does not manage this flag from the tag name alone. Do not remove this step.

## Version format

`^v[0-9]+\.[0-9]+\.[0-9]+$` is the canonical format. Both `auto-release.yml` and
`release.yml` validate this. Any change must be applied to both files consistently.

## CI required checks (branch protection)

The job names below must match `.github/workflows/ci.yml` exactly:

- `Quality (vet + lint + tidy)`
- `Test (race)`
- `Security (govulncheck)`
- `Build (darwin binaries)`

Renaming a job without updating branch protection rules breaks mergeability.

## GoReleaser version

Pinned to `~> v2`. Do not upgrade to v3+ without reviewing `.goreleaser.yml` for
breaking changes and running `goreleaser release --snapshot --clean` first.

## Action version pinning

All actions are pinned to exact versions (e.g. `actions/checkout@v4.1.1`).
Do not use floating `@v4` or `@latest` tags - this keeps CI deterministic.