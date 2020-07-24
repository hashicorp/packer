package common

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

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

func (s *StepExport) generateArgs(c *DriverConfig, displayName string, hidePassword bool) ([]string, error) {

	ovftool_uri := fmt.Sprintf("vi://%s/%s", c.RemoteHost, displayName)
	u, err := url.Parse(ovftool_uri)
	if err != nil {
		return []string{}, err
	}

	password := c.RemotePassword
	if hidePassword {
		password = "<password>"
	}
	u.User = url.UserPassword(c.RemoteUser, password)

	args := []string{
		"--noSSLVerify=true",
		"--skipManifestCheck",
		"-tt=" + s.Format,
		u.String(),
		s.OutputDir,
	}
	return append(s.OVFToolOptions, args...), nil
}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("driverConfig").(*DriverConfig)
	ui := state.Get("ui").(packer.Ui)

	// Skip export if requested
	if s.SkipExport {
		ui.Say("Skipping export of virtual machine...")
		return multistep.ActionContinue
	}

	if c.RemoteType != "esx5" {
		ui.Say("Skipping export of virtual machine (export is allowed only for ESXi)...")
		return multistep.ActionContinue
	}

	ovftool := GetOVFTool()
	if ovftool == "" {
		err := fmt.Errorf("Error ovftool not found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Export the VM
	if s.OutputDir == "" {
		s.OutputDir = s.VMName + "." + s.Format
	}

	os.MkdirAll(s.OutputDir, 0755)

	ui.Say("Exporting virtual machine...")
	var displayName string
	if v, ok := state.GetOk("display_name"); ok {
		displayName = v.(string)
	}
	ui_args, err := s.generateArgs(c, displayName, true)
	if err != nil {
		err := fmt.Errorf("Couldn't generate ovftool uri: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Message(fmt.Sprintf("Executing: %s %s", ovftool, strings.Join(ui_args, " ")))
	var out bytes.Buffer
	args, err := s.generateArgs(c, displayName, false)
	if err != nil {
		err := fmt.Errorf("Couldn't generate ovftool uri: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	cmd := exec.Command(ovftool, args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error exporting virtual machine: %s\n%s\n", err, out.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message(out.String())

	return multistep.ActionContinue
}

func (s *StepExport) Cleanup(state multistep.StateBag) {}
