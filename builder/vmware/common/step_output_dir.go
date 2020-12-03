package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepOutputDir sets up the output directory by creating it if it does
// not exist, deleting it if it does exist and we're forcing, and cleaning
// it up when we're done with it.
type StepOutputDir struct {
	Force bool

	OutputConfig *OutputConfig
	VMName       string

	RemoteType string

	success bool
}

func (s *StepOutputDir) SetOutputAndExportDirs(state multistep.StateBag) OutputDir {
	driver := state.Get("driver")

	// Hold on to your pants. The output configuration is a little more complex
	// than you'd expect because of all the moving parts between local and
	// remote output, and exports, and legacy behavior.
	var dir OutputDir
	switch d := driver.(type) {
	case OutputDir:
		// The driver fulfils the OutputDir interface so that it can create
		// output files on the remote instance.
		dir = d
	default:
		// The driver will be running the build and creating the output
		// directory locally
		dir = new(LocalOutputDir)
	}

	// If remote type is esx, we need to track both the output dir on the remote
	// instance and the output dir locally. exportOutputPath is where we track
	// the local output dir.
	exportOutputPath := s.OutputConfig.OutputDir

	if s.RemoteType != "" {
		if s.OutputConfig.RemoteOutputDir != "" {
			// User set the remote output dir.
			s.OutputConfig.OutputDir = s.OutputConfig.RemoteOutputDir
		} else {
			// Default output dir to vm name. On remote esx instance, this will
			// become something like /vmfs/volumes/mydatastore/vmname/vmname.vmx
			s.OutputConfig.OutputDir = s.VMName
		}
	}
	// Remember, this one's either the output from a local build, or the remote
	// output from a remote build. Not the local export path for a remote build.
	dir.SetOutputDir(s.OutputConfig.OutputDir)

	// Set dir in the state for use in file cleanup and artifact
	state.Put("dir", dir)
	state.Put("export_output_path", exportOutputPath)
	return dir
}

func (s *StepOutputDir) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Configuring output and export directories...")

	dir := s.SetOutputAndExportDirs(state)
	exists, err := dir.DirExists()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if exists {
		if s.Force {
			ui.Message("Deleting previous output directory...")
			dir.RemoveAll()
		} else {
			state.Put("error", fmt.Errorf(
				"Output directory '%s' already exists.", dir.String()))
			return multistep.ActionHalt
		}
	}

	if err := dir.MkdirAll(); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.success = true
	return multistep.ActionContinue
}

func (s *StepOutputDir) Cleanup(state multistep.StateBag) {
	if !s.success {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if cancelled || halted {
		dir := state.Get("dir").(OutputDir)
		ui := state.Get("ui").(packersdk.Ui)

		exists, _ := dir.DirExists()
		if exists {
			ui.Say("Deleting output directory...")
			for i := 0; i < 5; i++ {
				err := dir.RemoveAll()
				if err == nil {
					break
				}

				log.Printf("Error removing output dir: %s", err)
				time.Sleep(2 * time.Second)
			}
		}
	}
}
