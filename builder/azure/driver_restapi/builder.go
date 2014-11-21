// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver_restapi

import (
	"errors"
	"fmt"
	"log"
	"time"
	"os"
	"regexp"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/builder/azure/utils"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/driver"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/targets"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/targets/lin"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/targets/win"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
)

// Builder implements packer.Builder and builds the actual Azure
// images.
type Builder struct {
	config azure_config
	runner multistep.Runner
}

type azure_config struct {
	SubscriptionName        string     	`mapstructure:"subscription_name"`
	PublishSettingsPath     string     	`mapstructure:"publish_settings_path"`
	StorageAccount          string     	`mapstructure:"storage_account"`
	StorageAccountContainer	string     	`mapstructure:"storage_account_container"`
	OsType         			string   	`mapstructure:"os_type"`
	OsImageLabel         	string   	`mapstructure:"os_image_label"`
	Location 				string 		`mapstructure:"location"`
	InstanceSize 			string		`mapstructure:"instance_size"`
	UserImageLabel 			string		`mapstructure:"user_image_label"`
	common.PackerConfig           		`mapstructure:",squash"`
	tpl *packer.ConfigTemplate

	username          		string		`mapstructure:"username"`
	tmpVmName              	string
	tmpServiceName          string
	tmpContainerName        string
	userImageName          	string
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
	warnings := make([]string, 0)


	templates := map[string]*string{
		"subscription_name"			: &b.config.SubscriptionName,
		"publish_settings_path"		: &b.config.PublishSettingsPath,
		"storage_account"			: &b.config.StorageAccount,
		"storage_account_container"	: &b.config.StorageAccountContainer,
		"os_type"					: &b.config.OsType,
		"os_image_label"			: &b.config.OsImageLabel,
		"location"					: &b.config.Location,
		"instance_size"				: &b.config.InstanceSize,
		"user_image_label"			: &b.config.UserImageLabel,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}


	if b.config.SubscriptionName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("subscription_name: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","subscription_name", b.config.SubscriptionName))

	if b.config.PublishSettingsPath == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("publish_settings_path: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","publish_settings_path", b.config.PublishSettingsPath))

	if b.config.StorageAccount == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("storage_account: The option can't be missed."))
	}

	if _, err := os.Stat(b.config.PublishSettingsPath); err != nil {
		errs = packer.MultiErrorAppend(errs, errors.New("publish_settings_path: Check the path is correct."))
	}
	log.Println(fmt.Sprintf("%s: %v","storage_account", b.config.StorageAccount))

	if b.config.StorageAccountContainer == "" {
		b.config.StorageAccountContainer = "vhds"
	}
	log.Println(fmt.Sprintf("%s: %v","storage_account", b.config.StorageAccountContainer))

	osTypeIsValid := false
	osTypeArr := []string{
		targets.Linux,
		targets.Windows,
	}

	log.Println(fmt.Sprintf("%s: %v","os_type", b.config.OsType))

	for _, osType := range osTypeArr {
		if b.config.OsType == osType {
			osTypeIsValid = true
			break
		}
	}

	if !osTypeIsValid {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("os_type: The value is invalid. Must be one of: %v", osTypeArr))
	}

	if b.config.OsImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("os_image_label: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","os_image_label", b.config.OsImageLabel))

	if b.config.Location == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("location: The option can't be missed."))
	}
	log.Println(fmt.Sprintf("%s: %v","location", b.config.Location))

	sizeIsValid := false
	instanceSizeArr := []string{
		targets.ExtraSmall,
		targets.Small,
		targets.Medium,
		targets.Large,
		targets.ExtraLarge,
		targets.A5,
		targets.A6,
		targets.A7,
		targets.A8,
		targets.A9,
	}

	log.Println(fmt.Sprintf("%s: %v","instance_size", b.config.InstanceSize))

	for _, instanceSize := range instanceSizeArr {
		if b.config.InstanceSize == instanceSize {
			sizeIsValid = true
			break
		}
	}

	if !sizeIsValid {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("instance_size: The value is invalid. Must be one of: %v", instanceSizeArr))
	}

	if b.config.UserImageLabel == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("user_image_label: The option can't be missed."))
	}

	// check
	pattern := "^[A-Za-z][A-Za-z0-9-_.]*[A-Za-z0-9]$"
	value :=  b.config.UserImageLabel

	match, _ := regexp.MatchString(pattern, value)
	if !match {
		errs = packer.MultiErrorAppend(errs, errors.New("user_image_label: '"+ value +"'. The first and last characters should be letter or digit the others should follow the pattern [A-Za-z0-9-_.]*."))
	}

	//	log.Println(fmt.Sprintf("%s: %v","user_image_label", b.config.UserImageLabel))

	b.config.userImageName = utils.DecorateImageName(b.config.UserImageLabel)
	log.Println(fmt.Sprintf("%s: %v","user_image_name", b.config.userImageName))

	b.config.tmpContainerName = utils.BuildContainerName()

	log.Println(fmt.Sprintf("%s: %v","user_image_label", b.config.UserImageLabel))

	// for Win  - the computer name cannot be more than 15 characters long
	const tmpServiceNamePrefix = "PkrSrv"
	const tmpVmNamePrefix = "PkrVM"
	randSuffix := utils.BuildAzureVmNameRandomSuffix(tmpVmNamePrefix)

	if b.config.tmpVmName == "" {
		b.config.tmpVmName = fmt.Sprintf("%s%s", tmpVmNamePrefix, randSuffix)
	}
	log.Println(fmt.Sprintf("%s: %v","tmpVmName", b.config.tmpVmName))

	if b.config.tmpServiceName == "" {
		b.config.tmpServiceName = fmt.Sprintf("%s%s", tmpServiceNamePrefix, randSuffix)
	}
	log.Println(fmt.Sprintf("%s: %v","tmpServiceName", b.config.tmpServiceName))

	if b.config.username == "" {
		b.config.username = fmt.Sprintf("%s", "packer")
	}
	log.Println(fmt.Sprintf("%s: %v","username", b.config.username))

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a PS Azure appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {

	var err error
	ui.Say("Preparing builder...")

	ui.Message("Getting subscription info...")
	var subscriptionInfo *SubscriptionInfo
	subscriptionInfo, err = ParsePublishSettings(b.config.PublishSettingsPath, b.config.SubscriptionName)
	if err != nil {
		return nil, fmt.Errorf("ParsePublishSettings error: %s\n", err.Error())
	}

	ui.Message("Creating rest api driver...")
	var driverRest driver.IDriverRest
	driverRest, err = driver.NewTlsDriver(subscriptionInfo.CertData)
	if err != nil {
		return nil, fmt.Errorf("Failed creating rest api driver: %s", err)
	}

	ui.Message("Creating request manager...")
	reqManager := &request.Manager{
			SubscrId: subscriptionInfo.Id,
			Driver : driverRest,
		}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put(constants.Config, &b.config)
	state.Put(constants.RequestManager, reqManager)
	state.Put("hook", hook)
	state.Put(constants.Ui, ui)

	// complete flags
	state.Put(constants.CertCreated, 0)
	state.Put(constants.SrvExists, 0)
	state.Put(constants.CertInstalled, 0)
	state.Put(constants.CertUploaded, 0)
	state.Put(constants.VmExists, 0)
	state.Put(constants.DiskExists, 0)
	state.Put(constants.VmRunning, 0)
	state.Put(constants.ImageCreated, 0)

	ui.Say("Validating Azure options...")
	err = b.validateAzureOptions(ui, state, reqManager)
	if err != nil {
		return nil, fmt.Errorf("Some Azure options failed: %s", err)
	}

	var steps []multistep.Step

	if b.config.OsType == targets.Linux {
		certFileName:= "cert.pem"
		keyFileName := "key.pem"

		steps = []multistep.Step{
			&lin.StepCreateCert {
				CertFileName: certFileName,
				KeyFileName: keyFileName,
				TmpServiceName: b.config.tmpServiceName,
				},
			&targets.StepCreateService {
				Location: b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
				},
			&targets.StepUploadCertificate {
				TmpServiceName: b.config.tmpServiceName,
				},
			&lin.StepCreateVm {
				OsType: b.config.OsType,
				StorageAccount: b.config.StorageAccount,
				StorageAccountContainer: b.config.StorageAccountContainer,
				OsImageLabel: b.config.OsImageLabel,
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				InstanceSize: b.config.InstanceSize,
				Username: b.config.username,
				},

			&targets.StepPollStatus {
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName: b.config.tmpVmName,
				OsType: b.config.OsType,
			},

			&common.StepConnectSSH {
				SSHAddress:     lin.SSHAddress,
				SSHConfig:      lin.SSHConfig(b.config.username),
				SSHWaitTimeout: 20*time.Minute,
				},
			&common.StepProvision {},

			&lin.StepGeneralizeOs{
				Command: "sudo /usr/sbin/waagent -force -deprovision+user && export HISTSIZE=0 && sync",
				},
			&targets.StepStopVm {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				},
			&targets.StepCreateImage {
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName: b.config.tmpVmName,
				UserImageName: b.config.userImageName,
				UserImageLabel: b.config.UserImageLabel,
				RecommendedVMSize: b.config.InstanceSize,
			},
		}
	} else if b.config.OsType == targets.Windows {
//		b.config.tmpVmName = "shchTemp"
//		b.config.tmpServiceName = "shchTemp"
		steps = []multistep.Step {

			&targets.StepCreateService {
				Location: b.config.Location,
				TmpServiceName: b.config.tmpServiceName,
				},
			&win.StepCreateVm {
				OsType: b.config.OsType,
				StorageAccount: b.config.StorageAccount,
				StorageAccountContainer: b.config.StorageAccountContainer,
				OsImageLabel: b.config.OsImageLabel,
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
				InstanceSize: b.config.InstanceSize,
				Username: b.config.username,
				Password: "Zxcv1234",
				},
			&targets.StepPollStatus {
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName: b.config.tmpVmName,
				OsType: b.config.OsType,
			},
			&win.StepSetProvisionInfrastructure {
				VmName: b.config.tmpVmName,
				ServiceName: b.config.tmpServiceName,
				StorageAccountName: b.config.StorageAccount,
				TempContainerName: b.config.tmpContainerName,
				},
			&common.StepProvision{},
			&targets.StepStopVm {
				TmpVmName: b.config.tmpVmName,
				TmpServiceName: b.config.tmpServiceName,
			},
			&targets.StepCreateImage {
				TmpServiceName: b.config.tmpServiceName,
				TmpVmName: b.config.tmpVmName,
				UserImageName: b.config.userImageName,
				UserImageLabel: b.config.UserImageLabel,
				RecommendedVMSize: b.config.InstanceSize,
			},
		}

	} else {
		return nil, fmt.Errorf("Unkonwn OS type: %s", b.config.OsType)
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

	requestData := reqManager.GetVmImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		log.Printf("reqManager.GetVmImages returned error: %s", err.Error())
		return nil, fmt.Errorf("Can't create artifact")
	}

	vmImageList, err := response.ParseVmImageList(resp.Body)

	if err != nil {
		log.Printf("response.ParseVmImageList returned error: %s", err.Error())
		return nil, fmt.Errorf("Can't create artifact")
	}

	userImage :=  vmImageList.First(b.config.userImageName)
	if userImage == nil {
		log.Printf("vmImageList.First returned nil")
		return nil, fmt.Errorf("Can't create artifact")
	}

	return &artifact {
		imageLabel: userImage.Label,
		imageName: userImage.Name,
		mediaLocation: userImage.OSDiskConfiguration.MediaLink,
		}, nil
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder)validateAzureOptions(ui packer.Ui, state *multistep.BasicStateBag, reqManager *request.Manager) error {

	var err error

	// Check Storage account (& container)
	ui.Message("Checking storage account...")

	requestData := reqManager.CheckStorageAccountNameAvailability(b.config.StorageAccount)
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return err
	}

	availabilityResponse, err := response.ParseAvailabilityResponse(resp.Body)

	log.Printf("availabilityResponse:\n %v", availabilityResponse)

	if availabilityResponse.Result == "true" {
		return fmt.Errorf("Can't find storage account '%s'", b.config.StorageAccount)
	}

	// Check image exists
	exists, err := b.checkOsImageExists(ui, state, reqManager)
	if err != nil {
		return err
	}

	if( exists == false) {
		exists, err = b.checkOsUserImageExists(ui, state, reqManager)
		if err != nil {
			return err
		}

		if( exists == false) {
			err = fmt.Errorf("Can't find OS image '%s' located at '%s'", b.config.OsImageLabel, b.config.Location)
			return err
		}
	}

	return nil
}

