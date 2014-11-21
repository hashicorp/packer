// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package lin

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"code.google.com/p/go.crypto/ssh"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
)

// SSHAddress returns a function that can be given to the SSH communicator
func SSHAddress(state multistep.StateBag) (string, error) {
	azureVmAddr := state.Get(constants.AzureVmAddr).(string)
	return azureVmAddr, nil
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		privateKey := state.Get(constants.PrivateKey).(string)

		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		}, nil
	}
}

