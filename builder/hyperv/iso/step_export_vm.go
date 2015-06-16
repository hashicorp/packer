// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
	"io/ioutil"
)

const(
	vhdDir string = "Virtual Hard Disks"
	vmDir string = "Virtual Machines"
)

type StepExportVm struct {
}

func (s *StepExportVm) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*iso_config)
	driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	var err error
	var errorMsg string

	vmName := state.Get("vmName").(string)
	tmpPath :=	state.Get("packerTempDir").(string)
	outputPath := config.OutputDir

	// create temp path to export vm
	errorMsg = "Error creating temp export path: %s"
	vmExportPath , err := ioutil.TempDir(tmpPath, "export")
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Exporting vm...")
	errorMsg = "Error exporting vm: %s"

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$vmName='" + vmName + "';")
	blockBuffer.WriteString("$path='" + vmExportPath + "';")
	blockBuffer.WriteString("Export-VM -Name $vmName -Path $path")
	blockBuffer.WriteString("}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// copy to output dir

	ui.Say("Coping to output dir...")

	errorMsg = "Error copying vm: %s"

	expPath := filepath.Join(vmExportPath,vmName)

	blockBuffer.Reset()
	blockBuffer.WriteString("Invoke-Command -scriptblock {")
	blockBuffer.WriteString("$srcPath='" + expPath + "';")
	blockBuffer.WriteString("$dstPath='" + outputPath + "';")
	blockBuffer.WriteString("$vhdDirName='" + vhdDir + "';")
	blockBuffer.WriteString("$vmDir='" + vmDir + "';")
	blockBuffer.WriteString("cpi \"$srcPath\\$vhdDirName\"  $dstPath -recurse;")
	blockBuffer.WriteString("cpi \"$srcPath\\$vmDir\"  \"$dstPath\";")
	blockBuffer.WriteString("cpi \"$srcPath\\$vmDir\\*.xml\"  \"$dstPath\\$vmDir\";")
	blockBuffer.WriteString("}")

	err = driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepExportVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
