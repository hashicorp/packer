package qemu

import (
	"log"

	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using ssh_host value: %s", host)
			return host, nil
		}

		return "127.0.0.1", nil
	}
}

func commPort(state multistep.StateBag) (int, error) {
	sshHostPort := state.Get("sshHostPort").(int)
	return int(sshHostPort), nil
}
