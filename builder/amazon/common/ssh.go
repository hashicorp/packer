package common

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/communicator/ssh"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the instance DNS name.
func SSHAddress(port int) func(map[string]interface{}) (string, error) {
	return func(state map[string]interface{}) (string, error) {
		instance := state["instance"].(*ec2.Instance)
		return fmt.Sprintf("%s:%d", instance.DNSName, port), nil
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(map[string]interface{}) (*gossh.ClientConfig, error) {
	return func(state map[string]interface{}) (*gossh.ClientConfig, error) {
		privateKey := state["privateKey"].(string)

		keyring := new(ssh.SimpleKeychain)
		if err := keyring.AddPEMKey(privateKey); err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &gossh.ClientConfig{
			User: username,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		}, nil
	}
}
