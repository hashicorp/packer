package common

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			return host, nil
		} else {
			return state.Get("ip").(string), nil
		}
	}
}
