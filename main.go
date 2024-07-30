package main

import (
	"log/slog"
	"os"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/jamestelfer/docker-credential-env/helper"
)

// FIXME set at build time: goreleaser or Make
var Version string = "v0.0.0-unknown"
var Revision string = "unknown"

func main() {
	// configure the plugin details in the credentials package
	credentials.Name = "docker-credential-env"
	credentials.Package = "github.com/jamestelfer/docker-credential-env/credentials"
	credentials.Version = Version
	credentials.Revision = Revision

	configureLogging()

	credentialsOptional := os.Getenv("DOCKER_CREDENTIALS_ENV_OPTIONAL") == "true"

	credentials.Serve(helper.EnvHelper{CredentialsOptional: credentialsOptional})
}

func configureLogging() {
	logger := slog.
		New(slog.NewTextHandler(os.Stderr, nil)).
		With("pid", os.Getpid())

	slog.SetDefault(logger)
}
