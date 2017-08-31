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

	var auth []ssh.AuthMethod

	if config.Comm.SSHPrivateKey != "" {
		privateKey, err := ioutil.ReadFile(config.Comm.SSHPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("Error loading configured private key file: %s", err)
		}

		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		auth = []ssh.AuthMethod{
			ssh.Password(config.Comm.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		}
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.Comm.SSHUsername,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	clientConfig.Auth = auth

	return clientConfig, nil
}
