package env

import (
	"os"
	"strings"
)

func HasClientID() bool {
	return hasEnvVar(HCPClientID)
}

func HasClientSecret() bool {
	return hasEnvVar(HCPClientSecret)
}

func HasPackerRegistryBucket() bool {
	return hasEnvVar(HCPPackerBucket)
}

func hasEnvVar(varName string) bool {
	val, ok := os.LookupEnv(varName)
	if !ok {
		return false
	}
	return val != ""
}

func HasHCPCredentials() bool {
	checks := []func() bool{
		HasClientID,
		HasClientSecret,
	}

	for _, check := range checks {
		if !check() {
			return false
		}
	}

	return true
}

func IsHCPDisabled() bool {
	hcp, ok := os.LookupEnv(HCPPackerRegistry)
	return ok && strings.ToLower(hcp) == "off" || hcp == "0"
}
