// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

package auth

import (
	"crypto/rsa"
	"fmt"
	"github.com/oracle/oci-go-sdk/v36/common"
)

type instancePrincipalDelegationTokenConfigurationProvider struct {
	instancePrincipalKeyProvider instancePrincipalKeyProvider
	delegationToken              string
}

//InstancePrincipalDelegationTokenConfigurationProvider returns a configuration for obo token instance principals
func InstancePrincipalDelegationTokenConfigurationProvider(delegationToken *string) (common.ConfigurationProvider, error) {
	if delegationToken == nil || len(*delegationToken) == 0 {
		return nil, fmt.Errorf("failed to create a delagationTokenConfigurationProvider: token is a mondatory input paras")
	}
	return newInstancePrincipalDelegationTokenConfigurationProvider(delegationToken, nil)
}

func newInstancePrincipalDelegationTokenConfigurationProvider(delegationToken *string, modifier func(common.HTTPRequestDispatcher) (common.HTTPRequestDispatcher,
	error)) (common.ConfigurationProvider, error) {

	keyProvider, err := newInstancePrincipalKeyProvider(modifier)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new key provider for instance principal: %s", err.Error())
	}
	return instancePrincipalDelegationTokenConfigurationProvider{*keyProvider, *delegationToken}, err
}

func (p instancePrincipalDelegationTokenConfigurationProvider) getInstancePrincipalDelegationTokenConfigurationProvider() (instancePrincipalDelegationTokenConfigurationProvider, error) {
	return p, nil
}

func (p instancePrincipalDelegationTokenConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return p.instancePrincipalKeyProvider.PrivateRSAKey()
}

func (p instancePrincipalDelegationTokenConfigurationProvider) KeyID() (string, error) {
	return p.instancePrincipalKeyProvider.KeyID()
}

func (p instancePrincipalDelegationTokenConfigurationProvider) TenancyOCID() (string, error) {
	return p.instancePrincipalKeyProvider.TenancyOCID()
}

func (p instancePrincipalDelegationTokenConfigurationProvider) UserOCID() (string, error) {
	return "", nil
}

func (p instancePrincipalDelegationTokenConfigurationProvider) KeyFingerprint() (string, error) {
	return "", nil
}

func (p instancePrincipalDelegationTokenConfigurationProvider) Region() (string, error) {
	region := p.instancePrincipalKeyProvider.RegionForFederationClient()
	return string(region), nil
}

func (p instancePrincipalDelegationTokenConfigurationProvider) AuthType() (common.AuthConfig, error) {
	token := p.delegationToken
	return common.AuthConfig{common.InstancePrincipalDelegationToken, false, &token}, nil
}
