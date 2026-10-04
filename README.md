# docker-credential-env plugin

A Docker credentials plugin that sources Docker credentials from environment
variables. This is an alternative to using `docker login` directly.

Some CI providers use environment variables to communicate the configuration of
Docker Hub. In these cases, instead of performing a `docker login` when
bootstrapping the agent for use, one can instead build this helper into the
agent image.

This has some benefits:

- Reliability: the login will fail only when running an action on an image
  (pull, push etc). If the step has no docker actions, it will not be affected.
  Many CI systems now allow agents to login to Docker Hub all the time,
  regardless of the actions that will be performed. Using this plugin means that
  only actions that attempt to use the failing credentials will cause a failure.
- Performance: since login is only attempted when the credentials are required,
  agents avoid an unnecessary setup step. This may seem insignificant, but it
  can add up in aggregate.

When environment variables are set appropriately, `docker` will use the
credentials as needed.

## Installation

Pre-built binaries for Linux, macOS, and Windows (amd64/arm64), plus Linux
packages, are published to
[GitHub Releases](https://github.com/jamestelfer/docker-credential-env/releases).
Releases produced by the shared release pipeline carry build-provenance
attestations — see [Verifying releases](#verifying-releases). Older releases
(v1.0.0 and earlier) do not have these attestations.

<details>
<summary><strong>mise (recommended)</strong></summary>

[mise](https://mise.jdx.dev/) installs directly from GitHub Releases via its
[GitHub backend](https://mise.jdx.dev/dev-tools/backends/github.html). With
`github_attestations` enabled, it verifies build provenance as well as the
artifact checksum:

```sh
mise use -g github:jamestelfer/docker-credential-env
```

See [Verifying releases](#verifying-releases) for the provenance guarantees.

</details>

<details>
<summary><strong>Manual download</strong></summary>

Download the archive for your platform from the
[releases page](https://github.com/jamestelfer/docker-credential-env/releases),
verify its provenance, and put the binary on your `PATH`. For Linux or macOS:

```sh
OS=linux ARCH=amd64 # or darwin, arm64
curl -fsSLO "https://github.com/jamestelfer/docker-credential-env/releases/latest/download/docker-credential-env_${OS}_${ARCH}.tar.gz"
gh attestation verify "docker-credential-env_${OS}_${ARCH}.tar.gz" \
  --repo jamestelfer/docker-credential-env \
  --signer-workflow chinmina/.github/.github/workflows/goreleaser-release.yml
tar -xzf "docker-credential-env_${OS}_${ARCH}.tar.gz" docker-credential-env
mkdir -p ~/.local/bin
install -m 0755 docker-credential-env ~/.local/bin/
```

Ensure `~/.local/bin` is on Docker's `PATH`. Windows archives are also `.tar.gz`;
extract `docker-credential-env.exe` and put it on your `PATH`.
See [Verifying releases](#verifying-releases) before installing.

</details>

<details>
<summary><strong>go install</strong></summary>

Build from source:

```sh
go install github.com/jamestelfer/docker-credential-env@latest
```

Ensure Go's binary installation directory is on Docker's `PATH`. Source builds
are not covered by release artifact attestations and report the unstamped
version `v0.0.0-unknown`.

</details>

## Verifying releases

The shared release workflow generates SLSA build-provenance attestations for
archives and Linux packages using [Sigstore](https://www.sigstore.dev/)
keyless signing, before publishing the GitHub Release. No long-lived signing
key is used. Each attested artifact is bound by digest to its source commit
and build workflow.

With an authenticated [GitHub CLI](https://cli.github.com/) (2.49.0 or newer),
verify a downloaded artifact:

```sh
gh attestation verify docker-credential-env_linux_amd64.tar.gz \
  --repo jamestelfer/docker-credential-env \
  --signer-workflow chinmina/.github/.github/workflows/goreleaser-release.yml
```

The signer is the shared reusable workflow, while the source repository is
`jamestelfer/docker-credential-env`.
Releases at v1.0.0 and earlier predate this pipeline and have no attestations.
`checksums.txt` is also published for checksum-only verification with
`sha256sum --check checksums.txt`; checksums alone do not prove build provenance.

For maintainers, see [Releasing](docs/releasing.md) for how to publish a release.

## Configuration

> [!NOTE]
> The Docker CLI (`docker`) is the process that calls credential helpers, not
> the daemon. All environment variables (`PATH` and others) need to be set for
> the process calling Docker, and the executing user needs to be able to execute
> the helper binary.

### Registry configuration

The plugin uses environment variables in a particular format to supply
credentials to the Docker process. These environment variables need to be
present in the process executing the `docker` CLI command.

All environment variables have the form:
`DOCKER_CREDENTIALS_ENV_<REGISTRYURL>_<USER|PASSWORD>`, where:

- `REGISTRY_URL` identifies the URL that the credentials are for. This is the
  registry URL (really just the host) as configured in `config.json`, with no
  leading scheme and no trailing slash. For the ECR public registry, the
  registry host is `public.ecr.aws`, so the environment variable pair will
  include `PUBLIC_ECR_AWS` in the name. The special value `DEFAULT` can be used
  for the Docker Hub registry.
- `<USER|PASSWORD>`: credentials are supplied as a pair of variables: the `USER`
  and the `PASSWORD`.

### Fail silently

If your environment should fail quietly if the authentication variables are not
present, set the `DOCKER_CREDENTIALS_ENV_OPTIONAL` variable to `true`. When this
variable is set, the plugin will write an error indicating that the credentials
weren't found, but will not return an error to Docker.

### Examples

#### Docker Hub

Environment:

```shell
export DOCKER_CREDENTIALS_ENV_INDEX_DOCKER_IO_USER=dockerhubusername
export DOCKER_CREDENTIALS_ENV_INDEX_DOCKER_IO_PASSWORD=userapikey
```

`~/.docker/config.json` fragment:

```json
{
 "credHelpers": {
    "https://index.docker.io/v1/": "env"
  },
}
```

> [!IMPORTANT]
> Unlike other registries, the default registry must be specified in
> `config.json` as a full URL. This quirk is **only** relevant to the Docker Hub
> registry.

#### ECR public registry

Environment:

```shell
export DOCKER_CREDENTIALS_ENV_PUBLIC_ECR_AWS_USER=AWS
export DOCKER_CREDENTIALS_ENV_PUBLIC_ECR_AWS_PASSWORD=password-from-aws-cli
```

`~/.docker/config.json` fragment:

```json
{
 "credHelpers": {
    "public.ecr.aws": "env"
  },
}
```
