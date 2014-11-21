// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mitchellh/multistep"
	hypervcommon "github.com/mitchellh/packer/builder/hyperv/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"regexp"
	"code.google.com/p/go-uuid/uuid"
	"strings"
)

// EVALUATION EDITIONS
const (
	WS2012R2DC string 	= "WindowsServer2012R2Datacenter"
	PRODUCT_DATACENTER_EVALUATION_SERVER int64 = 80
	PRODUCT_DATACENTER_SERVER int64 = 8
)

// Builder implements packer.Builder and builds the actual Hyperv
// images.
type Builder struct {
	config iso_config
	runner multistep.Runner
}

type iso_config struct {
	DiskSizeGB            uint     			`mapstructure:"disk_size_gb"`
	RamSizeMB             uint     			`mapstructure:"ram_size_mb"`
	GuestOSType         string   			`mapstructure:"guest_os_type"`
	RawSingleISOUrl 	string 				`mapstructure:"iso_url"`
	SleepTimeMinutes 	time.Duration		`mapstructure:"wait_time_minutes"`
	ProductKey 			string				`mapstructure:"product_key"`

	common.PackerConfig           			`mapstructure:",squash"`
	hypervcommon.OutputConfig     			`mapstructure:",squash"`

	VMName              string
	SwitchName          string

tpl *packer.ConfigTemplate
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {

	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("%s: %v", "PackerUserVars", b.config.PackerUserVars))

	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors and warnings
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.OutputConfig.Prepare(b.config.tpl, &b.config.PackerConfig)...)
	warnings := make([]string, 0)

	if b.config.DiskSizeGB == 0 {
		b.config.DiskSizeGB = 40
	}
	log.Println(fmt.Sprintf("%s: %v", "DiskSize", b.config.DiskSizeGB))

	if(b.config.DiskSizeGB < 10 ){
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("disk_size_gb: Windows server requires disk space >= 10 GB, but defined: %v", b.config.DiskSizeGB))
	} else if b.config.DiskSizeGB > 65536 {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("disk_size_gb: Windows server requires disk space <= 65536 GB, but defined: %v", b.config.DiskSizeGB))
	}

	if b.config.RamSizeMB == 0 {
		b.config.RamSizeMB = 1024
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", b.config.RamSizeMB))

	var ramMinMb uint = 512
	var ramMaxMb uint = 6538

	if(b.config.RamSizeMB < ramMinMb ){
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("ram_size_mb: Windows server requires memory size >= %v MB, but defined: %v", ramMinMb, b.config.RamSizeMB))
	} else if b.config.RamSizeMB > ramMaxMb {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("ram_size_mb: Windows server requires memory size <= %v MB, but defined: %v", ramMaxMb, b.config.RamSizeMB))
	}

	warnings = appendWarnings( warnings, fmt.Sprintf("Hyper-V might fail to create a VM if there is no available memory in the system."))


	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("pvm_%s", uuid.New())
	}

	if b.config.SwitchName == "" {
		b.config.SwitchName = fmt.Sprintf("pis_%s", uuid.New())
	}

	if b.config.SleepTimeMinutes == 0 {
		b.config.SleepTimeMinutes = 10
	} else if b.config.SleepTimeMinutes < 0 {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("wait_time_minutes: '%v' %s", int64(b.config.SleepTimeMinutes), "the value can't be negative" ))
	} else if b.config.SleepTimeMinutes > 1440 {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("wait_time_minutes: '%v' %s", uint(b.config.SleepTimeMinutes), "the value is too big" ))
	} else if b.config.SleepTimeMinutes > 120 {
		warnings = appendWarnings( warnings, fmt.Sprintf("wait_time_minutes: '%v' %s", uint(b.config.SleepTimeMinutes), "You may want to decrease the value. Usually 20 min is enough."))
	}
	log.Println(fmt.Sprintf("%s: %v", "SleepTimeMinutes", uint(b.config.SleepTimeMinutes)))


	// Errors
	templates := map[string]*string{
		"iso_url":            &b.config.RawSingleISOUrl,
		"product_key":        &b.config.ProductKey,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	pk := strings.TrimSpace(b.config.ProductKey)
	if len(pk) != 0 {
		pattern := "^[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$"
		value := pk

		match, _ := regexp.MatchString(pattern, value)
		if !match {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("product_key: Make sure the product_key follows the pattern: XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"))
		}

		warnings = appendWarnings( warnings, fmt.Sprintf("product_key: %s", "value is not empty. Packer will try to activate Windows with the product key. To do this Packer will need an Internet connection."))
	}

	log.Println(fmt.Sprintf("%s: %v","VMName", b.config.VMName))
	log.Println(fmt.Sprintf("%s: %v","SwitchName", b.config.SwitchName))
	log.Println(fmt.Sprintf("%s: %v","ProductKey", b.config.ProductKey))



	if b.config.RawSingleISOUrl == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("iso_url: The option can't be missed and a path must be specified."))
	}else if _, err := os.Stat(b.config.RawSingleISOUrl); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("iso_url: Check the path is correct"))
	}

	log.Println(fmt.Sprintf("%s: %v","RawSingleISOUrl", b.config.RawSingleISOUrl))

	guestOSTypesIsValid := false
	guestOSTypes := []string{
		WS2012R2DC,
//		WS2012R2St,
	}

	log.Println(fmt.Sprintf("%s: %v","GuestOSType", b.config.GuestOSType))

	for _, guestType := range guestOSTypes {
		if b.config.GuestOSType == guestType {
			guestOSTypesIsValid = true
			break
		}
	}

	if !guestOSTypesIsValid {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("guest_os_type: The value is invalid. Must be one of: %v", guestOSTypes))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a Hyperv appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Hyperv
	driver, err := hypervcommon.NewHypervPS4Driver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Hyper-V driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

// TODO: comment the next line (debug purpose only)
//	state.Put("vmName", "pvm_cbdde38e-4e2a-4eb6-9718-0f0175d6dd06")

	steps := []multistep.Step{
//		new(hypervcommon.StepAcceptEula),

		new(hypervcommon.StepCreateTempDir),
		&hypervcommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},

		&hypervcommon.StepCreateSwitch{
			SwitchName: b.config.SwitchName,
		},
		new(StepCreateVM),
		new(hypervcommon.StepEnableIntegrationService),
		new(StepMountDvdDrive),
		new(StepMountFloppydrive),
		new(hypervcommon.StepStartVm),
		&hypervcommon.StepSleep{ Minutes: b.config.SleepTimeMinutes, ActionName: "Installing" },

		new(hypervcommon.StepConfigureIp),
		new(hypervcommon.StepSetRemoting),
		new(common.StepProvision),
		new(StepInstallProductKey),

		new(StepExportVm),

//		new(hypervcommon.StepConfigureIp),
//		new(hypervcommon.StepSetRemoting),
//		new(hypervcommon.StepCheckRemoting),
//		new(msbldcommon.StepSysprep),
	}

	// Run the steps.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}
	b.runner.Run(state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return hypervcommon.NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func appendWarnings(slice []string, data ...string) []string {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]string, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

