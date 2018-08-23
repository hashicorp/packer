package common

import (
	"github.com/hashicorp/packer/helper/multistep"
)

// CommHost returns the VM's IP address which should be used to access it by SSH.
func CommHost(state multistep.StateBag) (string, error) {
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
