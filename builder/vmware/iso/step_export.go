package iso

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"path/filepath"
	"strings"
)

type StepExport struct {
	Format string
}

func (s *StepExport) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	if c.RemoteType != "esx5" {
		return multistep.ActionContinue
	}

	if _, err := exec.LookPath("ovftool"); err != nil {
		err := fmt.Errorf("Error ovftool not found: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Export the VM
	outputPath := filepath.Join(c.VMName, c.VMName+"."+s.Format)

	args := []string{
		"--noSSLVerify=true",
		"--skipManifestCheck",
		"-tt=" + s.Format,
		"vi://" + c.RemoteUser + ":" + c.RemotePassword + "@" + c.RemoteHost + "/" + c.VMName,
		outputPath,
	}

	ui.Say("Exporting virtual machine...")
	ui.Message(fmt.Sprintf("Executing: ovftool %s", strings.Join(args, " ")))
	var out bytes.Buffer
	cmd := exec.Command("ovftool", args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error exporting virtual machine: %s", err, out.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("%s", out.String()))

	state.Put("exportPath", outputPath)

	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
