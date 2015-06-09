package digitalocean

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(Config)
	ipAddress := state.Get("droplet_ip").(string)
	return fmt.Sprintf("%s:%d", ipAddress, config.SSHPort), nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(Config)
	privateKey := state.Get("privateKey").(string)

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return &ssh.ClientConfig{
		User: config.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}, nil
}
