package dtl

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/devtestlabs/mgmt/2018-09-15/dtl"
)

type templateFactoryFuncDtl func(*Config) (*dtl.LabVirtualMachineCreationParameter, error)

func newBool(val bool) *bool {
	b := true
	if val == b {
		return &b
	} else {
		b = false
		return &b
	}
}

func getCustomImageId(config *Config) *string {
	if config.CustomManagedImageName != "" && config.CustomManagedImageResourceGroupName != "" {
		customManagedImageID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/images/%s",
			config.ClientConfig.SubscriptionID,
			config.CustomManagedImageResourceGroupName,
			config.CustomManagedImageName)
		return &customManagedImageID
	}
	return nil
}

func GetVirtualMachineDeployment(config *Config) (*dtl.LabVirtualMachineCreationParameter, error) {

	galleryImageRef := dtl.GalleryImageReference{
		Offer:     &config.ImageOffer,
		Publisher: &config.ImagePublisher,
		Sku:       &config.ImageSku,
		OsType:    &config.OSType,
		Version:   &config.ImageVersion,
	}

	labVirtualNetworkID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DevTestLab/labs/%s/virtualnetworks/%s",
		config.ClientConfig.SubscriptionID,
		config.tmpResourceGroupName,
		config.LabName,
		config.LabVirtualNetworkName)

	dtlArtifacts := []dtl.ArtifactInstallProperties{}

	if config.DtlArtifacts != nil {
		for i := range config.DtlArtifacts {
			if config.DtlArtifacts[i].RepositoryName == "" {
				config.DtlArtifacts[i].RepositoryName = "public repo"
			}
			config.DtlArtifacts[i].ArtifactId = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DevTestLab/labs/%s/artifactSources/%s/artifacts/%s",
				config.ClientConfig.SubscriptionID,
				config.tmpResourceGroupName,
				config.LabName,
				config.DtlArtifacts[i].RepositoryName,
				config.DtlArtifacts[i].ArtifactName)

			dparams := []dtl.ArtifactParameterProperties{}
			for j := range config.DtlArtifacts[i].Parameters {

				dp := &dtl.ArtifactParameterProperties{}
				dp.Name = &config.DtlArtifacts[i].Parameters[j].Name
				dp.Value = &config.DtlArtifacts[i].Parameters[j].Value

				dparams = append(dparams, *dp)
			}
			dtlArtifact := &dtl.ArtifactInstallProperties{
				ArtifactTitle: &config.DtlArtifacts[i].ArtifactName,
				ArtifactID:    &config.DtlArtifacts[i].ArtifactId,
				Parameters:    &dparams,
			}
			dtlArtifacts = append(dtlArtifacts, *dtlArtifact)
		}
	}

	labMachineProps := &dtl.LabVirtualMachineCreationParameterProperties{
		CreatedByUserID:            &config.ClientConfig.ClientID,
		OwnerObjectID:              &config.ClientConfig.ObjectID,
		OsType:                     &config.OSType,
		Size:                       &config.VMSize,
		UserName:                   &config.UserName,
		Password:                   &config.Password,
		SSHKey:                     &config.sshAuthorizedKey,
		IsAuthenticationWithSSHKey: newBool(true),
		LabSubnetName:              &config.LabSubnetName,
		LabVirtualNetworkID:        &labVirtualNetworkID,
		DisallowPublicIPAddress:    &config.DisallowPublicIP,
		GalleryImageReference:      &galleryImageRef,
		CustomImageID:              getCustomImageId(config),
		PlanID:                     &config.PlanID,

		AllowClaim:                   newBool(false),
		StorageType:                  &config.StorageType,
		VirtualMachineCreationSource: dtl.FromGalleryImage,
		Artifacts:                    &dtlArtifacts,
	}

	labMachine := &dtl.LabVirtualMachineCreationParameter{
		Name:     &config.tmpComputeName,
		Location: &config.Location,
		Tags:     config.AzureTags,
		LabVirtualMachineCreationParameterProperties: labMachineProps,
	}

	return labMachine, nil
}
