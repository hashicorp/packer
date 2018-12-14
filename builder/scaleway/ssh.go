package scaleway

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("server_ip").(string)
	return ipAddress, nil
}
