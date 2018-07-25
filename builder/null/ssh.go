package null

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		return host, nil
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the specified host via SSH
// private_key_file has precedence over password!
func SSHConfig(useAgent bool, username string, password string, privateKeyFile string) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		if useAgent {
			authSock := os.Getenv("SSH_AUTH_SOCK")
			if authSock == "" {
				return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
			}

			sshAgent, err := net.Dial("unix", authSock)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
			}

			return &gossh.ClientConfig{
				User: username,
				Auth: []gossh.AuthMethod{
					gossh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
				},
				HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			}, nil
		}

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
				HostKeyCallback: gossh.InsecureIgnoreHostKey(),
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
				HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			}, nil
		}
	}
}
