package googlecompute

import (
	"fmt"

	gossh "code.google.com/p/go.crypto/ssh"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
)

// sshAddress returns the ssh address.
func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*Config)
	ipAddress := state.Get("instance_ip").(string)
	return fmt.Sprintf("%s:%d", ipAddress, config.SSHPort), nil
}

// sshConfig returns the ssh configuration.
func sshConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get("config").(*Config)
	privateKey := state.Get("ssh_private_key").(string)

	keyring := new(ssh.SimpleKeychain)
	if err := keyring.AddPEMKey(privateKey); err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	sshConfig := &gossh.ClientConfig{
		User: config.SSHUsername,
		Auth: []gossh.ClientAuth{gossh.ClientAuthKeyring(keyring)},
	}

	return sshConfig, nil
}
