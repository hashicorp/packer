// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package env provides HCP Packer environment variables.
package env

import (
	"os"
	"strings"
)

func HasProjectID() bool {
	return hasEnvVar(HCPProjectID)
}

func HasOrganizationID() bool {
	return hasEnvVar(HCPOrganizationID)
}

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

// IsHCPExplicitelyEnabled returns true if the client enabled HCP_PACKER_REGISTRY explicitely, i.e. it is defined and not 0 or off
func IsHCPExplicitelyEnabled() bool {
	_, ok := os.LookupEnv(HCPPackerRegistry)
	return ok && !IsHCPDisabled()
}
