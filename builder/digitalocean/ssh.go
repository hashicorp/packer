package digitalocean

import (
	"fmt"
	"golang.org/x/crypto/ssh"

	"github.com/mitchellh/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("droplet_ip").(string)
	return ipAddress, nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(Config)
	privateKey := state.Get("privateKey").(string)

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
