package common

import (
	"fmt"

	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

func CommHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("instance_ip").(string)
	return ipAddress, nil
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the private key
// or password.
func SSHConfig(username, password string) func(state multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		privateKey, hasKey := state.GetOk("privateKey")
		if hasKey {

			signer, err := ssh.ParsePrivateKey([]byte(privateKey.(string)))
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}
			return &ssh.ClientConfig{
				User:            username,
				Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}, nil

		}

		return &ssh.ClientConfig{
			User:            username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
				ssh.KeyboardInteractive(packerssh.PasswordKeyboardInteractive(password)),
			},
		}, nil
	}
}
