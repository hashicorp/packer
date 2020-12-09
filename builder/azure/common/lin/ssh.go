package lin

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/builder/azure/common/constants"
)

func SSHHost(state multistep.StateBag) (string, error) {
	host := state.Get(constants.SSHHost).(string)
	return host, nil
}
