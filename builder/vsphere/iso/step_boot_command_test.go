package iso

import (
	"bytes"
	"context"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"testing"
)

func TestStepBootCommand_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("debug", false)
	state.Put("vm", new(driver.VirtualMachine))

	state.Put("http_port", 2222)
	state.Put("http_ip", "0.0.0.0")

	step := &StepBootCommand{
		Config: &BootConfig{
			BootConfig: bootcommand.BootConfig{
				BootCommand: []string{
					"<leftShiftOn><enter><wait><f6><wait><esc><wait>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
					"<bs><bs><bs>",
					"/install/vmlinuz",
					" initrd=/install/initrd.gz",
					" priority=critical",
					" locale=en_US",
					" file=/media/preseed_hardcoded_ip.cfg",
					" netcfg/get_ipaddress=0.0.0.0",
					" netcfg/get_gateway=0.0.0.0",
					"<enter>",
				},
			},
		},
	}
	step.Run(context.TODO(), state)
}
