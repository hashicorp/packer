package common

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// Defining this interface ensures that we use the common step download, or the
// mock created to test this wrapper
type DownloadStep interface {
	Run(context.Context, multistep.StateBag) multistep.StepAction
	Cleanup(multistep.StateBag)
	UseSourceToFindCacheTarget(source string) (*url.URL, string, error)
}

// VSphere has a specialized need -- before we waste time downloading an iso,
// we need to check whether that iso already exists on the remote datastore.
// if it does, we skip the download. This wrapping-step still uses the common
// StepDownload but only if the image isn't already present on the datastore.
type StepDownload struct {
	DownloadStep DownloadStep
	// These keys are VSphere-specific and used to check the remote datastore.
	Url       []string
	ResultKey string
	Datastore string
	Host      string
}

func (s *StepDownload) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(driver.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	// Check whether iso is present on remote datastore.
	ds, err := driver.FindDatastore(s.Datastore, s.Host)
	if err != nil {
		state.Put("error", fmt.Errorf("datastore doesn't exist: %v", err))
		return multistep.ActionHalt
	}

	// loop over URLs to see if any are already present. If they are, store that
	// one instate and continue
	for _, source := range s.Url {
		_, targetPath, err := s.DownloadStep.UseSourceToFindCacheTarget(source)
		if err != nil {
			state.Put("error", fmt.Errorf("Error getting target path: %s", err))
			return multistep.ActionHalt
		}
		_, remotePath, _, _ := GetRemoteDirectoryAndPath(targetPath, ds)

		if exists := ds.FileExists(remotePath); exists {
			ui.Say(fmt.Sprintf("File %s already uploaded; continuing", targetPath))
			state.Put(s.ResultKey, targetPath)
			return multistep.ActionContinue
		}
	}

	// ISO is not present on datastore, so we need to download, then upload it.
	// Pass through to the common download step.
	return s.DownloadStep.Run(ctx, state)
}

func (s *StepDownload) Cleanup(state multistep.StateBag) {
}
