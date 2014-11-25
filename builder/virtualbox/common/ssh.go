package common

import (
	"fmt"

	gossh "code.google.com/p/go.crypto/ssh"
	"github.com/mitchellh/multistep"
	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/communicator/ssh"
)

func SSHAddress(state multistep.StateBag) (string, error) {
	sshHostPort := state.Get("sshHostPort").(uint)
	return fmt.Sprintf("127.0.0.1:%d", sshHostPort), nil
}

func SSHConfigFunc(config SSHConfig) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		auth := []gossh.AuthMethod{
			gossh.Password(config.SSHPassword),
			gossh.KeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHPassword)),
		}

		if config.SSHKeyPath != "" {
			signer, err := commonssh.FileSigner(config.SSHKeyPath)
			if err != nil {
				return nil, err
			}

			auth = append(auth, gossh.PublicKeys(signer))
		}

		return &gossh.ClientConfig{
			User: config.SSHUser,
			Auth: auth,
		}, nil
	}
}
