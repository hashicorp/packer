package vagrant

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepPackage struct {
	SkipPackage bool
	Include     []string
	Vagrantfile string
}

func (s *StepPackage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packer.Ui)

	if s.SkipPackage {
		ui.Say("skip_package flag set; not going to call Vagrant package on this box.")
		return multistep.ActionContinue
	}
	ui.Say("Packaging box...")
	packageArgs := []string{}
	if len(s.Include) > 0 {
		packageArgs = append(packageArgs, "--include", strings.Join(s.Include, ","))
	}
	if s.Vagrantfile != "" {
		packageArgs = append(packageArgs, "--vagrantfile", s.Vagrantfile)
	}

	err := driver.Package(packageArgs)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPackage) Cleanup(state multistep.StateBag) {
}
