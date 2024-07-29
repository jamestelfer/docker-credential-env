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

## Configuration

> [!NOTE]
> The Docker CLI (`docker`) is the process that calls credential helpers, not
> the daemon. All environment variables (`PATH` and others) need to be set for
> the process calling Docker, and the executing user needs to be able to execute
> the helper binary.

### Installation

The binary needs to be added to the local `PATH` in order to be accessible to
Docker for use. The Docker CLI calls the helper (not the daemon), so the
executing user's `PATH` is used.

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
