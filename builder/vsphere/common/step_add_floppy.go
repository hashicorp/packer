//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type FloppyConfig

package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type FloppyConfig struct {
	// Datastore path to a floppy image that will be mounted to the VM.
	// Example: `[datastore1] ISO/pvscsi-Windows8.flp`.
	FloppyIMGPath string `mapstructure:"floppy_img_path"`
	// List of local files to be mounted to the VM floppy drive. Can be used to
	// make Debian preseed or RHEL kickstart files available to the VM.
	FloppyFiles []string `mapstructure:"floppy_files"`
	// List of directories to copy files from.
	FloppyDirectories []string `mapstructure:"floppy_dirs"`
	// The label to use for the floppy disk that
	// is attached when the VM is booted. This is most useful for cloud-init,
	// Kickstart or other early initialization tools, which can benefit from labelled floppy disks.
	// By default, the floppy label will be 'packer'.
	FloppyLabel string `mapstructure:"floppy_label"`
}

type StepAddFloppy struct {
	Config                     *FloppyConfig
	Datastore                  string
	Host                       string
	SetHostForDatastoreUploads bool
}

func (s *StepAddFloppy) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)
	d := state.Get("driver").(driver.Driver)

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
		if err := ds.UploadFile(floppyPath.(string), uploadPath, s.Host, s.SetHostForDatastoreUploads); err != nil {
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

	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(driver.Driver)

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
