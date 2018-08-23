package triton

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	driver := state.Get("driver").(Driver)
	machineID := state.Get("machine").(string)

	machine, err := driver.GetMachineIP(machineID)
	if err != nil {
		return "", err
	}

	return machine, nil
}
