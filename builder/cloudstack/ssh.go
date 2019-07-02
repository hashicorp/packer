package cloudstack

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using ssh_host value: %s", host)
			return host, nil
		}
		ip, hasIP := state.Get("ipaddress").(string)
		if !hasIP {
			return "", fmt.Errorf("Failed to retrieve IP address")
		}

		return ip, nil
	}
}

func commPort(state multistep.StateBag) (int, error) {
	commPort, hasPort := state.Get("commPort").(int)
	if !hasPort {
		return 0, fmt.Errorf("Failed to retrieve communication port")
	}

	return commPort, nil
}
