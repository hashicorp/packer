package iso

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"path/filepath"
)

type StepRemoteUpload struct {
	Datastore string
	Host      string
}

func (s *StepRemoteUpload) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	if path, ok := state.GetOk("iso_path"); ok {
		filename := filepath.Base(path.(string))

		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			state.Put("error", fmt.Errorf("datastore doesn't exist: %v", err))
			return multistep.ActionHalt
		}

		remotePath := fmt.Sprintf("packer_cache/%s", filename)
		remoteDirectory := fmt.Sprintf("[%s] packer_cache/", ds.Name())
		fullRemotePath := fmt.Sprintf("%s/%s", remoteDirectory, filename)

		ui.Say(fmt.Sprintf("Uploading %s to %s", filename, remotePath))

		if exists := ds.FileExists(remotePath); exists == true {
			ui.Say("File already upload")
			state.Put("iso_remote_path", fullRemotePath)
			return multistep.ActionContinue
		}

		if err := ds.MakeDirectory(remoteDirectory); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		if err := ds.UploadFile(path.(string), remotePath, s.Host); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("iso_remote_path", fullRemotePath)
	}

	return multistep.ActionContinue
}

func (s *StepRemoteUpload) Cleanup(state multistep.StateBag) {}
