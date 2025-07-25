// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package env provides HCP Packer environment variables.
package env

import (
	"errors"
	"log"
	"os"
	"strings"
)

func HasHCPAuth() (bool, error) {
	hasClientCredentials := HasHCPClientCredentials()
	hasCertificate, err := HasHCPCertificateFile()
	if err != nil {
		return false, err
	}
	if hasClientCredentials && hasCertificate {
		log.Print("HCP Client Credentials and certificate are both supplied, only one is required. Packer will allow the SDK to determine which authentication mechanism to configure here.")
	}
	return (hasClientCredentials || hasCertificate), nil
}

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

func HasHCPClientCredentials() bool {
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

// Checks
func HasHCPCertificateFile() (bool, error) {
	credFilePath := HCPDefaultCredFilePath
	envVarCredFile, _ := os.LookupEnv(HCPCredFile)
	var envVarCertExists bool
	var err error
	if envVarCredFile != "" {
		envVarCertExists, err = fileExists(envVarCredFile)
		if err != nil {
			return false, err
		}
	}
	defaultPathCertExists, err := fileExists(credFilePath)
	if err != nil {
		return false, err
	}

	if envVarCertExists && defaultPathCertExists {
		log.Print("A HCP credential file was found at the default path, and an HCP_CRED_FILE was specified, the HCP SDK will use the HCP_CRED_FILE")
	}
	if envVarCertExists || defaultPathCertExists {
		return true, nil
	}
	return false, nil
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

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // Path exists, no error
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil // Path does not exist
	}
	return false, err // Another error occurred
}
