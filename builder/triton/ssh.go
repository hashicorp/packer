package triton

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	driver := state.Get("driver").(Driver)
	machineID := state.Get("machine").(string)

	machine, err := driver.GetMachine(machineID)
	if err != nil {
		return "", err
	}

	return machine, nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(Config)

	signer, err := ssh.ParsePrivateKey([]byte(config.Comm.SSHPrivateKey))
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
