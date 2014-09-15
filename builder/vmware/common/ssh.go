package common

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/packer/communicator/ssh"
	"io/ioutil"
	"os"
)

func SSHAddressFunc(config *SSHConfig, driver Driver) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if config.SSHHost != "" {
			return fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort), nil
		}

		ipAddress, err := driver.IPAddress(state)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s:%d", ipAddress, config.SSHPort), nil
	}
}

func SSHConfigFunc(config *SSHConfig) func(multistep.StateBag) (*gossh.ClientConfig, error) {
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
