package packer_registry

import (
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tt := []struct {
		name            string
		envs            []string
		errorExpected   bool
		errorStatusCode uint
	}{
		{
			name:            "NonRegistryEnabledBuild",
			errorExpected:   true,
			errorStatusCode: NonRegistryEnabled,
		},
		{
			name:            "HCP variables with no PACKER_ARTIFACT_REGISTRY variable",
			envs:            []string{"HCP_CLIENT_ID=foo", "HCP_CLIENT_SECRET=bar"},
			errorExpected:   true,
			errorStatusCode: NonRegistryEnabled,
		},
		{
			name:          "Malformed PACKER_ARTIFACT_REGISTRY",
			envs:          []string{"PACKER_ARTIFACT_REGISTRY=home"},
			errorExpected: true,
		},
		{
			name:            "PACKER_ARTIFACT_REGISTRY variable but no HCP credentials",
			envs:            []string{"PACKER_ARTIFACT_REGISTRY=myorgid/myprojectid"},
			errorExpected:   true,
			errorStatusCode: InvalidHCPConfig,
		},
		{
			name:            "PACKER_ARTIFACT_REGISTRY variable but not all HCP credentials",
			envs:            []string{"PACKER_ARTIFACT_REGISTRY=myorgid/myprojectid", "HCP_CLIENT_ID=foo"},
			errorExpected:   true,
			errorStatusCode: InvalidHCPConfig,
		},
		{
			name: "All required variables are set",
			envs: []string{"HCP_CLIENT_ID=foo", "HCP_CLIENT_SECRET=bar", "PACKER_ARTIFACT_REGISTRY=myorgid/myprojectid"},
		},
	}

	for _, tc := range tt {

		envKeys := loadEnvVariables(tc.envs)

		_, err := NewClient(ClientConfig{})
		if tc.errorExpected && err == nil {
			t.Errorf("creating a Client using env variables for its config should error when given only the following env. variables: %q", strings.Join(tc.envs, ","))
		}

		if err != nil {
			rawErr, ok := err.(*ClientError)
			if !ok {
				t.Errorf("expected a ClientError but got a %t instead", err)
			}

			if rawErr.StatusCode != tc.errorStatusCode {
				t.Errorf("expected a different error for this client config: %v", rawErr)
			}
		}

		for _, k := range envKeys {
			os.Unsetenv(k)
		}

	}
}

//ClientID     string
func TestNewClient_ClientConfig(t *testing.T) {
	tt := []struct {
		name            string
		cfg             ClientConfig
		errorExpected   bool
		errorStatusCode uint
	}{
		{
			name:            "NonRegistryEnabledBuild",
			errorExpected:   true,
			errorStatusCode: NonRegistryEnabled,
		},
	}

	for _, tc := range tt {

		_, err := NewClient(tc.cfg)
		if tc.errorExpected && err == nil {
			t.Errorf("creating a Client using should error when given only the following client config %v", tc.cfg)
		}

		if err != nil {
			rawErr, ok := err.(*ClientError)
			if !ok {
				t.Errorf("expected a ClientError but got a %t instead", err)
			}

			if rawErr.StatusCode != tc.errorStatusCode {
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
