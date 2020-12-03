package common

import (
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

// CommHost returns the VM's IP address which should be used to access it by SSH.
func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}
		vmName := state.Get("vmName").(string)
		driver := state.Get("driver").(Driver)

		mac, err := driver.MAC(vmName)
		if err != nil {
			return "", err
		}

		ip, err := driver.IPAddress(mac)
		if err != nil {
			return "", err
		}

		return ip, nil
	}
}
