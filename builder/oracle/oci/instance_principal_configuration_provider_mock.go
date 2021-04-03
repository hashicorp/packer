package oci

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/oracle/oci-go-sdk/v36/common"
)

// Mock struct to be used during testing to obtain Instance Principals.
type instancePrincipalConfigurationProviderMock struct {
}

func (p instancePrincipalConfigurationProviderMock) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 1024)
}

func (p instancePrincipalConfigurationProviderMock) KeyID() (string, error) {
	return "some_random_key_id", nil
}

func (p instancePrincipalConfigurationProviderMock) TenancyOCID() (string, error) {
	return "some_random_tenancy", nil
}

func (p instancePrincipalConfigurationProviderMock) UserOCID() (string, error) {
	return "", nil
}

func (p instancePrincipalConfigurationProviderMock) KeyFingerprint() (string, error) {
	return "", nil
}

func (p instancePrincipalConfigurationProviderMock) Region() (string, error) {
	return "some_random_region", nil
}

func (p instancePrincipalConfigurationProviderMock) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{
		AuthType:         common.InstancePrincipal,
		IsFromConfigFile: false,
		OboToken:         nil}, nil
}
