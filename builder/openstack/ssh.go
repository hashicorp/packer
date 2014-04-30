package openstack

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/rackspace/gophercloud"
	"time"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the server AccessIPv4 setting..
func SSHAddress(csp gophercloud.CloudServersProvider, port int) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		s := state.Get("server").(*gophercloud.Server)

		if ip := state.Get("access_ip").(gophercloud.FloatingIp); ip.Ip != "" {
			return fmt.Sprintf("%s:%d", ip.Ip, port), nil
		}

		ip_pools, err := s.AllAddressPools()
		if err != nil {
			return "", errors.New("Error parsing SSH addresses")
		}
		for pool, addresses := range ip_pools {
			if pool != "" {
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
		time.Sleep(1 * time.Second)

		return "", errors.New("couldn't determine IP address for server")
	}
}

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the instance created over SSH using the generated
// private key.
func SSHConfig(username string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		privateKey := state.Get("privateKey").(string)

		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("Error setting up SSH config: %s", err)
		}

		return &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		}, nil
	}
}
