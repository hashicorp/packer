package iso

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepRemoteUpload struct {
	Datastore                  string
	Host                       string
	SetHostForDatastoreUploads bool
}

func (s *StepRemoteUpload) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(driver.Driver)

	if path, ok := state.GetOk("iso_path"); ok {
		// user-supplied boot iso
		fullRemotePath, err := s.uploadFile(path.(string), d, ui)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("iso_remote_path", fullRemotePath)
	}
	if cdPath, ok := state.GetOk("cd_path"); ok {
		// Packer-created cd_files disk
		fullRemotePath, err := s.uploadFile(cdPath.(string), d, ui)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("cd_path", fullRemotePath)
	}

	return multistep.ActionContinue
}

func (s *StepRemoteUpload) uploadFile(path string, d driver.Driver, ui packer.Ui) (string, error) {
	ds, err := d.FindDatastore(s.Datastore, s.Host)
	if err != nil {
		return "", fmt.Errorf("datastore doesn't exist: %v", err)
	}

	filename := filepath.Base(path)
	remotePath := fmt.Sprintf("packer_cache/%s", filename)
	remoteDirectory := fmt.Sprintf("[%s] packer_cache/", ds.Name())
	fullRemotePath := fmt.Sprintf("%s/%s", remoteDirectory, filename)

	ui.Say(fmt.Sprintf("Uploading %s to %s", filename, remotePath))

	if exists := ds.FileExists(remotePath); exists == true {
		ui.Say(fmt.Sprintf("File %s already uploaded; continuing", filename))
		return fullRemotePath, nil
	}

	if err := ds.MakeDirectory(remoteDirectory); err != nil {
		return "", err
	}

	if err := ds.UploadFile(path, remotePath, s.Host, s.SetHostForDatastoreUploads); err != nil {
		return "", err
	}
	return fullRemotePath, nil
}

func (s *StepRemoteUpload) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.VCenterDriver)

	if UploadedCDPath, ok := state.GetOk("cd_path"); ok {
		ui.Say("Deleting cd_files image from remote datastore ...")

		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			state.Put("error", err)
			return
		}

		err = ds.Delete(UploadedCDPath.(string))
		if err != nil {
			state.Put("error", err)
			return
		}

	}
}
