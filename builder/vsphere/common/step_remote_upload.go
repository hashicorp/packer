package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRemoteUpload struct {
	Datastore                  string
	Host                       string
	SetHostForDatastoreUploads bool
	UploadedCustomCD           bool
}

func (s *StepRemoteUpload) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
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
		s.UploadedCustomCD = true
		state.Put("cd_path", fullRemotePath)
	}

	return multistep.ActionContinue
}

func GetRemoteDirectoryAndPath(path string, ds driver.Datastore) (string, string, string, string) {
	filename := filepath.Base(path)
	remotePath := fmt.Sprintf("packer_cache/%s", filename)
	remoteDirectory := fmt.Sprintf("[%s] packer_cache/", ds.Name())
	fullRemotePath := fmt.Sprintf("%s/%s", remoteDirectory, filename)

	return filename, remotePath, remoteDirectory, fullRemotePath

}
func (s *StepRemoteUpload) uploadFile(path string, d driver.Driver, ui packersdk.Ui) (string, error) {
	ds, err := d.FindDatastore(s.Datastore, s.Host)
	if err != nil {
		return "", fmt.Errorf("datastore doesn't exist: %v", err)
	}

	filename, remotePath, remoteDirectory, fullRemotePath := GetRemoteDirectoryAndPath(path, ds)

	if exists := ds.FileExists(remotePath); exists == true {
		ui.Say(fmt.Sprintf("File %s already exists; skipping upload.", fullRemotePath))
		return fullRemotePath, nil
	}

	ui.Say(fmt.Sprintf("Uploading %s to %s", filename, remotePath))

	if exists := ds.DirExists(remotePath); exists == false {
		log.Printf("Remote directory doesn't exist; creating...")
		if err := ds.MakeDirectory(remoteDirectory); err != nil {
			return "", err
		}
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

	if !s.UploadedCustomCD {
		return
	}

	UploadedCDPath, ok := state.GetOk("cd_path")
	if !ok {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	d := state.Get("driver").(*driver.VCenterDriver)
	ui.Say("Deleting cd_files image from remote datastore ...")

	ds, err := d.FindDatastore(s.Datastore, s.Host)
	if err != nil {
		log.Printf("Error finding datastore to delete custom CD; please delete manually: %s", err)
		return
	}

	err = ds.Delete(UploadedCDPath.(string))
	if err != nil {
		log.Printf("Error deleting custom CD from remote datastore; please delete manually: %s", err)
		return

	}
}
