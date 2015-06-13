package googlecompute

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

// sshAddress returns the ssh address.
func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*Config)
	ipAddress := state.Get("instance_ip").(string)
	return fmt.Sprintf("%s:%d", ipAddress, config.Comm.SSHPort), nil
}

// sshConfig returns the ssh configuration.
func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(*Config)
	privateKey := state.Get("ssh_private_key").(string)

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return &ssh.ClientConfig{
		User: config.Comm.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}, nil
}
