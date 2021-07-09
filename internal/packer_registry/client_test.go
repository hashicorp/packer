package packer_registry

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tc := []struct {
		name          string
		envs          []string
		errorExpected bool
		errCheckFunc  func(err error) bool
	}{
		{
			name:          "NonRegistryEnabledBuild",
			errorExpected: true,
			errCheckFunc:  checkRegistryEnabledError,
		},
		{
			name:          "HCP variables with no HCP_PACKER_REGISTRY variable",
			envs:          []string{"HCP_CLIENT_ID=foo", "HCP_CLIENT_SECRET=bar"},
			errorExpected: true,
			errCheckFunc:  checkRegistryEnabledError,
		},
		{
			name:          "Malformed HCP_PACKER_REGISTRY",
			envs:          []string{"HCP_PACKER_REGISTRY=home"},
			errorExpected: true,
			errCheckFunc: func(err error) bool {
				return checkErrorStatus(err, InvalidClientConfig)
			},
		},
		{
			name:          "HCP_PACKER_REGISTRY variable but no HCP credentials",
			envs:          []string{"HCP_PACKER_REGISTRY=myorgid/myprojectid"},
			errorExpected: true,
			errCheckFunc: func(err error) bool {
				return checkErrorStatus(err, InvalidClientConfig)
			},
		},
		{
			name:          "HCP_PACKER_REGISTRY variable but not all HCP credentials",
			envs:          []string{"HCP_PACKER_REGISTRY=myorgid/myprojectid", "HCP_CLIENT_ID=foo"},
			errorExpected: true,
			errCheckFunc: func(err error) bool {
				return checkErrorStatus(err, InvalidClientConfig)
			},
		},
		{
			name: "All required variables are set",
			envs: []string{"HCP_CLIENT_ID=foo", "HCP_CLIENT_SECRET=bar", "HCP_PACKER_REGISTRY=myorgid/myprojectid"},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			envKeys := loadEnvVariables(tt.envs)

			_, err := NewClient(ClientConfig{})
			if tt.errorExpected && err == nil {
				t.Errorf("creating a Client using env variables for its config should error when given only the following env. variables: %q", strings.Join(tt.envs, ","))
			}

			if err != nil {
				rawErr, ok := err.(*ClientError)
				if !ok {
					t.Errorf("expected a ClientError but got a %t instead", err)
				}

				if !tt.errCheckFunc(err) {
					t.Errorf("expected a different error for this client config: %v", rawErr)
				}
			}

			for _, k := range envKeys {
				os.Unsetenv(k)
			}
		})
	}
}

//ClientID     string
func TestNewClient_ClientConfig(t *testing.T) {
	tc := []struct {
		name          string
		cfg           ClientConfig
		errorExpected bool
		errCheckFunc  func(err error) bool
	}{
		{
			name:          "NonRegistryEnabledBuild",
			errorExpected: true,
			errCheckFunc:  checkRegistryEnabledError,
		},
		{
			name:          "HCP variables with no HCP_PACKER_REGISTRY variable",
			cfg:           ClientConfig{ClientID: "foo", ClientSecret: "bar"},
			errorExpected: true,
			errCheckFunc:  checkRegistryEnabledError,
		},
	}

	for _, tt := range tc {

		_, err := NewClient(tt.cfg)
		if tt.errorExpected && err == nil {
			t.Errorf("creating a Client using should error when given only the following client config %v", tt.cfg)
		}

		if err != nil {
			rawErr, ok := err.(*ClientError)
			if !ok {
				t.Errorf("expected a ClientError but got a %t instead", err)
			}

			if !tt.errCheckFunc(err) {
				t.Errorf("expected a different error for this client config: %v", rawErr)
			}
		}
	}
}

// loadEnvVariables sets all of the envs using os.Setenv defined. Expected format for envs is []string{"key=value",key2=value2"}
// The return is a slice of keys that should be unset by the caller once they envs are no longer needed.
func loadEnvVariables(envs []string) []string {
	envKeys := make([]string, 0, len(envs))
	for _, env := range envs {
		r := strings.Split(env, "=")
		os.Setenv(r[0], r[1])
		envKeys = append(envKeys, r[0])
	}
	return envKeys
}

func checkRegistryEnabledError(err error) bool {
	var clientError *ClientError
	if errors.As(err, &clientError) {
		return IsNonRegistryEnabledError(clientError)
	}

	return false
}

func checkErrorStatus(err error, statusCode uint) bool {
	var clientError *ClientError
	if errors.As(err, &clientError) {
		return clientError.StatusCode == statusCode
	}

	return false
}
