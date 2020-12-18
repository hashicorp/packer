package common

import (
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {

		// Skip IP auto detection if the configuration has an ssh host configured.
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}

		vmName := state.Get("vmName").(string)
		driver := state.Get("driver").(Driver)

		mac, err := driver.Mac(vmName)
		if err != nil {
			return "", err
		}

		ip, err := driver.IpAddress(mac)
		if err != nil {
			return "", err
		}

		return ip, nil
	}
}
