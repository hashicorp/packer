package wim

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

const (
	installWIM     = `install.wim`
	installWIMPath = `sources\` + installWIM
)

type StepExtractWIM struct {
	TempPath   string
	ISOPathKey string
	ResultKey  string

	// Set only when an ISO has been mounted. It's unset when the ISO is dismounted
	isoPath string
}

func (s *StepExtractWIM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	isoPath := state.Get(s.ISOPathKey).(string)

	ui.Say("Extracting WIM...")

	// Mount ISO
	devicePath, err := s.mountDiskImage(isoPath)
	if err != nil {
		err = fmt.Errorf("Error mounting ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Copy WIM out
	srcWimPath := fmt.Sprintf("%s\\%s", devicePath, installWIMPath)

	_, err = os.Stat(srcWimPath)
	if os.IsNotExist(err) {
		err = fmt.Errorf("Error copying ISO: %s", srcWimPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	srcWim, err := os.Open(srcWimPath)
	if err != nil {
		err = fmt.Errorf("Error copying WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer srcWim.Close()

	dstWimPath := fmt.Sprintf("%s/%s", s.TempPath, installWIM)

	dstWim, err := os.Create(dstWimPath)
	if err != nil {
		err = fmt.Errorf("Error copying WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer dstWim.Close()

	_, err = io.Copy(dstWim, srcWim)
	if err != nil {
		err = fmt.Errorf("Error copying WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(s.ResultKey, dstWimPath)

	// Dismount ISO

	return multistep.ActionContinue
}

func (s *StepExtractWIM) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)

	dstWimPath := path.Join(s.TempPath, installWIM)
	_ = os.Remove(dstWimPath)

	err := s.dismountDiskImage()
	if err != nil {
		err = fmt.Errorf("Error mounting ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
	}
}

func (s *StepExtractWIM) dismountDiskImage() error {
	if s.isoPath == "" {
		return nil
	}

	var script = `
param([string]$imagePath)
Dismount-DiskImage -ImagePath $imagePath
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, s.isoPath)
	return err
}

func (s *StepExtractWIM) mountDiskImage(isoPath string) (string, error) {

	var script = `
param([string]$imagePath)
$diskImage = Mount-DiskImage -ImagePath $imagePath
if ($diskImage -ne $null) {
    $diskImage.DevicePath
}
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, isoPath)

	if err != nil {
		return "", err
	}

	var devicePath = strings.TrimSpace(cmdOut)
	return devicePath, err
}
