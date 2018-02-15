package oneandone

import (
	"fmt"

	"github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	gossh "golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("server_ip").(string)
	return ipAddress, nil
}

func sshConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get("config").(*Config)
	var privateKey string

	var auth []gossh.AuthMethod

	if config.Comm.SSHPassword != "" {
		auth = []gossh.AuthMethod{
			gossh.Password(config.Comm.SSHPassword),
			gossh.KeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		}
	}

	if config.Comm.SSHPrivateKey != "" {
		if priv, ok := state.GetOk("privateKey"); ok {
			privateKey = priv.(string)
		}
		signer, err := gossh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}
		if err != nil {
			return nil, err
		}

		auth = append(auth, gossh.PublicKeys(signer))
	}
	return &gossh.ClientConfig{
		User:            config.Comm.SSHUsername,
		Auth:            auth,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}, nil
}
