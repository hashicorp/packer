package lin

import (
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
)

func SSHHost(state multistep.StateBag) (string, error) {
	host := state.Get(constants.SSHHost).(string)
	return host, nil
}
