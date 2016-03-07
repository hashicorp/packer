// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in the project root for license information.

package arm

import (
	"errors"
	"fmt"
	"log"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/builder/azure/common/lin"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
)

type Builder struct {
	config   *Config
	stateBag multistep.StateBag
	runner   multistep.Runner
}

const (
	DefaultPublicIPAddressName = "packerPublicIP"
)

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := newConfig(raws...)
	if errs != nil {
		return warnings, errs
	}

	b.config = c

	b.stateBag = new(multistep.BasicStateBag)
	err := b.configureStateBag(b.stateBag)
	if err != nil {
		return nil, err
	}

	return warnings, errs
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	ui.Say("Preparing builder ...")

	b.stateBag.Put("hook", hook)
	b.stateBag.Put(constants.Ui, ui)

	servicePrincipalToken, err := b.createServicePrincipalToken()
	if err != nil {
		return nil, err
	}

	ui.Message("Creating Azure Resource Manager (ARM) client ...")
	azureClient, err := NewAzureClient(b.config.SubscriptionID, b.config.ResourceGroupName, b.config.StorageAccount, servicePrincipalToken)
	if err != nil {
		return nil, err
	}

	steps := []multistep.Step{
		NewStepCreateResourceGroup(azureClient, ui),
		NewStepValidateTemplate(azureClient, ui),
		NewStepDeployTemplate(azureClient, ui),
		NewStepGetIPAddress(azureClient, ui),
		&communicator.StepConnectSSH{
			Config:    &b.config.Comm,
			Host:      lin.SSHHost,
			SSHConfig: lin.SSHConfig(b.config.UserName),
		},
		&common.StepProvision{},
		NewStepGetOSDisk(azureClient, ui),
		NewStepPowerOffCompute(azureClient, ui),
		NewStepCaptureImage(azureClient, ui),
		NewStepDeleteResourceGroup(azureClient, ui),
		NewStepDeleteOSDisk(azureClient, ui),
	}

	if b.config.PackerDebug {
		ui.Message(fmt.Sprintf("temp admin user: '%s'", b.config.UserName))
		ui.Message(fmt.Sprintf("temp admin password: '%s'", b.config.Password))
	}

	b.runner = b.createRunner(&steps, ui)
	b.runner.Run(b.stateBag)

	// Report any errors.
	if rawErr, ok := b.stateBag.GetOk(constants.Error); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := b.stateBag.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := b.stateBag.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return &artifact{}, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) createRunner(steps *[]multistep.Step, ui packer.Ui) multistep.Runner {
	if b.config.PackerDebug {
		return &multistep.DebugRunner{
			Steps:   *steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	}

	return &multistep.BasicRunner{
		Steps: *steps,
	}
}

func (b *Builder) configureStateBag(stateBag multistep.StateBag) error {
	stateBag.Put(constants.AuthorizedKey, b.config.sshAuthorizedKey)
	stateBag.Put(constants.PrivateKey, b.config.sshPrivateKey)

	stateBag.Put(constants.ArmComputeName, b.config.tmpComputeName)
	stateBag.Put(constants.ArmDeploymentName, b.config.tmpDeploymentName)
	stateBag.Put(constants.ArmLocation, b.config.Location)
	stateBag.Put(constants.ArmResourceGroupName, b.config.tmpResourceGroupName)
	stateBag.Put(constants.ArmTemplateParameters, b.config.toTemplateParameters())
	stateBag.Put(constants.ArmVirtualMachineCaptureParameters, b.config.toVirtualMachineCaptureParameters())

	stateBag.Put(constants.ArmPublicIPAddressName, DefaultPublicIPAddressName)

	return nil
}

func (b *Builder) createServicePrincipalToken() (*azure.ServicePrincipalToken, error) {
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(b.config.TenantID)
	if err != nil {
		return nil, err
	}

	spt, err := azure.NewServicePrincipalToken(
		*oauthConfig,
		b.config.ClientID,
		b.config.ClientSecret,
		azure.PublicCloud.ResourceManagerEndpoint)

	return spt, err
}
