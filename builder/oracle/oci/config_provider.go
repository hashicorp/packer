package oci

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/oracle/oci-go-sdk/common"
)

// rawConfigurationProvider allows a user to simply construct a configuration
// provider from raw values. It errors on access when those values are empty.
type rawConfigurationProvider struct {
	tenancy              string
	user                 string
	region               string
	fingerprint          string
	privateKey           string
	privateKeyPassphrase *string
}

// NewRawConfigurationProvider will create a rawConfigurationProvider.
func NewRawConfigurationProvider(tenancy, user, region, fingerprint, privateKey string, privateKeyPassphrase *string) common.ConfigurationProvider {
	return rawConfigurationProvider{tenancy, user, region, fingerprint, privateKey, privateKeyPassphrase}
}

func (p rawConfigurationProvider) PrivateRSAKey() (key *rsa.PrivateKey, err error) {
	return common.PrivateKeyFromBytes([]byte(p.privateKey), p.privateKeyPassphrase)
}

func (p rawConfigurationProvider) KeyID() (keyID string, err error) {
	tenancy, err := p.TenancyOCID()
	if err != nil {
		return
	}

	user, err := p.UserOCID()
	if err != nil {
		return
	}

	fingerprint, err := p.KeyFingerprint()
	if err != nil {
		return
	}

	return fmt.Sprintf("%s/%s/%s", tenancy, user, fingerprint), nil
}

func (p rawConfigurationProvider) TenancyOCID() (string, error) {
	if p.tenancy == "" {
		return "", errors.New("no tenancy provided")
	}
	return p.tenancy, nil
}

func (p rawConfigurationProvider) UserOCID() (string, error) {
	if p.user == "" {
		return "", errors.New("no user provided")
	}
	return p.user, nil
}

func (p rawConfigurationProvider) KeyFingerprint() (string, error) {
	if p.fingerprint == "" {
		return "", errors.New("no fingerprint provided")
	}
	return p.fingerprint, nil
}

func (p rawConfigurationProvider) Region() (string, error) {
	if p.region == "" {
		return "", errors.New("no region provided")
	}
	return p.region, nil
}
