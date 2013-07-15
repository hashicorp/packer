package ebs

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/communicator/ssh"
)

func sshAddress(state map[string]interface{}) (string, error) {
	config := state["config"].(config)
	instance := state["instance"].(*ec2.Instance)
	return fmt.Sprintf("%s:%d", instance.DNSName, config.SSHPort), nil
}

func sshConfig(state map[string]interface{}) (*gossh.ClientConfig, error) {
	config := state["config"].(config)
	privateKey := state["privateKey"].(string)

	keyring := new(ssh.SimpleKeychain)
	if err := keyring.AddPEMKey(privateKey); err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return &gossh.ClientConfig{
		User: config.SSHUsername,
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthKeyring(keyring),
		},
	}, nil
}
