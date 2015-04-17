package openstack_id3

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"time"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the server AccessIPv4 setting..
func SSHAddress(compute_client *gophercloud.ServiceClient, sshinterface string, port int) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		
		s := state.Get("server").(*servers.Server)

		if ip := state.Get("access_ip").(*floatingip.FloatingIP); ip.IP != "" {
			return fmt.Sprintf("%s:%d", ip.IP, port), nil
		} else {
			// We wrap up things here for now
			return "", errors.New("Error parsing SSH addresses")			
		}
/*
		// FIXME: Support for selecting sshinterface. Leaving old code here for now
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
		serverState, err := servers.Get(compute_client, s.ID).Extract()
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
