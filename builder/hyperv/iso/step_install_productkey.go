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
	"strings"
	"log"
	"strconv"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
)

type StepInstallProductKey struct {
}

func (s *StepInstallProductKey) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*iso_config)
//	driver := state.Get("driver").(hypervcommon.Driver)

	pk := config.ProductKey

	if len(pk) == 0{
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	guestOsType := config.GuestOSType


		comm := state.Get("communicator").(packer.Communicator)
	//	vmName := state.Get("vmName").(string)


		var err error
	//	var fmtError error
		var stderrString string
		var stdoutString string
		errorMsg := "Error installing product key: %s"

		ui.Say("Installing product key...")

	// get windows edition

		var remoteCmd packer.RemoteCmd
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		var blockBuffer bytes.Buffer

		blockBuffer.WriteString("{")
		blockBuffer.WriteString("(Get-WmiObject -class Win32_OperatingSystem).OperatingSystemSKU")
		blockBuffer.WriteString("}")

		remoteCmd.Command = "-ScriptBlock " + blockBuffer.String()
		remoteCmd.Stdout = stdout
		remoteCmd.Stderr = stderr

		err = comm.Start(&remoteCmd)

		stderrString = strings.TrimSpace(stderr.String())
		stdoutString = strings.TrimSpace(stdout.String())

		log.Printf("stdout: %s", stdoutString)
		log.Printf("stderr: %s", stderrString)

		if len(stderrString) > 0 {
			err = fmt.Errorf(errorMsg, stderrString)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		sku, err := strconv.ParseInt(stdoutString, 10, 32)

		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	// installing product key

	var steps []multistep.Step

	if guestOsType == WS2012R2DC {

		if sku == PRODUCT_DATACENTER_EVALUATION_SERVER {
			ui.Say("Product type: Server Datacenter (evaluation installation)")

		// turn eval edition into full

			ui.Say("Turnig evaluation edition into full...")

			blockBuffer.Reset()
			blockBuffer.WriteString("{")
			blockBuffer.WriteString(" Start-Process DISM -NoNewWindow -Wait -Argument '/online /Set-Edition:ServerDatacenter /ProductKey:")
			blockBuffer.WriteString(pk)
			blockBuffer.WriteString("  /AcceptEula /NoRestart'")
			blockBuffer.WriteString("}")

			remoteCmd.Command = "-ScriptBlock " + blockBuffer.String()
			stderr.Reset()
			stdout.Reset()

			err = comm.Start(&remoteCmd)

			stderrString := strings.TrimSpace(stderr.String())
			stdoutString := strings.TrimSpace(stdout.String())

			log.Printf("stdout: %s", stdoutString)
			log.Printf("stderr: %s", stderrString)

			if len(stderrString) > 0 {
				err = fmt.Errorf(errorMsg, stderrString)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			steps = []multistep.Step{
				new(hypervcommon.StepCreateExternalSwitch),
				new(hypervcommon.StepDisableVlan),
				new(hypervcommon.StepRebootVm),
				new(hypervcommon.StepConfigureIp),
				new(hypervcommon.StepSetRemoting),
				new(hypervcommon.StepCheckRemoting),
				new(hypervcommon.StepExecuteOnlineActivation),
			}

		} else if sku == PRODUCT_DATACENTER_SERVER {
			ui.Say("Product type: Server Datacenter (full installation)")

			steps = []multistep.Step{
				new(hypervcommon.StepCreateExternalSwitch),
				new(hypervcommon.StepDisableVlan),
				new(hypervcommon.StepRebootVm),
				new(hypervcommon.StepConfigureIp),
				new(hypervcommon.StepSetRemoting),
				new(hypervcommon.StepCheckRemoting),
				&hypervcommon.StepExecuteOnlineActivationFull{Pk:pk},
			}

		} else {
			err := fmt.Errorf(errorMsg, "Unsupported product type (SKU)")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Create external switch and connect to the VM
		ui.Say("Packer needs an internet connection to activate Windows. Packer will try to create an external switch connected to the Internet and execute an activation script. In case of error, you'll have to do it manually.")


		runner := multistep.BasicRunner{Steps: steps}
		runner.Run(state)
	}

	return multistep.ActionContinue
}

func (s *StepInstallProductKey) Cleanup(state multistep.StateBag) {
}
