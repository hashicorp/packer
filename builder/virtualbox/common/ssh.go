package common

import (
	gossh "code.google.com/p/gosshold/ssh"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"io/ioutil"
	"os"
)

func SSHAddress(state multistep.StateBag) (string, error) {
	sshHostPort := state.Get("sshHostPort").(uint)
	return fmt.Sprintf("127.0.0.1:%d", sshHostPort), nil
}

func SSHConfigFunc(config SSHConfig) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		auth := []gossh.ClientAuth{
			gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
			gossh.ClientAuthKeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHPassword)),
		}

		if config.SSHKeyPath != "" {
			keyring, err := sshKeyToKeyring(config.SSHKeyPath)
			if err != nil {
				return nil, err
			}

			auth = append(auth, gossh.ClientAuthKeyring(keyring))
		}

		return &gossh.ClientConfig{
			User: config.SSHUser,
			Auth: auth,
		}, nil
	}
}

func sshKeyToKeyring(path string) (gossh.ClientKeyring, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keyBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	keyring := new(ssh.SimpleKeychain)
	if err := keyring.AddPEMKey(string(keyBytes)); err != nil {
		return nil, err
	}

	return keyring, nil
}
