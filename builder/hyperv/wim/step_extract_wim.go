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
	res, err := s.mountDiskImage(isoPath)
	if err != nil {
		err = fmt.Errorf("Error mounting ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if res["DevicePath"] == "" {
		err = fmt.Errorf("Error reading DevicePath")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Copy WIM out
	srcWimPath := fmt.Sprintf("%s\\%s", res["DevicePath"], installWIMPath)

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

	_, err := s.dismountDiskImage()
	if err != nil {
		err = fmt.Errorf("Error mounting ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
	}
}

func (s *StepExtractWIM) dismountDiskImage() (map[string]string, error) {
	if s.isoPath != "" {
		cmd := fmt.Sprintf("Dismount-DiskImage -ImagePath %s", s.isoPath)
		res, err := s.execPSCmd(cmd)
		if err != nil {
			s.isoPath = ""
		}
		return res, err
	}
	return nil, nil
}

// TODO: dedup
func (s *StepExtractWIM) execPSCmd(cmd string) (map[string]string, error) {
	var ps powershell.PowerShellCmd

	cmdOut, err := ps.Output(cmd)
	if err == nil {
		res := make(map[string]string)
		source := strings.Split(cmdOut, "\n")
		for _, line := range source {
			kv := strings.SplitN(line, ":", 2)
			if len(kv) > 1 {
				res[strings.Trim(kv[0], " ")] = strings.Trim(kv[1], " \n\r")
			}
		}
		return res, err
	} else {
		return nil, err
	}
}

func (s *StepExtractWIM) mountDiskImage(isoPath string) (map[string]string, error) {
	cmd := fmt.Sprintf("Mount-DiskImage -ImagePath %s", isoPath)
	res, err := s.execPSCmd(cmd)
	if err == nil {
		s.isoPath = isoPath
	}
	return res, err
}
