package common

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost(config *SSHConfig) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		driver := state.Get("driver").(Driver)

		if config.Comm.SSHHost != "" {
			return config.Comm.SSHHost, nil
		}

		ipAddress, err := driver.GuestIP(state)
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			return "", fmt.Errorf("IP lookup failed: %s", err)
		}

		if ipAddress == "" {
			log.Println("IP is blank, no IP yet.")
			return "", errors.New("IP is blank")
		}

		log.Printf("Detected IP: %s", ipAddress)
		return ipAddress, nil
	}
}
