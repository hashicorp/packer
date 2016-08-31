package arm

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/builder/azure/common/template"
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
		OSDiskName:                 &template.TemplateParameter{Value: config.tmpOSDiskName},
		StorageAccountBlobEndpoint: &template.TemplateParameter{Value: config.storageAccountBlobEndpoint},
		VMSize: &template.TemplateParameter{Value: config.VMSize},
		VMName: &template.TemplateParameter{Value: config.tmpComputeName},
	}

	builder, _ := template.NewTemplateBuilder(template.BasicTemplate)
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
	} else {
		builder.SetMarketPlaceImage(config.ImagePublisher, config.ImageOffer, config.ImageSku, config.ImageVersion)
	}

	if config.VirtualNetworkName != "" {
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
