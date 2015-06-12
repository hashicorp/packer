package openstack

import (
	"errors"
	"fmt"
	"log"
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
		ip := state.Get("access_ip").(*floatingip.FloatingIP)
		if ip != nil && ip.FixedIP != "" {
			return fmt.Sprintf("%s:%d", ip.FixedIP, port), nil
		}

		if s.AccessIPv4 != "" {
			return fmt.Sprintf("%s:%d", s.AccessIPv4, port), nil
		}

		// Get all the addresses associated with this server. This
		// was taken directly from Terraform.
		for _, networkAddresses := range s.Addresses {
			elements, ok := networkAddresses.([]interface{})
			if !ok {
				log.Printf(
					"[ERROR] Unknown return type for address field: %#v",
					networkAddresses)
				continue
			}

			for _, element := range elements {
				var addr string
				address := element.(map[string]interface{})
				if address["OS-EXT-IPS:type"] == "floating" {
					addr = address["addr"].(string)
				} else {
					if address["version"].(float64) == 4 {
						addr = address["addr"].(string)
					}
				}
				if addr != "" {
					return fmt.Sprintf("%s:%d", addr, port), nil
				}
			}
		}

		s, err := servers.Get(client, s.ID).Extract()
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
