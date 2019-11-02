package linode

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/linode/linodego"
)

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using ssh_host value: %s", host)
			return host, nil
		}

		instance := state.Get("instance").(*linodego.Instance)
		if len(instance.IPv4) == 0 {
			return "", fmt.Errorf("Linode instance %d has no IPv4 addresses!", instance.ID)
		}
		return instance.IPv4[0].String(), nil
	}
}
