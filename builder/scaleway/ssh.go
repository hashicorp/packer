package scaleway

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
	ipAddress := state.Get("server_ip").(string)
	return ipAddress, nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(Config)
	var privateKey string

	var auth []ssh.AuthMethod

	if config.Comm.SSHAgentAuth {
		authSock := os.Getenv("SSH_AUTH_SOCK")
		if authSock == "" {
			return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
		}

		sshAgent, err := net.Dial("unix", authSock)
		if err != nil {
			return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
		}
		auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
		}
	}

	if config.Comm.SSHPassword != "" {
		auth = append(auth,
			ssh.Password(config.Comm.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		)
	}

	if config.Comm.SSHPrivateKey != "" {
		if priv, ok := state.GetOk("privateKey"); ok {
			privateKey = priv.(string)
		}
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	return &ssh.ClientConfig{
		User:            config.Comm.SSHUsername,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
