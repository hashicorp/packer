package env

import "os"

func HasClientID() bool {
	_, ok := os.LookupEnv(HCPClientID)
	return ok
}

func HasClientSecret() bool {
	_, ok := os.LookupEnv(HCPClientSecret)
	return ok
}

func HasPackerRegistryDestionation() bool {
	_, ok := os.LookupEnv(HCPPackerRegistry)
	return ok
}

func HasPackerRegistryBucket() bool {
	_, ok := os.LookupEnv(HCPPackerBucket)
	return ok
}

func InPARMode() bool {
	checks := []func() bool{
		HasClientID,
		HasClientSecret,
		HasPackerRegistryDestionation,
	}

	for _, check := range checks {
		if !check() {
			return false
		}
	}

	return true
}
