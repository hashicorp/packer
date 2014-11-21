// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"fmt"
	"bytes"
	"os"
	"github.com/mitchellh/multistep"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
	"github.com/mitchellh/packer/packer"
	b64 "encoding/base64"
	"path/filepath"
)


const(
	FloppyFileName = "assets.vfd"
)

type StepMountFloppydrive struct {
	FileName string
	Dir string
}

func (s *StepMountFloppydrive) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*iso_config)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error mounting floppy drive: %s"
	vmName := state.Get("vmName").(string)
	packerTempDir :=  state.Get("packerTempDir").(string)

	var err error
	var decBytes []byte

	if config.GuestOSType == WS2012R2DC {
		decBytes, err = b64.StdEncoding.DecodeString(FileAsStringBase64Win2012R2)
//	} else if config.GuestOSType == WS2012DC {
//		decBytes, err = b64.StdEncoding.DecodeString(FileAsStringBase64Win2012)
	}

	var fmtError error

	if err != nil {
		fmtError = fmt.Errorf(errorMsg, err)
		state.Put("error", fmtError)
		ui.Error(fmtError.Error())
		return multistep.ActionHalt
	}

	f, err := os.Create(filepath.Join(packerTempDir,FloppyFileName))
	if err != nil {
		fmtError = fmt.Errorf(errorMsg, err)
		state.Put("error", fmtError)
		ui.Error(fmtError.Error())
		return multistep.ActionHalt
	}

	_, err = f.Write(decBytes)
	if err != nil {
		fmtError = fmt.Errorf(errorMsg, err)
		state.Put("error", fmtError)
		ui.Error(fmtError.Error())
		return multistep.ActionHalt
	}

	s.FileName = f.Name()
	s.Dir = packerTempDir
	defer f.Close()

	ui.Say("Mounting floppy drive...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-VMFloppyDiskDrive -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Path '")
	blockBuffer.WriteString(s.FileName)
	blockBuffer.WriteString("'}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		fmtError = fmt.Errorf(errorMsg, err)
		state.Put("error", fmtError)
		ui.Error(fmtError.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepMountFloppydrive) Cleanup(state multistep.StateBag) {
	if s.FileName == "" {
		return
	}

	errorMsg := "Error unmounting floppy drive: %s"

	vmName := state.Get("vmName").(string)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unmounting floppy drive...")

	var err error = nil

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-VMFloppyDiskDrive -VMName '")
	blockBuffer.WriteString(vmName)
	blockBuffer.WriteString("' -Path $null}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}

	err = os.Remove(s.FileName)

	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}
}
