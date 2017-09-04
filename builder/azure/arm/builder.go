// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	packerAzureCommon "github.com/hashicorp/packer/builder/azure/common"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/builder/azure/common/lin"

	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest/adal"
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type Builder struct {
	config   *Config
	stateBag multistep.StateBag
	runner   multistep.Runner
}

const (
	DefaultNicName             = "packerNic"
	DefaultPublicIPAddressName = "packerPublicIP"
	DefaultSasBlobContainer    = "system/Microsoft.Compute"
	DefaultSasBlobPermission   = "r"
	DefaultSecretName          = "packerKeyVaultSecret"
)

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := newConfig(raws...)
	if errs != nil {
		return warnings, errs
	}

	b.config = c

	b.stateBag = new(multistep.BasicStateBag)
	b.configureStateBag(b.stateBag)
	b.setTemplateParameters(b.stateBag)
	b.setImageParameters(b.stateBag)

	return warnings, errs
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	ui.Say("Running builder ...")

	if err := newConfigRetriever().FillParameters(b.config); err != nil {
		return nil, err
	}

	log.Print(":: Configuration")
	packerAzureCommon.DumpConfig(b.config, func(s string) { log.Print(s) })

	b.stateBag.Put("hook", hook)
	b.stateBag.Put(constants.Ui, ui)

	spnCloud, spnKeyVault, err := b.getServicePrincipalTokens(ui.Say)
	if err != nil {
		return nil, err
	}

	ui.Message("Creating Azure Resource Manager (ARM) client ...")
	azureClient, err := NewAzureClient(
		b.config.SubscriptionID,
		b.config.ResourceGroupName,
		b.config.StorageAccount,
		b.config.cloudEnvironment,
		spnCloud,
		spnKeyVault)

	if err != nil {
		return nil, err
	}

	resolver := newResourceResolver(azureClient)
	if err := resolver.Resolve(b.config); err != nil {
		return nil, err
	}

	if b.config.isManagedImage() {
		group, err := azureClient.GroupsClient.Get(b.config.ManagedImageResourceGroupName)
		if err != nil {
			return nil, fmt.Errorf("Cannot locate the managed image resource group %s.", b.config.ManagedImageResourceGroupName)
		}

		b.config.manageImageLocation = *group.Location

		// If a managed image already exists it cannot be overwritten.
		_, err = azureClient.ImagesClient.Get(b.config.ManagedImageResourceGroupName, b.config.ManagedImageName, "")
		if err == nil {
			return nil, fmt.Errorf("A managed image named %s already exists in the resource group %s.", b.config.ManagedImageName, b.config.ManagedImageResourceGroupName)
		}
	}

	if b.config.StorageAccount != "" {
		account, err := b.getBlobAccount(azureClient, b.config.ResourceGroupName, b.config.StorageAccount)
		if err != nil {
			return nil, err
		}
		b.config.storageAccountBlobEndpoint = *account.AccountProperties.PrimaryEndpoints.Blob

		if !equalLocation(*account.Location, b.config.Location) {
			return nil, fmt.Errorf("The storage account is located in %s, but the build will take place in %s. The locations must be identical", *account.Location, b.config.Location)
		}
	}

	endpointConnectType := PublicEndpoint
	if b.isPublicPrivateNetworkCommunication() && b.isPrivateNetworkCommunication() {
		endpointConnectType = PublicEndpointInPrivateNetwork
	} else if b.isPrivateNetworkCommunication() {
		endpointConnectType = PrivateEndpoint
	}

	b.setRuntimeParameters(b.stateBag)
	b.setTemplateParameters(b.stateBag)
	b.setImageParameters(b.stateBag)
	var steps []multistep.Step

	if b.config.OSType == constants.Target_Linux {
		steps = []multistep.Step{
			NewStepCreateResourceGroup(azureClient, ui),
			NewStepValidateTemplate(azureClient, ui, b.config, GetVirtualMachineDeployment),
			NewStepDeployTemplate(azureClient, ui, b.config, GetVirtualMachineDeployment),
			NewStepGetIPAddress(azureClient, ui, endpointConnectType),
			&communicator.StepConnectSSH{
				Config:    &b.config.Comm,
				Host:      lin.SSHHost,
				SSHConfig: lin.SSHConfig(b.config.UserName),
			},
			&packerCommon.StepProvision{},
			NewStepGetOSDisk(azureClient, ui),
			NewStepPowerOffCompute(azureClient, ui),
			NewStepCaptureImage(azureClient, ui),
			NewStepDeleteResourceGroup(azureClient, ui),
			NewStepDeleteOSDisk(azureClient, ui),
		}
	} else if b.config.OSType == constants.Target_Windows {
		steps = []multistep.Step{
			NewStepCreateResourceGroup(azureClient, ui),
			NewStepValidateTemplate(azureClient, ui, b.config, GetKeyVaultDeployment),
			NewStepDeployTemplate(azureClient, ui, b.config, GetKeyVaultDeployment),
			NewStepGetCertificate(azureClient, ui),
			NewStepSetCertificate(b.config, ui),
			NewStepValidateTemplate(azureClient, ui, b.config, GetVirtualMachineDeployment),
			NewStepDeployTemplate(azureClient, ui, b.config, GetVirtualMachineDeployment),
			NewStepGetIPAddress(azureClient, ui, endpointConnectType),
			&communicator.StepConnectWinRM{
				Config: &b.config.Comm,
				Host: func(stateBag multistep.StateBag) (string, error) {
					return stateBag.Get(constants.SSHHost).(string), nil
				},
				WinRMConfig: func(multistep.StateBag) (*communicator.WinRMConfig, error) {
					return &communicator.WinRMConfig{
						Username: b.config.UserName,
						Password: b.config.tmpAdminPassword,
					}, nil
				},
			},
			&packerCommon.StepProvision{},
			NewStepGetOSDisk(azureClient, ui),
			NewStepPowerOffCompute(azureClient, ui),
			NewStepCaptureImage(azureClient, ui),
			NewStepDeleteResourceGroup(azureClient, ui),
			NewStepDeleteOSDisk(azureClient, ui),
		}
	} else {
		return nil, fmt.Errorf("Builder does not support the os_type '%s'", b.config.OSType)
	}

	if b.config.PackerDebug {
		ui.Message(fmt.Sprintf("temp admin user: '%s'", b.config.UserName))
		ui.Message(fmt.Sprintf("temp admin password: '%s'", b.config.Password))

		if b.config.sshPrivateKey != "" {
			debugKeyPath := fmt.Sprintf("%s-%s.pem", b.config.PackerBuildName, b.config.tmpComputeName)
			ui.Message(fmt.Sprintf("temp ssh key: %s", debugKeyPath))

			b.writeSSHPrivateKey(ui, debugKeyPath)
		}
	}

	b.runner = packerCommon.NewRunner(steps, b.config.PackerConfig, ui)
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

	if b.config.isManagedImage() {
		return NewManagedImageArtifact(b.config.ManagedImageResourceGroupName, b.config.ManagedImageName, b.config.manageImageLocation)
	} else if template, ok := b.stateBag.GetOk(constants.ArmCaptureTemplate); ok {
		return NewArtifact(
			template.(*CaptureTemplate),
			func(name string) string {
				month := time.Now().AddDate(0, 1, 0).UTC()
				blob := azureClient.BlobStorageClient.GetContainerReference(DefaultSasBlobContainer).GetBlobReference(name)
				sasUrl, _ := blob.GetSASURI(month, DefaultSasBlobPermission)
				return sasUrl
			})
	}

	return &Artifact{}, nil
}

