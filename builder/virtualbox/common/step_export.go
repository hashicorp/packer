package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
	ExportOpts     []string
	Bundling       VBoxBundleConfig
	SkipNatMapping bool
	SkipExport     bool
}

func (s *StepExport) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// If ISO export is configured, ensure this option is propagated to VBoxManage.
	if s.Bundling.BundleISO {
		foundISOOption := false
		
		for _, option := range s.ExportOpts {
			if option == "--iso" || option == "-I" {
				foundISOOption = true
				break
			}
		}
		
		if !foundISOOption {
			s.ExportOpts = append(s.ExportOpts, "--iso")
		}
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Skip export if requested
	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	// Wait a second to ensure VM is really shutdown.
	log.Println("1 second timeout to ensure VM is really shutdown")
	time.Sleep(1 * time.Second)
	ui.Say("Preparing to export machine...")

	// Clear out the Packer-created forwarding rule
	sshPort := state.Get("sshHostPort")
	if !s.SkipNatMapping && sshPort != 0 {
		ui.Message(fmt.Sprintf(
			"Deleting forwarded port mapping for the communicator (SSH, WinRM, etc) (host port %d)", sshPort))
		command := []string{"modifyvm", vmName, "--natpf1", "delete", "packercomm"}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error deleting port forwarding rule: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Export the VM to an OVF
	outputPath := filepath.Join(s.OutputDir, vmName+"."+s.Format)

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
