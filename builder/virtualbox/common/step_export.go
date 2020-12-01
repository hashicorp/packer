package common

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step cleans up forwarded ports and exports the VM to an OVF.
//
// Uses:
//
// Produces:
//   exportPath string - The path to the resulting export.
type StepExport struct {
	Format         string
	OutputDir      string
	OutputFilename string
	ExportOpts     []string
	Bundling       VBoxBundleConfig
	SkipNatMapping bool
	SkipExport     bool
}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// If ISO export is configured, ensure this option is propagated to VBoxManage.
	for _, option := range s.ExportOpts {
		if option == "--iso" || option == "-I" {
			s.ExportOpts = append(s.ExportOpts, "--iso")
			break
		}
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)
	if s.OutputFilename == "" {
		s.OutputFilename = vmName
	}

	// Skip export if requested
	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	ui.Say("Preparing to export machine...")

	// Clear out the Packer-created forwarding rule
	commPort := state.Get("commHostPort")
	if !s.SkipNatMapping && commPort != 0 {
		ui.Message(fmt.Sprintf(
			"Deleting forwarded port mapping for the communicator (SSH, WinRM, etc) (host port %d)", commPort))
		command := []string{"modifyvm", vmName, "--natpf1", "delete", "packercomm"}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error deleting port forwarding rule: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Export the VM to an OVF
	outputPath := filepath.Join(s.OutputDir, s.OutputFilename+"."+s.Format)

	command := []string{
		"export",
		vmName,
		"--output",
		outputPath,
	}
	command = append(command, s.ExportOpts...)

	ui.Say("Exporting virtual machine...")
	ui.Message(fmt.Sprintf("Executing: %s", strings.Join(command, " ")))
	err := driver.VBoxManage(command...)
	if err != nil {
		err := fmt.Errorf("Error exporting virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("exportPath", outputPath)

	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
