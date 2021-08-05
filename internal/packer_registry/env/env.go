package env

import (
	"os"
	"strings"
)

func HasClientID() bool {
	_, ok := os.LookupEnv(HCPClientID)
	return ok
}

func HasClientSecret() bool {
	_, ok := os.LookupEnv(HCPClientSecret)
	return ok
}

func HasPackerRegistryBucket() bool {
	_, ok := os.LookupEnv(HCPPackerBucket)
	return ok
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

func IsPAREnabled() bool {
	val, ok := os.LookupEnv(HCPPackerRegistry)
	return ok && strings.ToLower(val) != "off" && val != "0"
}
