package common

import (
	"fmt"

	"github.com/mitchellh/multistep"
	commonssh "github.com/mitchellh/packer/common/ssh"
	packerssh "github.com/mitchellh/packer/communicator/ssh"
	"golang.org/x/crypto/ssh"
)

func SSHAddress(state multistep.StateBag) (string, error) {
	vmName := state.Get("vmName").(string)
	driver := state.Get("driver").(Driver)

	mac, err := driver.Mac(vmName)
	if err != nil {
		return "", err
	}

	ip, err := driver.IpAddress(mac)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:22", ip), nil
}

func SSHConfigFunc(config SSHConfig) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		auth := []ssh.AuthMethod{
			ssh.Password(config.Comm.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		}

		if config.SSHKeyPath != "" {
			signer, err := commonssh.FileSigner(config.Comm.SSHPrivateKey)
			if err != nil {
				return nil, err
			}

			auth = append(auth, ssh.PublicKeys(signer))
		}

		return &ssh.ClientConfig{
			User: config.Comm.SSHUsername,
			Auth: auth,
		}, nil
	}
}
