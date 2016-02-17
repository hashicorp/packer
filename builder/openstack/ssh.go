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

// CommHost looks up the host for the communicator.
func CommHost(
	client *gophercloud.ServiceClient,
	sshinterface string,
	sshipversion string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		s := state.Get("server").(*servers.Server)

		// If we have a specific interface, try that
		if sshinterface != "" {
			if addr := sshAddrFromPool(s, sshinterface, sshipversion); addr != "" {
				log.Printf("[DEBUG] Using IP address %s from specified interface %s to connect", addr, sshinterface)
				return addr, nil
			}
		}

		// If we have a floating IP, use that
		ip := state.Get("access_ip").(*floatingip.FloatingIP)
		if ip != nil && ip.IP != "" {
			log.Printf("[DEBUG] Using floating IP %s to connect", ip.IP)
			return ip.IP, nil
		}

		if s.AccessIPv4 != "" {
			log.Printf("[DEBUG] Using AccessIPv4 %s to connect", s.AccessIPv4)
			return s.AccessIPv4, nil
		}

		// Try to get it from the requested interface
		if addr := sshAddrFromPool(s, sshinterface, sshipversion); addr != "" {
			log.Printf("[DEBUG] Using IP address %s to connect", addr)
			return addr, nil
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

func sshAddrFromPool(s *servers.Server, desired string, sshIPVersion string) string {
	// Get all the addresses associated with this server. This
	// was taken directly from Terraform.
	for pool, networkAddresses := range s.Addresses {
		// If we have an SSH interface specified, skip it if no match
		if desired != "" && pool != desired {
			log.Printf(
				"[INFO] Skipping pool %s, doesn't match requested %s",
				pool, desired)
			continue
		}

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
			} else if sshIPVersion == "4" {
				if address["version"].(float64) == 4 {
					addr = address["addr"].(string)
				}
			} else if sshIPVersion == "6" {
				if address["version"].(float64) == 6 {
					addr = fmt.Sprintf("[%s]", address["addr"].(string))
				}
			} else {
				if address["version"].(float64) == 6 {
					addr = fmt.Sprintf("[%s]", address["addr"].(string))
				} else {
					addr = address["addr"].(string)
				}
			}

			if addr != "" {
				log.Printf("[DEBUG] Detected address: %s", addr)
				return addr
			}
		}
	}

	return ""
}
