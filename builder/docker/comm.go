package docker

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/mitchellh/multistep"
	gossh "golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	containerId := state.Get("container_id").(string)
	driver := state.Get("driver").(Driver)
	return driver.IPAddress(containerId)
}

func sshConfig(comm *communicator.Config) func(state multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		if comm.SSHPrivateKey != "" {
			// key based auth
			bytes, err := ioutil.ReadFile(comm.SSHPrivateKey)
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}
			privateKey := string(bytes)

			signer, err := gossh.ParsePrivateKey([]byte(privateKey))
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}

			return &gossh.ClientConfig{
				User: comm.SSHUsername,
				Auth: []gossh.AuthMethod{
					gossh.PublicKeys(signer),
				},
				HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			}, nil
		} else {
			// password based auth
			return &gossh.ClientConfig{
				User: comm.SSHUsername,
				Auth: []gossh.AuthMethod{
					gossh.Password(comm.SSHPassword),
					gossh.KeyboardInteractive(
						ssh.PasswordKeyboardInteractive(comm.SSHPassword)),
				},
				HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			}, nil
		}
	}
}
