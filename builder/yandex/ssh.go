package yandex

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func CommHost(state multistep.StateBag) (string, error) {
	ipAddress := state.Get("instance_ip").(string)
	return ipAddress, nil
}
