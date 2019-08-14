package vultr

import (
	"fmt"
	"net"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/vultr/govultr"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	return state.Get("server").(*govultr.Server).MainIP, nil
}

func keyboardInteractive(password string) ssh.KeyboardInteractiveChallenge {
	return func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		for i := range questions {
			answers[i] = password
		}
		return answers, nil
	}
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	b := state.Get("config").(*Config)
	c := b.Comm
	server := state.Get("server").(*govultr.Server)

	config := &ssh.ClientConfig{
		User: "root",
		Auth: make([]ssh.AuthMethod, 0),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil // accept anything
		},
	}

	if b.OSID == SnapshotOSID || b.OSID == CustomOSID {
		config.Auth = append(config.Auth, ssh.Password(c.SSHPassword), keyboardInteractive(c.SSHPassword))
	} else {
		config.Auth = append(config.Auth, ssh.Password(server.DefaultPassword), keyboardInteractive(server.DefaultPassword))
	}

	// Please note that here the private ssh key is completely independent
	// of snapshot, i.e. we attempt to use key login alongside password
	// whenever it appears in the template.
	// This allows us to change root password in provisioner and login with key
	// afterwards when builder needs to gracefully shutdown server.
	if c.SSHPrivateKeyFile != "" {
		privateKey, err := c.ReadSSHPrivateKeyFile()
		if err != nil {
			return nil, fmt.Errorf("Error on reading SSH private key: %s", err)
		}
		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("Error on parsing SSH private key: %s", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	return config, nil
}
