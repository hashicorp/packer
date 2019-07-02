package googlecompute

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	config := state.Get("config").(*Config)
	if config.Comm.SSHHost != "" {
		return config.Comm.SSHHost, nil
	}
	ipAddress := state.Get("instance_ip").(string)
	return ipAddress, nil
}
