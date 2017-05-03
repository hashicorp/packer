package openstack_id3

import (
	"errors"
	"fmt"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

// SSHAddress returns a function that can be given to the SSH communicator
// for determining the SSH address based on the server AccessIPv4 setting..
func SSHAddress(compute_client *gophercloud.ServiceClient, sshinterface string, port int) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {

		s := state.Get("server").(*servers.Server)

		if ip := state.Get("access_ip").(floatingip.FloatingIP); ip.IP != "" {
			return fmt.Sprintf("%s:%d", ip.IP, port), nil
		}

		ipPools := make(map[string][]servers.Address, len(s.Addresses))
		err := mapstructure.Decode(s.Addresses, &ipPools)
		if err != nil {
			return "", fmt.Errorf("Error parsing ip pools from the server: %s", err)
		}

		for pool, addresses := range ipPools {
			if sshinterface != "" {
				if pool != sshinterface {
					continue
				}
			}
			if pool != "" {
				for _, address := range addresses {
					if address.Address != "" && address.Version == 4 {
						return fmt.Sprintf("%s:%d", address.Address, port), nil
					}
				}
			}
		}

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
