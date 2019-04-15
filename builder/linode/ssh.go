package linode

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/linode/linodego"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	instance := state.Get("instance").(*linodego.Instance)
	if len(instance.IPv4) == 0 {
		return "", fmt.Errorf("Linode instance %d has no IPv4 addresses!", instance.ID)
	}
	return instance.IPv4[0].String(), nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	return &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(state.Get("root_pass").(string)),
		},
	}, nil
}
