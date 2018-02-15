package cloudstack

import (
	"fmt"
	"net"
	"os"

	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func commHost(state multistep.StateBag) (string, error) {
	ip, hasIP := state.Get("ipaddress").(string)
	if !hasIP {
		return "", fmt.Errorf("Failed to retrieve IP address")
	}

	return ip, nil
}

func commPort(state multistep.StateBag) (int, error) {
	commPort, hasPort := state.Get("commPort").(int)
	if !hasPort {
		return 0, fmt.Errorf("Failed to retrieve communication port")
	}

	return commPort, nil
}

func sshConfig(useAgent bool, username, password string) func(state multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		if useAgent {
			authSock := os.Getenv("SSH_AUTH_SOCK")
			if authSock == "" {
				return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
			}

			sshAgent, err := net.Dial("unix", authSock)
			if err != nil {
				return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
			}

			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}, nil
		}

		privateKey, hasKey := state.GetOk("privateKey")

		if hasKey {
			signer, err := ssh.ParsePrivateKey([]byte(privateKey.(string)))
			if err != nil {
				return nil, fmt.Errorf("Error setting up SSH config: %s", err)
			}

			return &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}, nil

		} else {

			return &ssh.ClientConfig{
				User:            username,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				Auth: []ssh.AuthMethod{
					ssh.Password(password),
					ssh.KeyboardInteractive(
						packerssh.PasswordKeyboardInteractive(password)),
				}}, nil
		}
	}
}
