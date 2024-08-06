package helper

import (
	"os"
	"testing"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/stretchr/testify/assert"
)

type entry struct {
	url          string
	user         string
	password     string
	omitPassword bool
}

func (e entry) Set(t *testing.T) {
	normalized := normalizeServerName(e.url)
	t.Setenv(envVarName(normalized, userSuffix), e.user)
	if !e.omitPassword {
		t.Setenv(envVarName(normalized, passwordSuffix), e.password)
	}
}

func TestEnvHelper_Get_Success(t *testing.T) {
	setTestEnv(t, "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USER", "testuser")
	setTestEnv(t, "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD", "testpassword")

	helper := EnvHelper{}

	user, password, err := helper.Get("example.com")

	assert.NoError(t, err)
	assert.Equal(t, "testuser", user)
	assert.Equal(t, "testpassword", password)
}

func TestEnvHelper_Get_CredentialsOptional_Success(t *testing.T) {

	helper := EnvHelper{CredentialsOptional: true}

	user, password, err := helper.Get("example.com")

	assert.NoError(t, err)
	assert.Empty(t, user)
	assert.Empty(t, password)
}

func TestEnvHelper_Get_Failure(t *testing.T) {
	helper := EnvHelper{}

	user, password, err := helper.Get("nonexistent.com")

	assert.Error(t, err)
	assert.Empty(t, user)
	assert.Empty(t, password)
}

func TestEnvHelper_Add_FailsSilently(t *testing.T) {
	helper := EnvHelper{}
	err := helper.Add(&credentials.Credentials{})
	assert.NoError(t, err)
}

func TestEnvHelper_Delete_FailsSilently(t *testing.T) {
	helper := EnvHelper{}
	err := helper.Delete("foo")
	assert.NoError(t, err)
}

func TestEnvHelper_List(t *testing.T) {
	// empty with none
	// users where supplied
	// skip when password is missing

	tests := []struct {
		name     string
		entries  []entry
		env      map[string]string
		expected map[string]string
	}{
		{
			name: "normal",
			entries: []entry{
				{url: "example.com", user: "testuser", password: "testpassword"},
				{url: defaultRegistryUrl, user: "hubuser", password: "hubpassword"},
			},
			expected: map[string]string{
				"example.com":      "testuser",
				defaultRegistryUrl: "hubuser",
			},
		},
		{
			name: "default registry",
			entries: []entry{
				{url: defaultRegistryUrl, user: "hubuser", password: "hubpassword"},
			},
			expected: map[string]string{
				defaultRegistryUrl: "hubuser",
			},
		},
		{
			name: "missing password",
			entries: []entry{
				{url: "nopassword.test", user: "missing", omitPassword: true},
			},
			expected: map[string]string{},
		},
		{
			name: "empty user",
			entries: []entry{
				{url: "example.com", user: "", password: "testpassword"},
			},
			expected: map[string]string{},
		},
		{
			name:    "malformed env",
			entries: []entry{},
			env: map[string]string{
				// these env entries are malformed and should be ignored
				"DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USR":       "testuser",
				"DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD":  "testpassword",
				"DOCKER_CREDENTIALS_ENV_INDEX_DOCKER_IO_USER":  "hubuser",
				"DOCKER_CREDENTIALS_ENV_INDEX_DOCKER_IO_PASS":  "hubpassword",
				"DOCKER_CREDENTIALS_ENV_EXAMPLE2_COM_USERS":    "testuser",
				"DOCKER_CREDENTIALS_ENV_EXAMPLE2_COM_PASSWORD": "testpassword",
			},
			expected: map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, e := range test.entries {
				e.Set(t)
			}
			if test.env != nil {
				for k, v := range test.env {
					t.Setenv(k, v)
				}
			}

			helper := EnvHelper{}
			r, err := helper.List()

			assert.NoError(t, err)
			assert.Equal(t, test.expected, r)
		})
	}

}

func TestCredentialsForServerSuccess(t *testing.T) {
	tests := []struct {
		serverURL string
		userEnv   string
		passEnv   string
		user      string
		password  string
	}{
		{
			serverURL: "example.com",
			userEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USER",
			passEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD",
			user:      "testuser",
			password:  "testpassword",
		},
		{
			serverURL: "example.com",
			userEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USER",
			passEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD",
			user:      "testuser",
			password:  "",
		},
		{
			serverURL: "https://index.docker.io/v1/",
			userEnv:   "DOCKER_CREDENTIALS_ENV_INDEX_DOCKER_IO_USER",
			passEnv:   "DOCKER_CREDENTIALS_ENV_INDEX_DOCKER_IO_PASSWORD",
			user:      "anotheruser",
			password:  "anotherpassword",
		},
	}

	for _, test := range tests {
		t.Run(test.serverURL, func(t *testing.T) {
			// Set environment variables with cleanup
			setTestEnv(t, test.userEnv, test.user)
			setTestEnv(t, test.passEnv, test.password)

			user, password, err := credentialsForServer(test.serverURL)
			assert.NoError(t, err)
			assert.Equal(t, test.user, user)
			assert.Equal(t, test.password, password)
		})
	}
}

func TestCredentialsForServerFailure(t *testing.T) {
	tests := []struct {
		serverURL string
		userEnv   string
		user      *string
		passEnv   string
	}{
		{
			serverURL: "https://example.com",
			userEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USER",
			passEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD",
		},
		{
			serverURL: "example.com",
			userEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USER",
			passEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD",
		},
		{
			serverURL: "example.com",
			userEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_USER",
			user:      ptr("testuser"),
			passEnv:   "DOCKER_CREDENTIALS_ENV_EXAMPLE_COM_PASSWORD",
		},
	}

	for _, test := range tests {
		t.Run(test.serverURL, func(t *testing.T) {
			setOptTestEnv(t, test.userEnv, test.user)
			setOptTestEnv(t, test.passEnv, nil)

			user, password, err := credentialsForServer(test.serverURL)
			assert.Error(t, err)
			assert.Empty(t, user)
			assert.Empty(t, password)
		})
	}
}

func TestNormalizeServerName(t *testing.T) {
	tests := []struct {
		serverURL string
		expected  string
	}{
		{
			serverURL: "https://index.docker.io/v1/",
			expected:  "INDEX_DOCKER_IO",
		},
		{
			serverURL: "https://index.docker.io/v1",
			expected:  "INDEX_DOCKER_IO",
		},
		{
			serverURL: "example.com",
			expected:  "EXAMPLE_COM",
		},
		{
			serverURL: "example.com:8080",
			expected:  "EXAMPLE_COM_8080",
		},
		{
			serverURL: "example.com/path",
			expected:  "EXAMPLE_COM_PATH",
		},
	}

	for _, test := range tests {
		t.Run(test.serverURL, func(t *testing.T) {
			result := normalizeServerName(test.serverURL)
			assert.Equal(t, test.expected, result, "normalizeServerName(%s)", test.serverURL)
		})
	}
}

// setTestEnv sets an environment variable to a new value and registers a cleanup function
// that restores the original value (unsetting it if the original value was unset).
func setTestEnv(t *testing.T, key, value string) {
	setOptTestEnv(t, key, &value)
}

func setOptTestEnv(t *testing.T, key string, value *string) {
	t.Helper()

	originalValue, wasSet := os.LookupEnv(key)
	if value != nil {
		os.Setenv(key, *value)
	} else {
		os.Unsetenv(key)
	}

	t.Cleanup(func() {
		if wasSet {
			os.Setenv(key, originalValue)
		} else {
			os.Unsetenv(key)
		}
	})
}

func ptr(s string) *string {
	return &s
}
