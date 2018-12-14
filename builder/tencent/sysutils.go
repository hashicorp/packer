package tencent

import (
	"errors"
	"fmt"
	"os"
)

// GetRequiredEnvVars takes a map and reads the values of the keys from the environment.
// If the environment variable doesn't exist, it uses the value from the map itself.
// If the environment variable exists, the map is updated with the value retrieved from the environment.
// This function requires SecretId and SecretKey to be defined in the environment as a special case.
// All other variables are retrieved from the keys defined in the map, then retrieved from the environment using the key as the variable name.
// If the environment variable is empty, it retains the value set originally in the map, which could be empty, or the default value.
// This is used currently only during testing.
func GetRequiredEnvVars(varsRequired map[string]string) {
	varsRequired[CSecretId] = ""
	varsRequired[CSecretKey] = ""
	for k, v := range varsRequired {
		envV := os.Getenv(k)
		if envV == "" && v == "" {
			err := errors.New(fmt.Sprintf("%s not defined in environment", k))
			panic(err)
		}
		if envV != "" {
			varsRequired[k] = envV
		} else {
			varsRequired[k] = v
			// os.Setenv(k, v) // this panics on Windows, so don't call it
		}
	}
}

// GetEnvVar returns a value retrieved from environment for the given name.
func GetEnvVar(name string) string {
	return os.Getenv(name)
}
