package qemu

import (
	"fmt"

	"github.com/mitchellh/multistep"
	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/communicator/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func sshAddress(state multistep.StateBag) (string, error) {
	sshHostPort := state.Get("sshHostPort").(uint)
	return fmt.Sprintf("127.0.0.1:%d", sshHostPort), nil
}

func sshConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get("config").(*Config)

	auth := []gossh.AuthMethod{
		gossh.Password(config.Comm.SSHPassword),
		gossh.KeyboardInteractive(
			ssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
	}

	if config.Comm.SSHPrivateKey != "" {
		signer, err := commonssh.FileSigner(config.Comm.SSHPrivateKey)
		if err != nil {
			return nil, err
		}

		auth = append(auth, gossh.PublicKeys(signer))
	}

	return &gossh.ClientConfig{
		User: config.Comm.SSHUsername,
		Auth: auth,
	}, nil
}
