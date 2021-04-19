package qemu

import (
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
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
	commHostPort, ok := state.Get("commHostPort").(int)
	if !ok {
		commHostPort = 22
	}
	return commHostPort, nil
}
