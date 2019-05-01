package common

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step exports a VM built on ESXi using ovftool
//
// Uses:
//   display_name string
type StepExport struct {
	Format         string
	SkipExport     bool
	VMName         string
	OVFToolOptions []string
	OutputDir      string
}

func GetOVFTool() string {
	ovftool := "ovftool"
	if runtime.GOOS == "windows" {
		ovftool = "ovftool.exe"
	}

	if _, err := exec.LookPath(ovftool); err != nil {
		return ""
	}
	return ovftool
}

func (s *StepExport) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Skip export if requested
	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	ui.Say("Skipping export of virtual machine (export is allowed only for ESXi)...")
	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
