package openstack

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/rackspace/gophercloud"
	"time"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the server AccessIPv4 setting..
func SSHAddress(csp gophercloud.CloudServersProvider, port int) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		for j := 0; j < 2; j++ {
			s := state.Get("server").(*gophercloud.Server)
			if s.AccessIPv4 != "" {
				return fmt.Sprintf("%s:%d", s.AccessIPv4, port), nil
			}
			if s.AccessIPv6 != "" {
				return fmt.Sprintf("[%s]:%d", s.AccessIPv6, port), nil
			}
			serverState, err := csp.ServerById(s.Id)

			if err != nil {
				return "", err
			}

			state.Put("server", serverState)
			time.Sleep(1 * time.Second)
		}

		return "", errors.New("couldn't determine IP address for server")
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		privateKey := state.Get("privateKey").(string)

		keyring := new(ssh.SimpleKeychain)
		if err := keyring.AddPEMKey(privateKey); err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &gossh.ClientConfig{
			User: username,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		}, nil
	}
}
