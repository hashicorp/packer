package common

import (
	"log"

	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {

		if host != "" {
			log.Printf("Using ssh_host value: %s", host)
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
