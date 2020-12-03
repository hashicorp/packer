package proxmoxiso

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// stepUploadAdditionalISOs uploads all additional ISO files that are mountet
// to the VM
type stepUploadAdditionalISOs struct{}

var _ uploader = &proxmox.Client{}

func (s *stepUploadAdditionalISOs) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("proxmoxClient").(uploader)
	c := state.Get("iso-config").(*Config)

	for idx := range c.AdditionalISOFiles {
		if !c.AdditionalISOFiles[idx].ShouldUploadISO {
			state.Put("additional_iso_files", c.AdditionalISOFiles)
			continue
		}

		p := state.Get(c.AdditionalISOFiles[idx].DownloadPathKey).(string)
		if p == "" {
			err := fmt.Errorf("Path to downloaded ISO was empty")
			state.Put("erroe", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		isoPath, _ := filepath.EvalSymlinks(p)
		r, err := os.Open(isoPath)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		filename := filepath.Base(c.AdditionalISOFiles[idx].ISOConfig.ISOUrls[0])
		err = client.Upload(c.Node, c.AdditionalISOFiles[idx].ISOStoragePool, "iso", filename, r)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		isoStoragePath := fmt.Sprintf("%s:iso/%s", c.AdditionalISOFiles[idx].ISOStoragePool, filename)
		c.AdditionalISOFiles[idx].ISOFile = isoStoragePath
		state.Put("additional_iso_files", c.AdditionalISOFiles)
	}
	return multistep.ActionContinue
}

func (s *stepUploadAdditionalISOs) Cleanup(state multistep.StateBag) {
}
