// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// Package env provides HCP Packer environment variables.
package env

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func HasHCPAuth() (bool, error) {
	// Client crendential authentication requires the following environment variables be set; `HCP_CLIENT_ID` and `HCP_CLIENT_SECRET`.
	hasClientCredentials := HasHCPClientCredentials()
	// Client certificate authentication requires a valid HCP certificate file placed in either the default location (~/.config/hcp/cred_file.json) or at a location specified in the `HCP_CRED_FILE` env var
	hasCertificate, err := HasHCPCertificateFile()
	if err != nil {
		return false, err
	}
	if hasClientCredentials && hasCertificate {
		fmt.Printf("HCP Client Credentials (HCP_CLIENT_ID/HCP_CLIENT_SECRET environment variables) and certificate (HCP_CRED_FILE environment variable, or certificate located at default path (%s) are both supplied, only one is required. The HCP SDK will determine which authentication mechanism to configure here, it is reccomended to only configure one authentication method", HCPDefaultCredFilePathFull)
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

func HasHCPCertificateFile() (bool, error) {
	envVarCredFile, _ := os.LookupEnv(HCPCredFile)
	var envVarCertExists bool
	var err error
	if envVarCredFile != "" {
		envVarCertExists, err = fileExists(envVarCredFile)
		if err != nil {
			return false, err
		}
	}
	// Get the user's home directory.
	userHome, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to retrieve user's home directory path: %v", err)
	}

	// builds file path ~/.config/hcp/cred_file.json, if we don't parse the home directory os.Stat can't find the default credential path
	defaultCredFilePath := filepath.Join(userHome, HCPDefaultCredFilePath, HCPDefaultCredFile)
	log.Printf("Checking for default HCP credential file at path %s", defaultCredFilePath)
	defaultPathCertExists, err := fileExists(defaultCredFilePath)
	if err != nil {
		return false, err
	}
	log.Printf("Default file found status - %t", defaultPathCertExists)
	if envVarCertExists && defaultPathCertExists {
		fmt.Println("A HCP credential file was found at the default path, and an HCP_CRED_FILE was specified, the HCP SDK will use the HCP_CRED_FILE")
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
