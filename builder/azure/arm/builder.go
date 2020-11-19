package arm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	armstorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-10-01/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/hcl/v2/hcldec"
	packerAzureCommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/builder/azure/common/lin"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type Builder struct {
	config   Config
	stateBag multistep.StateBag
	runner   multistep.Runner
}

const (
	DefaultSasBlobContainer = "system/Microsoft.Compute"
	DefaultSecretName       = "packerKeyVaultSecret"
)

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	b.stateBag = new(multistep.BasicStateBag)
	b.configureStateBag(b.stateBag)
	b.setTemplateParameters(b.stateBag)
	b.setImageParameters(b.stateBag)

	return nil, warnings, errs
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {

	ui.Say("Running builder ...")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// FillParameters function captures authType and sets defaults.
	err := b.config.ClientConfig.FillParameters()
	if err != nil {
		return nil, err
	}

	//When running Packer on an Azure instance using Managed Identity, FillParameters will update SubscriptionID from the instance
	// so lets make sure to update our state bag with the valid subscriptionID.
	if b.config.isManagedImage() && b.config.SharedGalleryDestination.SigDestinationGalleryName != "" {
		b.stateBag.Put(constants.ArmManagedImageSubscription, b.config.ClientConfig.SubscriptionID)
	}

	log.Print(":: Configuration")
	packerAzureCommon.DumpConfig(&b.config, func(s string) { log.Print(s) })

	b.stateBag.Put("hook", hook)
	b.stateBag.Put(constants.Ui, ui)

	spnCloud, spnKeyVault, err := b.getServicePrincipalTokens(ui.Say)
	if err != nil {
		return nil, err
	}

	ui.Message("Creating Azure Resource Manager (ARM) client ...")
	azureClient, err := NewAzureClient(
		b.config.ClientConfig.SubscriptionID,
		b.config.SharedGalleryDestination.SigDestinationSubscription,
		b.config.ResourceGroupName,
		b.config.StorageAccount,
		b.config.ClientConfig.CloudEnvironment(),
		b.config.SharedGalleryTimeout,
		b.config.PollingDurationTimeout,
		spnCloud,
		spnKeyVault)

	if err != nil {
		return nil, err
	}

	resolver := newResourceResolver(azureClient)
	if err := resolver.Resolve(&b.config); err != nil {
		return nil, err
	}
	if b.config.ClientConfig.ObjectID == "" {
		b.config.ClientConfig.ObjectID = getObjectIdFromToken(ui, spnCloud)
	} else {
		ui.Message("You have provided Object_ID which is no longer needed, azure packer builder determines this dynamically from the authentication token")
	}

	if b.config.ClientConfig.ObjectID == "" && b.config.OSType != constants.Target_Linux {
		return nil, fmt.Errorf("could not determine the ObjectID for the user, which is required for Windows builds")
	}

	if b.config.isManagedImage() {
		_, err := azureClient.GroupsClient.Get(ctx, b.config.ManagedImageResourceGroupName)
		if err != nil {
			return nil, fmt.Errorf("Cannot locate the managed image resource group %s.", b.config.ManagedImageResourceGroupName)
		}

		// If a managed image already exists it cannot be overwritten.
		_, err = azureClient.ImagesClient.Get(ctx, b.config.ManagedImageResourceGroupName, b.config.ManagedImageName, "")
		if err == nil {
			if b.config.PackerForce {
				ui.Say(fmt.Sprintf("the managed image named %s already exists, but deleting it due to -force flag", b.config.ManagedImageName))
				f, err := azureClient.ImagesClient.Delete(ctx, b.config.ManagedImageResourceGroupName, b.config.ManagedImageName)
				if err == nil {
					err = f.WaitForCompletionRef(ctx, azureClient.ImagesClient.Client)
				}
				if err != nil {
					return nil, fmt.Errorf("failed to delete the managed image named %s : %s", b.config.ManagedImageName, azureClient.LastError.Error())
				}
			} else {
				return nil, fmt.Errorf("the managed image named %s already exists in the resource group %s, use the -force option to automatically delete it.", b.config.ManagedImageName, b.config.ManagedImageResourceGroupName)
			}
		}
	} else {
		// User is not using Managed Images to build, warning message here that this path is being deprecated
		ui.Error("Warning: You are using Azure Packer Builder to create VHDs which is being deprecated, consider using Managed Images. Learn more https://www.packer.io/docs/builders/azure/arm#azure-arm-builder-specific-options")
	}

	if b.config.BuildResourceGroupName != "" {
		group, err := azureClient.GroupsClient.Get(ctx, b.config.BuildResourceGroupName)
		if err != nil {
			return nil, fmt.Errorf("Cannot locate the existing build resource resource group %s.", b.config.BuildResourceGroupName)
		}

		b.config.Location = *group.Location
	}

	b.config.validateLocationZoneResiliency(ui.Say)

	if b.config.StorageAccount != "" {
		account, err := b.getBlobAccount(ctx, azureClient, b.config.ResourceGroupName, b.config.StorageAccount)
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

	deploymentName := b.stateBag.Get(constants.ArmDeploymentName).(string)

	// For Managed Images, validate that Shared Gallery Image exists before publishing to SIG
	if b.config.isManagedImage() && b.config.SharedGalleryDestination.SigDestinationGalleryName != "" {
		_, err = azureClient.GalleryImagesClient.Get(ctx, b.config.SharedGalleryDestination.SigDestinationResourceGroup, b.config.SharedGalleryDestination.SigDestinationGalleryName, b.config.SharedGalleryDestination.SigDestinationImageName)
		if err != nil {
			return nil, fmt.Errorf("the Shared Gallery Image to which to publish the managed image version to does not exist in the resource group %s", b.config.SharedGalleryDestination.SigDestinationResourceGroup)
		}
		// SIG requires that replication regions include the region in which the Managed Image resides
		managedImageLocation := normalizeAzureRegion(b.stateBag.Get(constants.ArmLocation).(string))
		foundMandatoryReplicationRegion := false
		var normalizedReplicationRegions []string
		for _, region := range b.config.SharedGalleryDestination.SigDestinationReplicationRegions {
			// change region to lower-case and strip spaces
			normalizedRegion := normalizeAzureRegion(region)
			normalizedReplicationRegions = append(normalizedReplicationRegions, normalizedRegion)
			if strings.EqualFold(normalizedRegion, managedImageLocation) {
				foundMandatoryReplicationRegion = true
				continue
			}
		}
		if foundMandatoryReplicationRegion == false {
			b.config.SharedGalleryDestination.SigDestinationReplicationRegions = append(normalizedReplicationRegions, managedImageLocation)
		}
		b.stateBag.Put(constants.ArmManagedImageSharedGalleryReplicationRegions, b.config.SharedGalleryDestination.SigDestinationReplicationRegions)
	}

	var steps []multistep.Step
	if b.config.OSType == constants.Target_Linux {
		steps = []multistep.Step{
			NewStepCreateResourceGroup(azureClient, ui),
			NewStepValidateTemplate(azureClient, ui, &b.config, GetVirtualMachineDeployment),
			NewStepDeployTemplate(azureClient, ui, &b.config, deploymentName, GetVirtualMachineDeployment),
			NewStepGetIPAddress(azureClient, ui, endpointConnectType),
			&communicator.StepConnectSSH{
				Config:    &b.config.Comm,
				Host:      lin.SSHHost,
				SSHConfig: b.config.Comm.SSHConfigFunc(),
			},
			&commonsteps.StepProvision{},
			&commonsteps.StepCleanupTempKeys{
				Comm: &b.config.Comm,
			},
			NewStepGetOSDisk(azureClient, ui),
			NewStepGetAdditionalDisks(azureClient, ui),
			NewStepPowerOffCompute(azureClient, ui),
			NewStepSnapshotOSDisk(azureClient, ui, &b.config),
			NewStepSnapshotDataDisks(azureClient, ui, &b.config),
			NewStepCaptureImage(azureClient, ui),
			NewStepPublishToSharedImageGallery(azureClient, ui, &b.config),
		}
	} else if b.config.OSType == constants.Target_Windows {
		steps = []multistep.Step{
			NewStepCreateResourceGroup(azureClient, ui),
		}
		if b.config.BuildKeyVaultName == "" {
			keyVaultDeploymentName := b.stateBag.Get(constants.ArmKeyVaultDeploymentName).(string)
			steps = append(steps,
				NewStepValidateTemplate(azureClient, ui, &b.config, GetKeyVaultDeployment),
				NewStepDeployTemplate(azureClient, ui, &b.config, keyVaultDeploymentName, GetKeyVaultDeployment),
			)
		} else {
			steps = append(steps, NewStepCertificateInKeyVault(&azureClient.VaultClient, ui, &b.config))
		}
		steps = append(steps,
			NewStepGetCertificate(azureClient, ui),
			NewStepSetCertificate(&b.config, ui),
			NewStepValidateTemplate(azureClient, ui, &b.config, GetVirtualMachineDeployment),
			NewStepDeployTemplate(azureClient, ui, &b.config, deploymentName, GetVirtualMachineDeployment),
			NewStepGetIPAddress(azureClient, ui, endpointConnectType),
			&communicator.StepConnectWinRM{
				Config: &b.config.Comm,
				Host: func(stateBag multistep.StateBag) (string, error) {
					return stateBag.Get(constants.SSHHost).(string), nil
				},
				WinRMConfig: func(multistep.StateBag) (*communicator.WinRMConfig, error) {
					return &communicator.WinRMConfig{
						Username: b.config.UserName,
						Password: b.config.Password,
					}, nil
				},
			},
			&commonsteps.StepProvision{},
			NewStepGetOSDisk(azureClient, ui),
			NewStepGetAdditionalDisks(azureClient, ui),
			NewStepPowerOffCompute(azureClient, ui),
			NewStepSnapshotOSDisk(azureClient, ui, &b.config),
			NewStepSnapshotDataDisks(azureClient, ui, &b.config),
			NewStepCaptureImage(azureClient, ui),
			NewStepPublishToSharedImageGallery(azureClient, ui, &b.config),
		)
	} else {
		return nil, fmt.Errorf("Builder does not support the os_type '%s'", b.config.OSType)
	}

	if b.config.PackerDebug {
		ui.Message(fmt.Sprintf("temp admin user: '%s'", b.config.UserName))
		ui.Message(fmt.Sprintf("temp admin password: '%s'", b.config.Password))

		if len(b.config.Comm.SSHPrivateKey) != 0 {
			debugKeyPath := fmt.Sprintf("%s-%s.pem", b.config.PackerBuildName, b.config.tmpComputeName)
			ui.Message(fmt.Sprintf("temp ssh key: %s", debugKeyPath))

			b.writeSSHPrivateKey(ui, debugKeyPath)
		}
	}

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, b.stateBag)

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

	generatedData := map[string]interface{}{"generated_data": b.stateBag.Get("generated_data")}
	if b.config.isManagedImage() {
		managedImageID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/images/%s",
			b.config.ClientConfig.SubscriptionID, b.config.ManagedImageResourceGroupName, b.config.ManagedImageName)
		if b.config.SharedGalleryDestination.SigDestinationGalleryName != "" {
			return NewManagedImageArtifactWithSIGAsDestination(b.config.OSType,
				b.config.ManagedImageResourceGroupName,
				b.config.ManagedImageName,
				b.config.Location,
				managedImageID,
				b.config.ManagedImageOSDiskSnapshotName,
				b.config.ManagedImageDataDiskSnapshotPrefix,
				b.stateBag.Get(constants.ArmManagedImageSharedGalleryId).(string),
				generatedData)
		}
		return NewManagedImageArtifact(b.config.OSType,
			b.config.ManagedImageResourceGroupName,
			b.config.ManagedImageName,
			b.config.Location,
			managedImageID,
			b.config.ManagedImageOSDiskSnapshotName,
			b.config.ManagedImageDataDiskSnapshotPrefix,
			generatedData)
	} else if template, ok := b.stateBag.GetOk(constants.ArmCaptureTemplate); ok {
		return NewArtifact(
			template.(*CaptureTemplate),
			func(name string) string {
				blob := azureClient.BlobStorageClient.GetContainerReference(DefaultSasBlobContainer).GetBlobReference(name)
				options := storage.BlobSASOptions{}
				options.BlobServiceSASPermissions.Read = true
				options.Expiry = time.Now().AddDate(0, 1, 0).UTC() // one month
				sasUrl, _ := blob.GetSASURI(options)
				return sasUrl
			},
			b.config.OSType,
			generatedData)
	}

	return &Artifact{
		StateData: generatedData,
	}, nil
}

