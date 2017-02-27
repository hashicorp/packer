package cloudstack

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/multistep"
	packerssh "github.com/mitchellh/packer/communicator/ssh"
	"github.com/xanzy/go-cloudstack/cloudstack"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)

	if config.hostAddress == "" {
		ipAddr, _, err := client.Address.GetPublicIpAddressByID(config.PublicIPAddress)
		if err != nil {
			return "", fmt.Errorf("Failed to retrieve IP address: %s", err)
		}

		config.hostAddress = ipAddr.Ipaddress
	}

	return config.hostAddress, nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(*Config)

	clientConfig := &ssh.ClientConfig{
		User: config.Comm.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Comm.SSHPassword),
			ssh.KeyboardInteractive(
				packerssh.PasswordKeyboardInteractive(config.Comm.SSHPassword)),
		},
	}

	if config.Comm.SSHPrivateKey != "" {
		privateKey, err := ioutil.ReadFile(config.Comm.SSHPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("Error loading configured private key file: %s", err)
		}

		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		clientConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	return clientConfig, nil
}
