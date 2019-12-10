package iso

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type FloppyConfig struct {
	FloppyIMGPath     string   `mapstructure:"floppy_img_path"`
	FloppyFiles       []string `mapstructure:"floppy_files"`
	FloppyDirectories []string `mapstructure:"floppy_dirs"`
}

type StepAddFloppy struct {
	Config    *FloppyConfig
	Datastore string
	Host      string
}

func (s *StepAddFloppy) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)
	d := state.Get("driver").(*driver.Driver)

	if floppyPath, ok := state.GetOk("floppy_path"); ok {
		ui.Say("Uploading created floppy image")

		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		vmDir, err := vm.GetDir()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		uploadPath := fmt.Sprintf("%v/packer-tmp-created-floppy.flp", vmDir)
		if err := ds.UploadFile(floppyPath.(string), uploadPath, s.Host); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("uploaded_floppy_path", uploadPath)

		ui.Say("Adding generated Floppy...")
		floppyIMGPath := ds.ResolvePath(uploadPath)
		err = vm.AddFloppy(floppyIMGPath)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	if s.Config.FloppyIMGPath != "" {
		ui.Say("Adding Floppy image...")
		err := vm.AddFloppy(s.Config.FloppyIMGPath)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepAddFloppy) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	if UploadedFloppyPath, ok := state.GetOk("uploaded_floppy_path"); ok {
		ui.Say("Deleting Floppy image ...")

		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			state.Put("error", err)
			return
		}

		err = ds.Delete(UploadedFloppyPath.(string))
		if err != nil {
			state.Put("error", err)
			return
		}

	}
}
