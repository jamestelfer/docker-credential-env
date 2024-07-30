package helper

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
)

var ErrNotImplemented = errors.New("not implemented")

var nonAlphanumericPattern = regexp.MustCompile(`[^a-zA-Z0-9]`)

// must implement the Helper interface
var _ credentials.Helper = EnvHelper{}

type EnvHelper struct {
}

func (e EnvHelper) Get(serverURL string) (string, string, error) {
	logger := slog.Default().With("action", "get", "serverURL", serverURL)

	// FIXME: allow this to return empty on failure if so configured
	user, password, err := credentialsForServer(serverURL)

	logger.Info("Get", "user", user, "err", err)

	return user, password, err
}

func (e EnvHelper) Add(creds *credentials.Credentials) error {
	slog.Warn("Saving credentials is not supported by docker-credential-env", "action", "add", "serverURL", creds.ServerURL, "username", creds.Username)
	return nil
}

func (e EnvHelper) Delete(serverURL string) error {
	slog.Warn("Deleting credentials is not supported by docker-credential-env", "action", "delete", "serverURL", serverURL)
	return nil
}

func (e EnvHelper) List() (map[string]string, error) {
	slog.Info("", "action", "list")

	return nil, ErrNotImplemented
}

// credentialsForServer uses the normalized server URL to lookup a pair of environment variables.
func credentialsForServer(serverURL string) (string, string, error) {
	normalizedServerName := normalizeServerName(serverURL)

	// generate the names, used later in the error message if needed
	userEnv := envVarName(normalizedServerName, "USER")
	passwordEnv := envVarName(normalizedServerName, "PASSWORD")

	// user must have a value, password must be present but may be empty
	user, _ := os.LookupEnv(userEnv)
	password, passwordFound := os.LookupEnv(passwordEnv)

	// We require the password to be present even if it can be empty, so that the
	// user can tell between "this password is empty" and "the environment
	// variable name has a typo".
	if user == "" || !passwordFound {
		return "", "",
			fmt.Errorf("credentials for %s not found in environment variables %s and %s", serverURL, userEnv, passwordEnv)
	}

	return user, password, nil
}

// normalizeServerName converts the supplied name to a host name that can be
// used as an environment variable, taking the default Docker Hub special case
// into account.
func normalizeServerName(serverURL string) string {

	// Special case for the default index. This is passed as a URL, where no other
	// server is allowed to use this format.
	if strings.HasPrefix(serverURL, "https://index.docker.io/v1") {
		return "INDEX_DOCKER_IO"
	}

	return strings.ToUpper(
		strings.Trim(
			nonAlphanumericPattern.ReplaceAllString(serverURL, "_"),
			"_",
		),
	)
}

// envVarName looks up the environment variable with the given suffix using the normalized server name.
func envVarName(normalizedServerName string, suffix string) string {
	return fmt.Sprintf("DOCKER_CREDENTIALS_ENV_%s_%s", normalizedServerName, suffix)
}
