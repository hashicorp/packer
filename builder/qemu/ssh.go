package qemu

import (
	"log"

	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(host string) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		if host != "" {
			log.Printf("Using host value: %s", host)
			return host, nil
		}

		if guestAddress, ok := state.Get("guestAddress").(string); ok {
			return guestAddress, nil
		}

		return "127.0.0.1", nil
	}
}

func commPort(state multistep.StateBag) (int, error) {
	sshHostPort, ok := state.Get("sshHostPort").(int)
	if !ok {
		sshHostPort = 22
	}
	return int(sshHostPort), nil
}
