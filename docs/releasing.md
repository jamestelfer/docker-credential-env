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

Open the failed **Release** run in GitHub Actions and inspect the first failed
step. Choose the recovery path based on whether the fix changes the tagged
source.

### Retry the same release

For transient service failures, or permissions and environment settings that
can be corrected without changing repository files:

1. Resolve the reported cause, such as restoring Octo STS access or waiting for
   a GitHub outage to clear.
2. Confirm the release is still a draft and no other attempt is running. If
   it is already published, verify its artifacts instead of retrying this
   path; publication may have succeeded despite the failed run.
3. If the draft contains assets from the failed attempt, remove those assets
   using the release editor, retaining the draft and its tag. The current
   GoReleaser configuration reuses the draft but does not replace existing
   assets, so leaving them can cause duplicate-upload errors.
4. On the failed Actions run, select **Re-run jobs → Re-run failed jobs**.
   This reruns the whole build, attestation, and publication job for the same
   tagged commit, rather than resuming at the failed step.

### Release a corrected version

If the fix changes source code, build configuration, or pinned tool versions,
merge it into `main`. Rerunning the old release still checks out the old tagged
commit and will not include that fix.

Have release-please prepare a new, unused version higher than the failed one,
then review and merge that release PR. If it proposes the failed version again,
set `release-as` in `release-please-config.json` to the replacement version
(without the `v` prefix) and merge that configuration change into `main`.
Remove the override after merging the replacement release PR so future versions
are calculated automatically. Leave the failed version unpublished and retain
its tag; the replacement release gets its own tag and provenance.

### Confirm recovery

Wait for **Release** to succeed, confirm the GitHub Release is published, and
[verify a downloaded artifact](../README.md#verifying-releases). Let the workflow
publish after attestation rather than manually publishing the draft.

## Maintaining the pipeline

This repository calls the
[shared release pipeline](https://github.com/chinmina/.github/blob/main/docs/release-workflows.md)
using Octo STS and publishes only to GitHub Releases. Keep pipeline mechanics
and authentication guidance in the shared documentation rather than duplicating
them here.

Repository-specific settings live in `release-please-config.json` (versioning),
`.goreleaser.yaml` (builds and packaging), and `mise.toml` (tool versions).
Release-please maintains `.release-please-manifest.json` automatically.
