package docker

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	containerId := state.Get("container_id").(string)
	driver := state.Get("driver").(Driver)
	return driver.IPAddress(containerId)
}
