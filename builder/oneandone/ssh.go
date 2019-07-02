package oneandone

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

		ipAddress := state.Get("server_ip").(string)
		return ipAddress, nil
	}
}
