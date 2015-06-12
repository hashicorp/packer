package openstack

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"golang.org/x/crypto/ssh"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the server AccessIPv4 setting..
func SSHAddress(
	client *gophercloud.ServiceClient,
	sshinterface string, port int) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		s := state.Get("server").(*servers.Server)

		// If we have a floating IP, use that
		if ip := state.Get("access_ip").(*floatingip.FloatingIP); ip.FixedIP != "" {
			return fmt.Sprintf("%s:%d", ip.FixedIP, port), nil
		}

		if s.AccessIPv4 != "" {
			return fmt.Sprintf("%s:%d", s.AccessIPv4, port), nil
		}

		// Get all the addresses associated with this server
		/*
			ip_pools, err := s.AllAddressPools()
			if err != nil {
				return "", errors.New("Error parsing SSH addresses")
			}
			for pool, addresses := range ip_pools {
				if sshinterface != "" {
					if pool != sshinterface {
						continue
					}
				}
				if pool != "" {
					for _, address := range addresses {
						if address.Addr != "" && address.Version == 4 {
							return fmt.Sprintf("%s:%d", address.Addr, port), nil
						}
					}
				}
			}
		*/

		result := servers.Get(client, s.ID)
		err := result.Err
		if err == nil {
			s, err = result.Extract()
		}
		if err != nil {
			return "", err
		}

		state.Put("server", s)
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
