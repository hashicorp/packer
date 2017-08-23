package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

func (d *Driver) FindVM(name string) (*object.VirtualMachine, error) {
	return d.finder.VirtualMachine(d.ctx, name)
}

func (d *Driver) VMInfo(vm *object.VirtualMachine, params ...string) (*mo.VirtualMachine, error){
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var vmInfo mo.VirtualMachine
	err := vm.Properties(d.ctx, vm.Reference(), p, &vmInfo)
	if err != nil {
		return nil, err
	}
	return &vmInfo, nil
}
