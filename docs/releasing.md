# Releasing

**To release, merge the current release-please PR.** Merging ordinary changes
into `main` prepares that PR; it does not publish a release.

## Publish a release

1. Merge the changes you want to ship into `main` and wait for the
   **Release PR automation** workflow to update the release-please PR.
2. Review that PR's proposed version and changelog. Confirm it includes the
   intended changes and that its required checks pass, then merge it.
3. Follow the **Release** workflow in
   [GitHub Actions](https://github.com/jamestelfer/docker-credential-env/actions).
   It builds the archives and Linux packages, attests them, and publishes the
   [GitHub Release](https://github.com/jamestelfer/docker-credential-env/releases).

Release-please uses Conventional Commits to determine the next version and
release notes. If no release PR is open, check that there are unreleased
release-worthy commits and that **Release PR automation** succeeded.

## If publication fails

Inspect the failed workflow before taking action. A draft release is not a
completed release: publication happens only after the build and attestation
steps succeed. Do not manually publish the draft or create, move, or replace
its tag to bypass a failed run.

Once published, artifacts can be checked using the
[release verification instructions](../README.md#verifying-releases).

## Maintaining the pipeline

This repository calls the
[shared release pipeline](https://github.com/chinmina/.github/blob/main/docs/release-workflows.md)
using Octo STS and publishes only to GitHub Releases. Keep pipeline mechanics
and authentication guidance in the shared documentation rather than duplicating
them here.

Repository-specific settings live in `release-please-config.json` (versioning),
`.goreleaser.yaml` (builds and packaging), and `mise.toml` (tool versions).
Release-please maintains `.release-please-manifest.json` automatically.
