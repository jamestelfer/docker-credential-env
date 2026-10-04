# Releasing

This repository uses the
[shared chinmina release pipeline](https://github.com/chinmina/.github/blob/main/docs/adopting-the-release-pipeline.md)
with Octo STS authentication and GitHub Releases as its only publishing channel.
Binstaller, Homebrew, Docker login, and npm publishing are disabled. No release
secrets are required.

## Release flow

1. Conventional commits merged to `main` cause release-please to open or update
   a release PR.
2. Merging that PR creates a draft GitHub Release and pushes a plain `v*` tag
   using an Octo STS installation token.
3. `.github/workflows/release.yml` calls the shared build workflow. GoReleaser
   fills the existing draft with archives and Linux packages, the workflow
   attests those artifacts, and only then publishes the release.

Do not manually publish the draft or push a tag to start a normal release;
release-please owns that transition. The former `release.yaml` workflow has
been replaced, so there is only one tag-triggered publisher.

The migration manifest starts at the existing release `1.0.0`, not `0.0.0`.
`bootstrap-sha` identifies the `v1.0.0` commit so the initial release PR only
considers subsequent changes. Release-please maintains the manifest thereafter.
Go and GoReleaser release tool versions are pinned in `mise.toml`; `go.mod`
continues to describe the module's minimum Go version.

## Rollout gates

Complete these before merging the consumer workflows:

- Protect `jamestelfer/.github`'s `main` branch with required review and no Octo
  STS bypass. It had no branch protection or rulesets at onboarding; changing
  this shared security boundary requires owner confirmation.
- Confirm the Octo STS App is installed and can access both this repository and
  the central policy repository. The owner is a user account, so the guide's
  org installation-enumeration endpoint cannot establish this.

The following prerequisites were verified before opening the consumer PR:

- The [central Octo STS policy PR](https://github.com/jamestelfer/.github/pull/4)
  is merged. The identities are `release-please-docker-credential-env`
  (environment `automation`, ref `refs/heads/main`, contents and PR writes) and
  `release-docker-credential-env` (environment `release`, `v*` tag refs,
  contents writes). Both grant access only to this repository. Policies belong
  in `jamestelfer/.github`, never in this consumer repository.
- The `standard: tag creation` ruleset allows the Octo STS App (`801323`) as
  an `Integration` bypass actor with `bypass_mode: always`. The separate
  `standard: tag immutability` ruleset has no bypasses and continues to prevent
  tag updates and deletions.

- `automation` environment: custom deployment policy allowing only branch
  `main`.
- `release` environment: custom deployment policy allowing only tags `v*`.
- Actions policy permits all actions, including `chinmina/*`.

The old `Publish` environment is no longer referenced by these workflows.
Recheck environment gates before rollout; GitHub otherwise auto-creates missing
environments without protections.

After the prerequisites are ready, smoke-test the Octo STS mint from an
`automation` job on `main` before a real release. Then confirm a release PR
creates a draft, its App-pushed tag triggers the build, and publication happens
only after attestation. See [Verifying releases](../README.md#verifying-releases)
for verification against the shared reusable signer. A local snapshot does not
prove these GitHub/OIDC steps.

## Local checks

```sh
mise install
mise exec actionlint@1.7.12 -- actionlint
mise exec zizmor@1.25.2 -- zizmor .
mise exec -- go test ./... -race
mise exec -- goreleaser check
mise exec -- goreleaser release --snapshot --clean --skip=publish
```

The snapshot builds all six OS/architecture combinations and the Linux packages
without creating a tag, publishing assets, or minting credentials.