func (b *Builder) writeSSHPrivateKey(ui packersdk.Ui, debugKeyPath string) {
	f, err := os.Create(debugKeyPath)
	if err != nil {
		ui.Say(fmt.Sprintf("Error saving debug key: %s", err))
	}
	defer f.Close()

	// Write the key out
	if _, err := f.Write(b.config.Comm.SSHPrivateKey); err != nil {
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

func equalLocation(location1, location2 string) bool {
	return strings.EqualFold(canonicalizeLocation(location1), canonicalizeLocation(location2))
}

func canonicalizeLocation(location string) string {
	return strings.Replace(location, " ", "", -1)
}

func (b *Builder) getBlobAccount(ctx context.Context, client *AzureClient, resourceGroupName string, storageAccountName string) (*armstorage.Account, error) {
	account, err := client.AccountsClient.GetProperties(ctx, resourceGroupName, storageAccountName)
	if err != nil {
		return nil, err
	}

	return &account, err
}

func (b *Builder) configureStateBag(stateBag multistep.StateBag) {
	stateBag.Put(constants.AuthorizedKey, b.config.sshAuthorizedKey)

	stateBag.Put(constants.ArmTags, packerAzureCommon.MapToAzureTags(b.config.AzureTags))
	stateBag.Put(constants.ArmComputeName, b.config.tmpComputeName)
	stateBag.Put(constants.ArmDeploymentName, b.config.tmpDeploymentName)

	if b.config.OSType == constants.Target_Windows && b.config.BuildKeyVaultName == "" {
		stateBag.Put(constants.ArmKeyVaultDeploymentName, fmt.Sprintf("kv%s", b.config.tmpDeploymentName))
	}

	stateBag.Put(constants.ArmKeyVaultName, b.config.tmpKeyVaultName)
	stateBag.Put(constants.ArmIsExistingKeyVault, false)
	if b.config.BuildKeyVaultName != "" {
		stateBag.Put(constants.ArmKeyVaultName, b.config.BuildKeyVaultName)
		b.config.tmpKeyVaultName = b.config.BuildKeyVaultName
		stateBag.Put(constants.ArmIsExistingKeyVault, true)
	}

	stateBag.Put(constants.ArmNicName, b.config.tmpNicName)
	stateBag.Put(constants.ArmPublicIPAddressName, b.config.tmpPublicIPAddressName)
	stateBag.Put(constants.ArmResourceGroupName, b.config.BuildResourceGroupName)
	stateBag.Put(constants.ArmIsExistingResourceGroup, true)

	if b.config.tmpResourceGroupName != "" {
		stateBag.Put(constants.ArmResourceGroupName, b.config.tmpResourceGroupName)
		stateBag.Put(constants.ArmIsExistingResourceGroup, false)

		if b.config.BuildResourceGroupName != "" {
			stateBag.Put(constants.ArmDoubleResourceGroupNameSet, true)
		}
	}

	stateBag.Put(constants.ArmStorageAccountName, b.config.StorageAccount)
	stateBag.Put(constants.ArmIsManagedImage, b.config.isManagedImage())
	stateBag.Put(constants.ArmManagedImageResourceGroupName, b.config.ManagedImageResourceGroupName)
	stateBag.Put(constants.ArmManagedImageName, b.config.ManagedImageName)
	stateBag.Put(constants.ArmManagedImageOSDiskSnapshotName, b.config.ManagedImageOSDiskSnapshotName)
	stateBag.Put(constants.ArmManagedImageDataDiskSnapshotPrefix, b.config.ManagedImageDataDiskSnapshotPrefix)
	stateBag.Put(constants.ArmAsyncResourceGroupDelete, b.config.AsyncResourceGroupDelete)

	if b.config.isManagedImage() && b.config.SharedGalleryDestination.SigDestinationGalleryName != "" {
		stateBag.Put(constants.ArmManagedImageSigPublishResourceGroup, b.config.SharedGalleryDestination.SigDestinationResourceGroup)
		stateBag.Put(constants.ArmManagedImageSharedGalleryName, b.config.SharedGalleryDestination.SigDestinationGalleryName)
		stateBag.Put(constants.ArmManagedImageSharedGalleryImageName, b.config.SharedGalleryDestination.SigDestinationImageName)
		stateBag.Put(constants.ArmManagedImageSharedGalleryImageVersion, b.config.SharedGalleryDestination.SigDestinationImageVersion)
		stateBag.Put(constants.ArmManagedImageSubscription, b.config.ClientConfig.SubscriptionID)
		stateBag.Put(constants.ArmManagedImageSharedGalleryImageVersionEndOfLifeDate, b.config.SharedGalleryImageVersionEndOfLifeDate)
		stateBag.Put(constants.ArmManagedImageSharedGalleryImageVersionReplicaCount, b.config.SharedGalleryImageVersionReplicaCount)
		stateBag.Put(constants.ArmManagedImageSharedGalleryImageVersionExcludeFromLatest, b.config.SharedGalleryImageVersionExcludeFromLatest)
	}
}

// Parameters that are only known at runtime after querying Azure.
func (b *Builder) setRuntimeParameters(stateBag multistep.StateBag) {
	stateBag.Put(constants.ArmLocation, b.config.Location)
}

func (b *Builder) setTemplateParameters(stateBag multistep.StateBag) {
	stateBag.Put(constants.ArmVirtualMachineCaptureParameters, b.config.toVirtualMachineCaptureParameters())
}

func (b *Builder) setImageParameters(stateBag multistep.StateBag) {
	stateBag.Put(constants.ArmImageParameters, b.config.toImageParameters())
}

func (b *Builder) getServicePrincipalTokens(say func(string)) (*adal.ServicePrincipalToken, *adal.ServicePrincipalToken, error) {
	return b.config.ClientConfig.GetServicePrincipalTokens(say)
}

func getObjectIdFromToken(ui packersdk.Ui, token *adal.ServicePrincipalToken) string {
	claims := jwt.MapClaims{}
	var p jwt.Parser

	var err error

	_, _, err = p.ParseUnverified(token.OAuthToken(), claims)

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to parse the token,Error: %s", err.Error()))
		return ""
	}

	oid, _ := claims["oid"].(string)
	return oid
}

func normalizeAzureRegion(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "", -1))
}
