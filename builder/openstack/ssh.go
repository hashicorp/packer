package openstack

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/rackspace/gophercloud"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the server AccessIPv4 setting..
func SSHAddress(csp gophercloud.CloudServersProvider, port int,
	specify_ip_pool string) func(multistep.StateBag) (string, error) {

	return func(state multistep.StateBag) (string, error) {
		s := state.Get("server").(*gophercloud.Server)
		ip_pools, err := s.AllAddressPools()
		if err != nil {
			return "", errors.New("Error parsing SSH addresses")
		}
		for pool, addresses := range ip_pools {
			if specify_ip_pool != "" {
				if pool == specify_ip_pool {
					for _, address := range addresses {
						if address.Addr != "" {
							return fmt.Sprintf("%s:%d", address.Addr, port), nil
						}
					}
				}
			} else if pool != "" {
				for _, address := range addresses {
					if address.Addr != "" {
						return fmt.Sprintf("%s:%d", address.Addr, port), nil
					}
				}
			}
		}

		serverState, err := csp.ServerById(s.Id)

		if err != nil {
			return "", err
		}

		state.Put("server", serverState)

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
