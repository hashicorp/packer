package yandex

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("instance_ip").(string)
	return ipAddress, nil
}
