package common

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/multistep"
	packerssh "github.com/mitchellh/packer/communicator/ssh"
	"io/ioutil"
	"os"
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
			ssh.Password(config.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.SSHPassword)),
		}

		if config.SSHKeyPath != "" {
			signer, err := sshKeyToSigner(config.SSHKeyPath)
			if err != nil {
				return nil, err
			}

			auth = append(auth, ssh.PublicKeys(signer))
		}

		return &ssh.ClientConfig{
			User: config.SSHUser,
			Auth: auth,
		}, nil
	}
}

func sshKeyToSigner(path string) (ssh.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return signer, nil
}