func (b *Builder) writeSSHPrivateKey(ui packer.Ui, debugKeyPath string) {
	f, err := os.Create(debugKeyPath)
	if err != nil {
		ui.Say(fmt.Sprintf("Error saving debug key: %s", err))
	}
	defer f.Close()

	// Write the key out
	if _, err := f.Write([]byte(b.config.sshPrivateKey)); err != nil {
		ui.Say(fmt.Sprintf("Error saving debug key: %s", err))
		return
	}

	// Chmod it so that it is SSH ready
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0600); err != nil {
			ui.Say(fmt.Sprintf("Error setting permissions of debug key: %s", err))
		}
	}
}

func (b *Builder) isPublicPrivateNetworkCommunication() bool {
	return DefaultPrivateVirtualNetworkWithPublicIp != b.config.PrivateVirtualNetworkWithPublicIp
}

func (b *Builder) isPrivateNetworkCommunication() bool {
	return b.config.VirtualNetworkName != ""
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func equalLocation(location1, location2 string) bool {
	return strings.EqualFold(canonicalizeLocation(location1), canonicalizeLocation(location2))
}

func canonicalizeLocation(location string) string {
	return strings.Replace(location, " ", "", -1)
}

func (b *Builder) getBlobAccount(client *AzureClient, resourceGroupName string, storageAccountName string) (*storage.Account, error) {
	account, err := client.AccountsClient.GetProperties(resourceGroupName, storageAccountName)
	if err != nil {
		return nil, err
	}

	return &account, err
}

func (b *Builder) configureStateBag(stateBag multistep.StateBag) {
	stateBag.Put(constants.AuthorizedKey, b.config.sshAuthorizedKey)
	stateBag.Put(constants.PrivateKey, b.config.sshPrivateKey)

	stateBag.Put(constants.ArmTags, &b.config.AzureTags)
	stateBag.Put(constants.ArmComputeName, b.config.tmpComputeName)
	stateBag.Put(constants.ArmDeploymentName, b.config.tmpDeploymentName)
	stateBag.Put(constants.ArmKeyVaultName, b.config.tmpKeyVaultName)
	stateBag.Put(constants.ArmLocation, b.config.Location)
	stateBag.Put(constants.ArmNicName, DefaultNicName)
	stateBag.Put(constants.ArmPublicIPAddressName, DefaultPublicIPAddressName)
	stateBag.Put(constants.ArmResourceGroupName, b.config.tmpResourceGroupName)
	stateBag.Put(constants.ArmStorageAccountName, b.config.StorageAccount)

	stateBag.Put(constants.ArmIsManagedImage, b.config.isManagedImage())
	stateBag.Put(constants.ArmManagedImageResourceGroupName, b.config.ManagedImageResourceGroupName)
	stateBag.Put(constants.ArmManagedImageName, b.config.ManagedImageName)
}

// Parameters that are only known at runtime after querying Azure.
func (b *Builder) setRuntimeParameters(stateBag multistep.StateBag) {
	stateBag.Put(constants.ArmManagedImageLocation, b.config.manageImageLocation)
}

func (b *Builder) setTemplateParameters(stateBag multistep.StateBag) {
	stateBag.Put(constants.ArmVirtualMachineCaptureParameters, b.config.toVirtualMachineCaptureParameters())
}

func (b *Builder) setImageParameters(stateBag multistep.StateBag) {
	stateBag.Put(constants.ArmImageParameters, b.config.toImageParameters())
}

func (b *Builder) getServicePrincipalTokens(say func(string)) (*adal.ServicePrincipalToken, *adal.ServicePrincipalToken, error) {
	var servicePrincipalToken *adal.ServicePrincipalToken
	var servicePrincipalTokenVault *adal.ServicePrincipalToken

	var err error

	if b.config.useDeviceLogin {
		servicePrincipalToken, err = packerAzureCommon.Authenticate(*b.config.cloudEnvironment, b.config.TenantID, say)
		if err != nil {
			return nil, nil, err
		}
	} else {
		auth := NewAuthenticate(*b.config.cloudEnvironment, b.config.ClientID, b.config.ClientSecret, b.config.TenantID)

		servicePrincipalToken, err = auth.getServicePrincipalToken()
		if err != nil {
			return nil, nil, err
		}

		servicePrincipalTokenVault, err = auth.getServicePrincipalTokenWithResource(
			strings.TrimRight(b.config.cloudEnvironment.KeyVaultEndpoint, "/"))

		if err != nil {
			return nil, nil, err
		}
	}

	return servicePrincipalToken, servicePrincipalTokenVault, nil
}
