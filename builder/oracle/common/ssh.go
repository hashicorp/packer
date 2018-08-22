package common

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("instance_ip").(string)
	return ipAddress, nil
}
