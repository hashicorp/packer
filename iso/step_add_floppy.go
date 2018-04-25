package iso

import (
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"fmt"
	"context"
)

type FloppyConfig struct {
	FloppyIMGPath     string   `mapstructure:"floppy_img_path"`
	FloppyFiles       []string `mapstructure:"floppy_files"`
	FloppyDirectories []string `mapstructure:"floppy_dirs"`
}

func (c *FloppyConfig) Prepare() []error {
	var errs []error

	if c.FloppyIMGPath != "" && (c.FloppyFiles != nil || c.FloppyDirectories != nil) {
		errs = append(errs,
			fmt.Errorf("'floppy_img_path' cannot be used together with 'floppy_files' and 'floppy_dirs'"),
		)
	}

	return errs
}

type StepAddFloppy struct {
	Config    *FloppyConfig
	Datastore string
	Host      string
}

func (s *StepAddFloppy) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	err := s.runImpl(state)
	if err != nil {
		state.Put("error", fmt.Errorf("error adding floppy: %v", err))
		return multistep.ActionHalt
	}
	return multistep.ActionContinue
}

func (s *StepAddFloppy) runImpl(state multistep.StateBag) error {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)
	d := state.Get("driver").(*driver.Driver)

	tmpFloppy := state.Get("floppy_path")
	if tmpFloppy != nil {
		ui.Say("Uploading created floppy image")

		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			return err
		}
		vmDir, err := vm.GetDir()
		if err != nil {
			return err
		}

		uploadPath := fmt.Sprintf("%v/packer-tmp-created-floppy.img", vmDir)
		if err := ds.UploadFile(tmpFloppy.(string), uploadPath); err != nil {
			return fmt.Errorf("error uploading floppy image: %v", err)
		}
		state.Put("uploaded_floppy_path", uploadPath)

		floppyIMGPath := ds.ResolvePath(uploadPath)
		ui.Say("Adding generated Floppy...")
		err = vm.AddFloppy(floppyIMGPath)
		if err != nil {
			return err
		}
	}

	if s.Config.FloppyIMGPath != "" {
		floppyIMGPath := s.Config.FloppyIMGPath
		ui.Say("Adding Floppy image...")
		err := vm.AddFloppy(floppyIMGPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StepAddFloppy) Cleanup(state multistep.StateBag) {}
