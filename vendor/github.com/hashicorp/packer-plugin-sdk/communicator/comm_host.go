package communicator

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// CommHost determines the IP address of the cloud instance that Packer
// should connect to. A custom CommHost function can be implemented in each
// builder if need be; this is a generic function that should work for most
// cloud builders.
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