func  (b *Builder) checkOsImageExists(ui packer.Ui, state *multistep.BasicStateBag, reqManager *request.Manager) (bool, error) {
	ui.Message("Checking OS image with the label '"+ b.config.OsImageLabel +"' exists...")
	requestData := reqManager.GetOsImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return false, err
	}

	imageList, err := response.ParseOsImageList(resp.Body)

	if err != nil {
		return false, err
	}

	filteredImageList := imageList.Filter(b.config.OsImageLabel, b.config.Location)

	if len(filteredImageList) != 0 {

		ui.Message(fmt.Sprintf("Found %v image(s).", len(filteredImageList)))
		ui.Message("Take the most recent:")

		imageList.SortByDateDesc(filteredImageList)

		osImageName := filteredImageList[0].Name
		ui.Message("OS image label: " + filteredImageList[0].Label)
		ui.Message("OS image family: " + filteredImageList[0].ImageFamily)
		ui.Message("OS image name: " + osImageName)
		ui.Message("OS image published date: " + filteredImageList[0].PublishedDate)
		state.Put(constants.OsImageName, osImageName)
		state.Put(constants.IsOSImage, true)
		return true, nil
	}

	ui.Message("Image not found.")
	return false, nil
}

func (b *Builder) checkOsUserImageExists(ui packer.Ui, state *multistep.BasicStateBag, reqManager *request.Manager) (bool, error) {
	// check user images
	ui.Message("Checking VM image with the label '"+ b.config.OsImageLabel +"' exists...")

	requestData := reqManager.GetVmImages()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return false, err
	}

	imageList, err := response.ParseVmImageList(resp.Body)

	if err != nil {
		return false, err
	}

	filteredImageList := imageList.Filter(b.config.OsImageLabel, b.config.Location)

	if len(filteredImageList) != 0 {

		ui.Message(fmt.Sprintf("Found %v image(s).", len(filteredImageList)))
		ui.Message("Take the most recent:")

		imageList.SortByDateDesc(filteredImageList)

		osImageName := filteredImageList[0].Name
		ui.Message("VM image label: " + filteredImageList[0].Label)
		ui.Message("VM image family: " + filteredImageList[0].ImageFamily)
		ui.Message("VM image name: " + osImageName)
		ui.Message("VM image published date: " + filteredImageList[0].PublishedDate)
		state.Put(constants.OsImageName, osImageName)
		state.Put(constants.IsOSImage, false)
		return true, nil
	}

	ui.Message("Image not found.")
	return false, nil
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

