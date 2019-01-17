package common

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step upload the VMX to the remote host
//
// Uses:
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepUploadVMX struct {
}

func (c *StepUploadVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	return multistep.ActionContinue
}

func (StepUploadVMX) Cleanup(multistep.StateBag) {}
