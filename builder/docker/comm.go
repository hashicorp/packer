package docker

import (
	"log"

	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using ssh_host value: %s", host)
			return host, nil
		}
		containerId := state.Get("container_id").(string)
		driver := state.Get("driver").(Driver)
		return driver.IPAddress(containerId)
	}
}
