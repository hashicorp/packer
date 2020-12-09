package triton

import (
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}

		driver := state.Get("driver").(Driver)
		machineID := state.Get("machine").(string)

		machine, err := driver.GetMachineIP(machineID)
		if err != nil {
			return "", err
		}

		return machine, nil
	}
}
