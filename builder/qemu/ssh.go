package qemu

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	return "127.0.0.1", nil
}

func commPort(state multistep.StateBag) (uint, error) {
	sshHostPort := state.Get("sshHostPort").(uint)
	return sshHostPort, nil
}
