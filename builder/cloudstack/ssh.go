package cloudstack

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	ip, hasIP := state.Get("ipaddress").(string)
	if !hasIP {
		return "", fmt.Errorf("Failed to retrieve IP address")
	}

	return ip, nil
}

func commPort(state multistep.StateBag) (int, error) {
	commPort, hasPort := state.Get("commPort").(int)
	if !hasPort {
		return 0, fmt.Errorf("Failed to retrieve communication port")
	}

	return commPort, nil
}
