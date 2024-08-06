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

const (
	variablePrefix = "DOCKER_CREDENTIALS_ENV"
	userSuffix     = "USER"
	passwordSuffix = "PASSWORD"

	defaultRegistryUrl = "https://index.docker.io/v1"
)

// must implement the Helper interface
var _ credentials.Helper = EnvHelper{}

type EnvHelper struct {
	CredentialsOptional bool
}

// Get retrieves the credentials for the server URL from the process
// environment.
func (e EnvHelper) Get(serverURL string) (string, string, error) {
	logger := slog.Default().With("action", "get", "serverURL", serverURL)

	user, password, err := credentialsForServer(serverURL)

	if err != nil {
		if e.CredentialsOptional {
			logger.Warn("Ignoring failed credential lookup, unset DOCKER_CREDENTIALS_ENV_OPTIONAL if this should cause Docker to fail", "err", err)
			return "", "", nil
		}

		logger.Error("Failed to retrieve credentials, set DOCKER_CREDENTIALS_ENV_OPTIONAL to ignore this error.", "err", err)
		return "", "", err
	}

	logger.Info("Got credentials from environment", "user", user, "err", err)

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
	return listCredentialsForEnvironment(), nil
}

// credentialsForServer uses the normalized server URL to lookup a pair of environment variables.
func credentialsForServer(serverURL string) (string, string, error) {
	normalizedServerName := normalizeServerName(serverURL)

	// generate the names, used later in the error message if needed
	userEnv := envVarName(normalizedServerName, userSuffix)
	passwordEnv := envVarName(normalizedServerName, passwordSuffix)

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
	if strings.HasPrefix(serverURL, defaultRegistryUrl) {
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
	return fmt.Sprintf("%s_%s_%s", variablePrefix, normalizedServerName, suffix)
}

// listCredentialsForEnvironment returns a map of server URLs to their
// credentials based on the current environment. It will only return credentials
// for servers that have correctly formed environment variables for user and
// password.
func listCredentialsForEnvironment() map[string]string {
	// Get all environment variables that start with the prefix
	prefix := fmt.Sprintf("%s_", variablePrefix)
	suffix := fmt.Sprintf("_%s", userSuffix)

	servers := map[string]string{}

	vars := os.Environ()
	for _, v := range vars {
		key, val, found := strings.Cut(v, "=")
		if !found {
			continue
		}

		serverURL := key

		serverURL, found = strings.CutPrefix(serverURL, prefix)
		if !found {
			continue
		}

		serverURL, found = strings.CutSuffix(serverURL, suffix)
		if !found {
			continue
		}

		if _, _, err := credentialsForServer(serverURL); err != nil {
			continue
		}

		serverURL = toHostname(serverURL)

		// Special case. Docker always uses the full URL for the default registry,
		// even when supplied to the CLI in short form.
		if serverURL == "index.docker.io" {
			serverURL = defaultRegistryUrl
		}

		servers[serverURL] = val
	}

	return servers
}

func toHostname(serverURL string) string {
	return strings.ToLower(strings.ReplaceAll(serverURL, "_", "."))
}
