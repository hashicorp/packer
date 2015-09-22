package brkt

import (
	"fmt"
	"log"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

// SSHost is a function that returns the IP of the deployed instance
func SSHost(state multistep.StateBag) (string, error) {
	instance, ok := state.Get("instance").(*brkt.Instance)
	if !ok {
		return "", fmt.Errorf("error getting workload")
	}

	log.Printf("instance IP: %s", instance.Data.IpAddress)

	return instance.Data.IpAddress, nil
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the provided
// private key
func SSHConfig(username string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		privateKey := state.Get("privateKey").(string) // ad hoc key
		privateKeyBastion := state.Get("privateKeyBastion").(string)

		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("error setting up SSH config: %s", err)
		}

		authMethods := []ssh.AuthMethod{ssh.PublicKeys(signer)}

		// if we have a Bastion, that key should also be supported
		if privateKeyBastion != "" {
			signerBastion, err := ssh.ParsePrivateKey([]byte(privateKeyBastion))
			if err != nil {
				return nil, fmt.Errorf("error setting up SSH config for Bastion: %s", err)
			}

			authMethods = append(authMethods, ssh.PublicKeys(signerBastion))

		}

		return &ssh.ClientConfig{
			User: username,
			Auth: authMethods,
		}, nil
	}
}
