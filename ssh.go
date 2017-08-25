package main

import (
	"fmt"
	"io/ioutil"

	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	return state.Get("ip").(string), nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(*Config)

	clientConfig := &ssh.ClientConfig{
		User: config.Config.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Config.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Config.SSHPassword)),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if config.Config.SSHPrivateKey != "" {
		privateKey, err := ioutil.ReadFile(config.Config.SSHPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("Error loading configured private key file: %s", err)
		}

		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	return clientConfig, nil
}
