package iso

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"fmt"
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

	uploadedFloppyPath string
}

func (s *StepAddFloppy) Run(state multistep.StateBag) multistep.StepAction {
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
	if s.Config.FloppyIMGPath != "" && tmpFloppy != nil {
		return fmt.Errorf("'floppy_img_path' cannot be used together with 'floppy_files' and 'floppy_dirs'")
	}

	var floppyIMGPath string
	if tmpFloppy != nil {
		ui.Say("Uploading created floppy image")

		ds, err := d.FindDatastore(s.Datastore)
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

		// remember the path to the temporary floppy image to remove it after the build is finished
		s.uploadedFloppyPath = uploadPath
		floppyIMGPath = ds.ResolvePath(uploadPath)
	} else {
		floppyIMGPath = s.Config.FloppyIMGPath
	}

	ui.Say("Adding Floppy...")
	err := vm.AddFloppy(floppyIMGPath)
	if err != nil {
		return err
	}

	return nil
}

func (s *StepAddFloppy) Cleanup(state multistep.StateBag) {
	if s.uploadedFloppyPath == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	ds, err := d.FindDatastore(s.Datastore)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	if err := ds.Delete(s.uploadedFloppyPath); err != nil {
		ui.Error(fmt.Sprintf("Error deleting floppy image '%v': %v", s.uploadedFloppyPath, err.Error()))
		return
	}

	// the img will be automatically ejected
}
