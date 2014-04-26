package null

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"io/ioutil"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address
func SSHAddress(host string, port int) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		return fmt.Sprintf("%s:%d", host, port), nil
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the specified host via SSH
// private_key_file has precedence over password!
func SSHConfig(username string, password string, privateKeyFile string) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {

		if privateKeyFile != "" {
			// key based auth

			bytes, err := ioutil.ReadFile(privateKeyFile)
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}
			privateKey := string(bytes)

			signer, err := gossh.ParsePrivateKey([]byte(privateKey))
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}

			return &gossh.ClientConfig{
				User: username,
				Auth: []gossh.AuthMethod{
					gossh.PublicKeys(signer),
				},
			}, nil
		} else {
			// password based auth

			return &gossh.ClientConfig{
				User: username,
				Auth: []gossh.AuthMethod{
					gossh.Password(password),
					gossh.KeyboardInteractive(
						ssh.PasswordKeyboardInteractive(password)),
				},
			}, nil
		}
	}
}
