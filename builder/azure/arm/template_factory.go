package arm

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"

	"fmt"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/builder/azure/common/template"
)

type templateFactoryFunc func(*Config) (*resources.Deployment, error)

func GetKeyVaultDeployment(config *Config) (*resources.Deployment, error) {
	params := &template.TemplateParameters{
		KeyVaultName:        &template.TemplateParameter{Value: config.tmpKeyVaultName},
		KeyVaultSecretValue: &template.TemplateParameter{Value: config.winrmCertificate},
		ObjectId:            &template.TemplateParameter{Value: config.ObjectID},
		TenantId:            &template.TemplateParameter{Value: config.TenantID},
	}

	builder, _ := template.NewTemplateBuilder(template.KeyVault)
	builder.SetTags(&config.AzureTags)

	doc, _ := builder.ToJSON()
	return createDeploymentParameters(*doc, params)
}

func GetVirtualMachineDeployment(config *Config) (*resources.Deployment, error) {
	params := &template.TemplateParameters{
		AdminUsername:              &template.TemplateParameter{Value: config.UserName},
		AdminPassword:              &template.TemplateParameter{Value: config.Password},
		DnsNameForPublicIP:         &template.TemplateParameter{Value: config.tmpComputeName},
		NicName:                    &template.TemplateParameter{Value: config.tmpNicName},
		OSDiskName:                 &template.TemplateParameter{Value: config.tmpOSDiskName},
		PublicIPAddressName:        &template.TemplateParameter{Value: config.tmpPublicIPAddressName},
		SubnetName:                 &template.TemplateParameter{Value: config.tmpSubnetName},
		StorageAccountBlobEndpoint: &template.TemplateParameter{Value: config.storageAccountBlobEndpoint},
		VirtualNetworkName:         &template.TemplateParameter{Value: config.tmpVirtualNetworkName},
		VMSize:                     &template.TemplateParameter{Value: config.VMSize},
		VMName:                     &template.TemplateParameter{Value: config.tmpComputeName},
	}

	builder, err := template.NewTemplateBuilder(template.BasicTemplate)
	if err != nil {
		return nil, err
	}
	osType := compute.Linux

	switch config.OSType {
	case constants.Target_Linux:
		builder.BuildLinux(config.sshAuthorizedKey)
	case constants.Target_Windows:
		osType = compute.Windows
		builder.BuildWindows(config.tmpKeyVaultName, config.tmpWinRMCertificateUrl)
	}

	if config.ImageUrl != "" {
		builder.SetImageUrl(config.ImageUrl, osType)
	} else if config.CustomManagedImageName != "" {
		builder.SetManagedDiskUrl(config.customManagedImageID, config.managedImageStorageAccountType)
	} else if config.ManagedImageName != "" && config.ImagePublisher != "" {
		imageID := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Compute/locations/%s/publishers/%s/ArtifactTypes/vmimage/offers/%s/skus/%s/versions/%s",
			config.SubscriptionID,
			config.Location,
			config.ImagePublisher,
			config.ImageOffer,
			config.ImageSku,
			config.ImageVersion)

		builder.SetManagedMarketplaceImage(config.Location, config.ImagePublisher, config.ImageOffer, config.ImageSku, config.ImageVersion, imageID, config.managedImageStorageAccountType)
	} else if config.SharedGallery.Subscription != "" {
		imageID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/galleries/%s/images/%s",
			config.SharedGallery.Subscription,
			config.SharedGallery.ResourceGroup,
			config.SharedGallery.GalleryName,
			config.SharedGallery.ImageName)
		if config.SharedGallery.ImageVersion != "" {
			imageID += fmt.Sprintf("/versions/%s",
				config.SharedGallery.ImageVersion)
		}

		builder.SetSharedGalleryImage(config.Location, imageID)
	} else {
		builder.SetMarketPlaceImage(config.ImagePublisher, config.ImageOffer, config.ImageSku, config.ImageVersion)
	}

	if config.OSDiskSizeGB > 0 {
		builder.SetOSDiskSizeGB(config.OSDiskSizeGB)
	}

	if len(config.AdditionalDiskSize) > 0 {
		builder.SetAdditionalDisks(config.AdditionalDiskSize, config.CustomManagedImageName != "" || (config.ManagedImageName != "" && config.ImagePublisher != ""))
	}

	if config.customData != "" {
		builder.SetCustomData(config.customData)
	}

	if config.PlanInfo.PlanName != "" {
		builder.SetPlanInfo(config.PlanInfo.PlanName, config.PlanInfo.PlanProduct, config.PlanInfo.PlanPublisher, config.PlanInfo.PlanPromotionCode)
	}

	if config.VirtualNetworkName != "" && DefaultPrivateVirtualNetworkWithPublicIp != config.PrivateVirtualNetworkWithPublicIp {
		builder.SetPrivateVirtualNetworkWithPublicIp(
			config.VirtualNetworkResourceGroupName,
			config.VirtualNetworkName,
			config.VirtualNetworkSubnetName)
	} else if config.VirtualNetworkName != "" {
		builder.SetVirtualNetwork(
			config.VirtualNetworkResourceGroupName,
			config.VirtualNetworkName,
			config.VirtualNetworkSubnetName)
	}

	builder.SetTags(&config.AzureTags)
	doc, _ := builder.ToJSON()
	return createDeploymentParameters(*doc, params)
}

func createDeploymentParameters(doc string, parameters *template.TemplateParameters) (*resources.Deployment, error) {
	var template map[string]interface{}
	err := json.Unmarshal(([]byte)(doc), &template)
	if err != nil {
		return nil, err
	}

	bs, err := json.Marshal(*parameters)
	if err != nil {
		return nil, err
	}

	var templateParameters map[string]interface{}
	err = json.Unmarshal(bs, &templateParameters)
	if err != nil {
		return nil, err
	}

	return &resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Mode:       resources.Incremental,
			Template:   &template,
			Parameters: &templateParameters,
		},
	}, nil
}
