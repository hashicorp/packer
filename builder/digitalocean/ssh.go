package digitalocean

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("droplet_ip").(string)
	return ipAddress, nil
}
