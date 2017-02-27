package iso

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepExport struct {
	Format     string
	SkipExport bool
}

func (s *StepExport) generateArgs(c *Config, outputPath string, hidePassword bool) []string {
	password := url.QueryEscape(c.RemotePassword)
	if hidePassword {
		password = "****"
	}
	args := []string{
		"--noSSLVerify=true",
		"--skipManifestCheck",
		"-tt=" + s.Format,
		"vi://" + c.RemoteUser + ":" + password + "@" + c.RemoteHost + "/" + c.VMName,
		outputPath,
	}
	return append(c.OVFToolOptions, args...)
}

func (s *StepExport) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	// Skip export if requested
	if c.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	if c.RemoteType != "esx5" || s.Format == "" {
		return multistep.ActionContinue
	}

	ovftool := "ovftool"
	if runtime.GOOS == "windows" {
		ovftool = "ovftool.exe"
	}

	if _, err := exec.LookPath(ovftool); err != nil {
		err := fmt.Errorf("Error %s not found: %s", ovftool, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Export the VM
	outputPath := filepath.Join(c.VMName, c.VMName+"."+s.Format)

	if s.Format == "ova" {
		os.MkdirAll(outputPath, 0755)
	}

	ui.Say("Exporting virtual machine...")
	ui.Message(fmt.Sprintf("Executing: %s %s", ovftool, strings.Join(s.generateArgs(c, outputPath, true), " ")))
	var out bytes.Buffer
	cmd := exec.Command(ovftool, s.generateArgs(c, outputPath, false)...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error exporting virtual machine: %s\n%s\n", err, out.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("%s", out.String()))

	state.Put("exportPath", outputPath)

	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
