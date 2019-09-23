package vminstance

import (
	"github.com/hashicorp/packer/builder/zstack/zstacktype"
	"github.com/hashicorp/packer/helper/multistep"
)

func getHostIp(state multistep.StateBag) (string, error) {
	vm := state.Get(Vm).(*zstacktype.VmInstance)
	return vm.PublicIp, nil
}

func getVmUuid(state multistep.StateBag) string {
	vm := state.Get(Vm).(*zstacktype.VmInstance)
	return vm.Uuid
}
