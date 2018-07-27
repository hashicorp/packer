package qemu

import (
	"fmt"
	"net"
	"os"

	commonssh "github.com/hashicorp/packer/common/ssh"
	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func commHost(state multistep.StateBag) (string, error) {
	return "127.0.0.1", nil
}

func commPort(state multistep.StateBag) (int, error) {
	sshHostPort := state.Get("sshHostPort").(uint)
	return int(sshHostPort), nil
}

func sshConfig(state multistep.StateBag) (*gossh.ClientConfig, error) {
	config := state.Get("config").(*Config)

	var auth []gossh.AuthMethod

	if config.Comm.SSHAgentAuth {
		authSock := os.Getenv("SSH_AUTH_SOCK")
		if authSock == "" {
			return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
		}

		sshAgent, err := net.Dial("unix", authSock)
		if err != nil {
			return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
		}
		auth = []gossh.AuthMethod{
			gossh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
		}
	}

	if config.Comm.SSHPassword != "" {
		auth = append(auth,
			gossh.Password(config.Comm.SSHPassword),
			gossh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		)
	}

	if config.Comm.SSHPrivateKey != "" {
		signer, err := commonssh.FileSigner(config.Comm.SSHPrivateKey)
		if err != nil {
			return nil, err
		}

		auth = append(auth, gossh.PublicKeys(signer))
	}

	return &gossh.ClientConfig{
		User:            config.Comm.SSHUsername,
		Auth:            auth,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}, nil
}
