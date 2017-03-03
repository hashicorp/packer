package ecs

import (
	"fmt"
	"github.com/mitchellh/multistep"
	packerssh "github.com/mitchellh/packer/communicator/ssh"
	"golang.org/x/crypto/ssh"
)

type alicloudSSHHelper interface {
}

// SSHHost returns a function that can be given to the SSH communicator
func SSHHost(e alicloudSSHHelper, private bool) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		ipAddress := state.Get("ipaddress").(string)
		return ipAddress, nil
	}
}

func SSHConfig(useAgent bool, username, password string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		privateKey, hasKey := state.GetOk("privateKey")
		if hasKey {

			signer, err := ssh.ParsePrivateKey([]byte(privateKey.(string)))
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}
			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
			}, nil

		} else {
			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.Password(password),
					ssh.KeyboardInteractive(
						packerssh.PasswordKeyboardInteractive(password)),
				}}, nil
		}
	}
}
