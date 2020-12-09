package communicator

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// Generic commHost function that should work for most cloud builders.
func CommHost(host string, statebagKey string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}
		ipAddress, hasIP := state.Get(statebagKey).(string)
		if !hasIP {
			return "", fmt.Errorf("Failed to retrieve IP address.")
		}
		return ipAddress, nil
	}
}
