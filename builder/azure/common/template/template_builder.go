package template

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-01-01/network"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	jsonPrefix = ""
	jsonIndent = "  "

	resourceKeyVaults             = "Microsoft.KeyVault/vaults"
	resourceNetworkInterfaces     = "Microsoft.Network/networkInterfaces"
	resourcePublicIPAddresses     = "Microsoft.Network/publicIPAddresses"
	resourceVirtualMachine        = "Microsoft.Compute/virtualMachines"
	resourceVirtualNetworks       = "Microsoft.Network/virtualNetworks"
	resourceNetworkSecurityGroups = "Microsoft.Network/networkSecurityGroups"

	variableSshKeyPath = "sshKeyPath"
)

type TemplateBuilder struct {
	template *Template
	osType   compute.OperatingSystemTypes
}

func NewTemplateBuilder(template string) (*TemplateBuilder, error) {
	var t Template

	err := json.Unmarshal([]byte(template), &t)
	if err != nil {
		return nil, err
	}

	return &TemplateBuilder{
		template: &t,
	}, nil
}

func (s *TemplateBuilder) BuildLinux(sshAuthorizedKey string) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.OsProfile
	profile.LinuxConfiguration = &compute.LinuxConfiguration{
		SSH: &compute.SSHConfiguration{
			PublicKeys: &[]compute.SSHPublicKey{
				{
					Path:    to.StringPtr(s.toVariable(variableSshKeyPath)),
					KeyData: to.StringPtr(sshAuthorizedKey),
				},
			},
		},
	}

	s.osType = compute.Linux
	return nil
}

func (s *TemplateBuilder) BuildWindows(keyVaultName, winRMCertificateUrl string) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.OsProfile

	profile.Secrets = &[]compute.VaultSecretGroup{
		{
			SourceVault: &compute.SubResource{
				ID: to.StringPtr(s.toResourceID(resourceKeyVaults, keyVaultName)),
			},
			VaultCertificates: &[]compute.VaultCertificate{
				{
					CertificateStore: to.StringPtr("My"),
					CertificateURL:   to.StringPtr(winRMCertificateUrl),
				},
			},
		},
	}

	profile.WindowsConfiguration = &compute.WindowsConfiguration{
		ProvisionVMAgent: to.BoolPtr(true),
		WinRM: &compute.WinRMConfiguration{
			Listeners: &[]compute.WinRMListener{
				{
					Protocol:       "https",
					CertificateURL: to.StringPtr(winRMCertificateUrl),
				},
			},
		},
	}

	s.osType = compute.Windows
	return nil
}

func (s *TemplateBuilder) SetManagedDiskUrl(managedImageId string, storageAccountType compute.StorageAccountTypes, cachingType compute.CachingTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.ImageReference = &compute.ImageReference{
		ID: &managedImageId,
	}
	profile.OsDisk.OsType = s.osType
	profile.OsDisk.CreateOption = compute.DiskCreateOptionTypesFromImage
	profile.OsDisk.Vhd = nil
	profile.OsDisk.Caching = cachingType
	profile.OsDisk.ManagedDisk = &compute.ManagedDiskParameters{
		StorageAccountType: storageAccountType,
	}

	return nil
}

func (s *TemplateBuilder) SetManagedMarketplaceImage(location, publisher, offer, sku, version, imageID string, storageAccountType compute.StorageAccountTypes, cachingType compute.CachingTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.ImageReference = &compute.ImageReference{
		Publisher: &publisher,
		Offer:     &offer,
		Sku:       &sku,
		Version:   &version,
	}
	profile.OsDisk.OsType = s.osType
	profile.OsDisk.CreateOption = compute.DiskCreateOptionTypesFromImage
	profile.OsDisk.Vhd = nil
	profile.OsDisk.Caching = cachingType
	profile.OsDisk.ManagedDisk = &compute.ManagedDiskParameters{
		StorageAccountType: storageAccountType,
	}

	return nil
}

func (s *TemplateBuilder) SetSharedGalleryImage(location, imageID string, cachingType compute.CachingTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	s.setVariable("apiVersion", "2018-04-01") // Required for Shared Image Gallery
	profile := resource.Properties.StorageProfile
	profile.ImageReference = &compute.ImageReference{ID: &imageID}
	profile.OsDisk.OsType = s.osType
	profile.OsDisk.Vhd = nil
	profile.OsDisk.Caching = cachingType

	return nil
}

func (s *TemplateBuilder) SetMarketPlaceImage(publisher, offer, sku, version string, cachingType compute.CachingTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.OsDisk.Caching = cachingType
	profile.ImageReference = &compute.ImageReference{
		Publisher: to.StringPtr(publisher),
		Offer:     to.StringPtr(offer),
		Sku:       to.StringPtr(sku),
		Version:   to.StringPtr(version),
	}

	return nil
}

func (s *TemplateBuilder) SetImageUrl(imageUrl string, osType compute.OperatingSystemTypes, cachingType compute.CachingTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.OsDisk.OsType = osType
	profile.OsDisk.Caching = cachingType

	profile.OsDisk.Image = &compute.VirtualHardDisk{
		URI: to.StringPtr(imageUrl),
	}

	return nil
}

func (s *TemplateBuilder) SetPlanInfo(name, product, publisher, promotionCode string) error {
	var promotionCodeVal *string = nil
	if promotionCode != "" {
		promotionCodeVal = to.StringPtr(promotionCode)
	}

	for i, x := range *s.template.Resources {
		if strings.EqualFold(*x.Type, resourceVirtualMachine) {
			(*s.template.Resources)[i].Plan = &Plan{
				Name:          to.StringPtr(name),
				Product:       to.StringPtr(product),
				Publisher:     to.StringPtr(publisher),
				PromotionCode: promotionCodeVal,
			}
		}
	}

	return nil
}

func (s *TemplateBuilder) SetOSDiskSizeGB(diskSizeGB int32) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.OsDisk.DiskSizeGB = to.Int32Ptr(diskSizeGB)

	return nil
}

func (s *TemplateBuilder) SetAdditionalDisks(diskSizeGB []int32, isManaged bool, cachingType compute.CachingTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	dataDisks := make([]DataDiskUnion, len(diskSizeGB))

	for i, additionalSize := range diskSizeGB {
		dataDisks[i].DiskSizeGB = to.Int32Ptr(additionalSize)
		dataDisks[i].Lun = to.IntPtr(i)
		dataDisks[i].Name = to.StringPtr(fmt.Sprintf("datadisk-%d", i+1))
		dataDisks[i].CreateOption = "Empty"
		dataDisks[i].Caching = cachingType
		if isManaged {
			dataDisks[i].Vhd = nil
			dataDisks[i].ManagedDisk = profile.OsDisk.ManagedDisk
		} else {
			dataDisks[i].Vhd = &compute.VirtualHardDisk{
				URI: to.StringPtr(fmt.Sprintf("[concat(parameters('storageAccountBlobEndpoint'),variables('vmStorageAccountContainerName'),'/datadisk-', '%d','.vhd')]", i+1)),
			}
			dataDisks[i].ManagedDisk = nil
		}
	}
	profile.DataDisks = &dataDisks
	return nil
}

func (s *TemplateBuilder) SetCustomData(customData string) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.OsProfile
	profile.CustomData = to.StringPtr(customData)

	return nil
}

func (s *TemplateBuilder) SetVirtualNetwork(virtualNetworkResourceGroup, virtualNetworkName, subnetName string) error {
	s.setVariable("virtualNetworkResourceGroup", virtualNetworkResourceGroup)
	s.setVariable("virtualNetworkName", virtualNetworkName)
	s.setVariable("subnetName", subnetName)

	s.deleteResourceByType(resourceVirtualNetworks)
	s.deleteResourceByType(resourcePublicIPAddresses)
	resource, err := s.getResourceByType(resourceNetworkInterfaces)
	if err != nil {
		return err
	}

	s.deleteResourceDependency(resource, func(s string) bool {
		return strings.Contains(s, "Microsoft.Network/virtualNetworks") ||
			strings.Contains(s, "Microsoft.Network/publicIPAddresses")
	})

	(*resource.Properties.IPConfigurations)[0].PublicIPAddress = nil

	return nil
}

func (s *TemplateBuilder) SetPrivateVirtualNetworkWithPublicIp(virtualNetworkResourceGroup, virtualNetworkName, subnetName string) error {
	s.setVariable("virtualNetworkResourceGroup", virtualNetworkResourceGroup)
	s.setVariable("virtualNetworkName", virtualNetworkName)
	s.setVariable("subnetName", subnetName)

	s.deleteResourceByType(resourceVirtualNetworks)
	resource, err := s.getResourceByType(resourceNetworkInterfaces)
	if err != nil {
		return err
	}

	s.deleteResourceDependency(resource, func(s string) bool {
		return strings.Contains(s, "Microsoft.Network/virtualNetworks")
	})

	return nil
}

func (s *TemplateBuilder) SetNetworkSecurityGroup(ipAddresses []string, port int) error {
	nsgResource, dependency, resourceId := s.createNsgResource(ipAddresses, port)
	if err := s.addResource(nsgResource); err != nil {
		return err
	}

	vnetResource, err := s.getResourceByType(resourceVirtualNetworks)
	if err != nil {
		return err
	}
	s.deleteResourceByType(resourceVirtualNetworks)

	s.addResourceDependency(vnetResource, dependency)

	if vnetResource.Properties == nil || vnetResource.Properties.Subnets == nil || len(*vnetResource.Properties.Subnets) != 1 {
		return fmt.Errorf("template: could not find virtual network/subnet to add default network security group to")
	}
	subnet := ((*vnetResource.Properties.Subnets)[0])
	if subnet.SubnetPropertiesFormat == nil {
		subnet.SubnetPropertiesFormat = &network.SubnetPropertiesFormat{}
	}
	if subnet.SubnetPropertiesFormat.NetworkSecurityGroup != nil {
		return fmt.Errorf("template: subnet already has an associated network security group")
	}
	subnet.SubnetPropertiesFormat.NetworkSecurityGroup = &network.SecurityGroup{
		ID: to.StringPtr(resourceId),
	}

	s.addResource(vnetResource)

	return nil
}

func (s *TemplateBuilder) SetTags(tags *map[string]*string) error {
	if tags == nil || len(*tags) == 0 {
		return nil
	}

	for i := range *s.template.Resources {
		(*s.template.Resources)[i].Tags = tags
	}
	return nil
}

func (s *TemplateBuilder) ToJSON() (*string, error) {
	bs, err := json.MarshalIndent(s.template, jsonPrefix, jsonIndent)

	if err != nil {
		return nil, err
	}
	return to.StringPtr(string(bs)), err
}

func (s *TemplateBuilder) getResourceByType(t string) (*Resource, error) {
	for _, x := range *s.template.Resources {
		if strings.EqualFold(*x.Type, t) {
			return &x, nil
		}
	}

	return nil, fmt.Errorf("template: could not find a resource of type %s", t)
}

func (s *TemplateBuilder) getResourceByType2(t string) (**Resource, error) {
	for _, x := range *s.template.Resources {
		if strings.EqualFold(*x.Type, t) {
			p := &x
			return &p, nil
		}
	}

	return nil, fmt.Errorf("template: could not find a resource of type %s", t)
}

func (s *TemplateBuilder) setVariable(name string, value string) {
	(*s.template.Variables)[name] = value
}

func (s *TemplateBuilder) toKeyVaultID(name string) string {
	return s.toResourceID(resourceKeyVaults, name)
}

func (s *TemplateBuilder) toResourceID(id, name string) string {
	return fmt.Sprintf("[resourceId(resourceGroup().name, '%s', '%s')]", id, name)
}

func (s *TemplateBuilder) toVariable(name string) string {
	return fmt.Sprintf("[variables('%s')]", name)
}

func (s *TemplateBuilder) addResource(newResource *Resource) error {
	for _, resource := range *s.template.Resources {
		if *resource.Type == *newResource.Type {
			return fmt.Errorf("template: found an existing resource of type %s", *resource.Type)
		}
	}

	resources := append(*s.template.Resources, *newResource)
	s.template.Resources = &resources
	return nil
}

func (s *TemplateBuilder) deleteResourceByType(resourceType string) {
	resources := make([]Resource, 0)

	for _, resource := range *s.template.Resources {
		if *resource.Type == resourceType {
			continue
		}
		resources = append(resources, resource)
	}

	s.template.Resources = &resources
}

func (s *TemplateBuilder) addResourceDependency(resource *Resource, dep string) {
	if resource.DependsOn != nil {
		deps := append(*resource.DependsOn, dep)
		resource.DependsOn = &deps
	} else {
		resource.DependsOn = &[]string{dep}
	}
}

func (s *TemplateBuilder) deleteResourceDependency(resource *Resource, predicate func(string) bool) {
	deps := make([]string, 0)

	for _, dep := range *resource.DependsOn {
		if !predicate(dep) {
			deps = append(deps, dep)
		}
	}

	*resource.DependsOn = deps
}

func (s *TemplateBuilder) createNsgResource(srcIpAddresses []string, port int) (*Resource, string, string) {
	resource := &Resource{
		ApiVersion: to.StringPtr("[variables('networkSecurityGroupsApiVersion')]"),
		Name:       to.StringPtr("[parameters('nsgName')]"),
		Type:       to.StringPtr(resourceNetworkSecurityGroups),
		Location:   to.StringPtr("[variables('location')]"),
		Properties: &Properties{
			SecurityRules: &[]network.SecurityRule{
				{
					Name: to.StringPtr("AllowIPsToSshWinRMInbound"),
					SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
						Description:              to.StringPtr("Allow inbound traffic from specified IP addresses"),
						Protocol:                 network.SecurityRuleProtocolTCP,
						Priority:                 to.Int32Ptr(100),
						Access:                   network.SecurityRuleAccessAllow,
						Direction:                network.SecurityRuleDirectionInbound,
						SourceAddressPrefixes:    &srcIpAddresses,
						SourcePortRange:          to.StringPtr("*"),
						DestinationAddressPrefix: to.StringPtr("VirtualNetwork"),
						DestinationPortRange:     to.StringPtr(strconv.Itoa(port)),
					},
				},
			},
		},
	}

	dependency := fmt.Sprintf("[concat('%s/', parameters('nsgName'))]", resourceNetworkSecurityGroups)
	resourceId := fmt.Sprintf("[resourceId('%s', parameters('nsgName'))]", resourceNetworkSecurityGroups)

	return resource, dependency, resourceId
}

// See https://github.com/Azure/azure-quickstart-templates for a extensive list of templates.

// Template to deploy a KeyVault.
//
// This template is still hard-coded unlike the ARM templates used for VMs for
// a couple of reasons.
//
//  1. The SDK defines no types for a Key Vault
//  2. The Key Vault template is relatively simple, and is static.
//
const KeyVault = `{
  "$schema": "http://schema.management.azure.com/schemas/2014-04-01-preview/deploymentTemplate.json",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "keyVaultName": {
      "type": "string"
    },
    "keyVaultSecretValue": {
      "type": "securestring"
    },
    "objectId": {
     "type": "string"
    },
    "tenantId": {
      "type": "string"
    }
  },
  "variables": {
    "apiVersion": "2015-06-01",
    "location": "[resourceGroup().location]",
    "keyVaultSecretName": "packerKeyVaultSecret"
  },
  "resources": [
    {
      "apiVersion": "[variables('apiVersion')]",
      "type": "Microsoft.KeyVault/vaults",
      "name": "[parameters('keyVaultName')]",
      "location": "[variables('location')]",
      "properties": {
        "enabledForDeployment": "true",
        "enabledForTemplateDeployment": "true",
        "tenantId": "[parameters('tenantId')]",
        "accessPolicies": [
          {
            "tenantId": "[parameters('tenantId')]",
            "objectId": "[parameters('objectId')]",
            "permissions": {
              "keys": [ "all" ],
              "secrets": [ "all" ]
            }
          }
        ],
        "sku": {
          "name": "standard",
          "family": "A"
        }
      },
      "resources": [
        {
          "apiVersion": "[variables('apiVersion')]",
          "type": "secrets",
          "name": "[variables('keyVaultSecretName')]",
          "dependsOn": [
            "[concat('Microsoft.KeyVault/vaults/', parameters('keyVaultName'))]"
          ],
          "properties": {
            "value": "[parameters('keyVaultSecretValue')]"
          }
        }
      ]
    }
  ]
}`

const BasicTemplate = `{
  "$schema": "http://schema.management.azure.com/schemas/2014-04-01-preview/deploymentTemplate.json",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "adminUsername": {
      "type": "string"
    },
    "adminPassword": {
      "type": "string"
    },
    "dnsNameForPublicIP": {
      "type": "string"
    },
	"nicName": {
      "type": "string"
	},
    "osDiskName": {
      "type": "string"
    },
    "publicIPAddressName": {
      "type": "string"
	},
	"subnetName": {
      "type": "string"
	},
    "storageAccountBlobEndpoint": {
      "type": "string"
    },
	"virtualNetworkName": {
      "type": "string"
	},
    "nsgName": {
      "type": "string"
    },
    "vmSize": {
      "type": "string"
    },
    "vmName": {
      "type": "string"
    }
  },
  "variables": {
    "addressPrefix": "10.0.0.0/16",
    "apiVersion": "2017-03-30",
    "managedDiskApiVersion": "2017-03-30",
    "networkInterfacesApiVersion": "2017-04-01",
    "publicIPAddressApiVersion": "2017-04-01",
    "virtualNetworksApiVersion": "2017-04-01",
    "networkSecurityGroupsApiVersion": "2019-04-01",
    "location": "[resourceGroup().location]",
    "publicIPAddressType": "Dynamic",
    "sshKeyPath": "[concat('/home/',parameters('adminUsername'),'/.ssh/authorized_keys')]",
    "subnetName": "[parameters('subnetName')]",
    "subnetAddressPrefix": "10.0.0.0/24",
    "subnetRef": "[concat(variables('vnetID'),'/subnets/',variables('subnetName'))]",
    "virtualNetworkName": "[parameters('virtualNetworkName')]",
    "virtualNetworkResourceGroup": "[resourceGroup().name]",
    "vmStorageAccountContainerName": "images",
    "vnetID": "[resourceId(variables('virtualNetworkResourceGroup'), 'Microsoft.Network/virtualNetworks', variables('virtualNetworkName'))]"
  },
  "resources": [
    {
      "apiVersion": "[variables('publicIPAddressApiVersion')]",
      "type": "Microsoft.Network/publicIPAddresses",
      "name": "[parameters('publicIPAddressName')]",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[parameters('dnsNameForPublicIP')]"
        }
      }
    },
    {
      "apiVersion": "[variables('virtualNetworksApiVersion')]",
      "type": "Microsoft.Network/virtualNetworks",
      "name": "[variables('virtualNetworkName')]",
      "location": "[variables('location')]",
      "properties": {
        "addressSpace": {
          "addressPrefixes": [
            "[variables('addressPrefix')]"
          ]
        },
        "subnets": [
          {
            "name": "[variables('subnetName')]",
            "properties": {
              "addressPrefix": "[variables('subnetAddressPrefix')]"
            }
          }
        ]
      }
    },
    {
      "apiVersion": "[variables('networkInterfacesApiVersion')]",
      "type": "Microsoft.Network/networkInterfaces",
      "name": "[parameters('nicName')]",
      "location": "[variables('location')]",
      "dependsOn": [
        "[concat('Microsoft.Network/publicIPAddresses/', parameters('publicIPAddressName'))]",
        "[concat('Microsoft.Network/virtualNetworks/', variables('virtualNetworkName'))]"
      ],
      "properties": {
        "ipConfigurations": [
          {
            "name": "ipconfig",
            "properties": {
              "privateIPAllocationMethod": "Dynamic",
              "publicIPAddress": {
                "id": "[resourceId('Microsoft.Network/publicIPAddresses', parameters('publicIPAddressName'))]"
              },
              "subnet": {
                "id": "[variables('subnetRef')]"
              }
            }
          }
        ]
      }
    },
    {
      "apiVersion": "[variables('apiVersion')]",
      "type": "Microsoft.Compute/virtualMachines",
      "name": "[parameters('vmName')]",
      "location": "[variables('location')]",
      "dependsOn": [
        "[concat('Microsoft.Network/networkInterfaces/', parameters('nicName'))]"
      ],
      "properties": {
        "hardwareProfile": {
          "vmSize": "[parameters('vmSize')]"
        },
        "osProfile": {
          "computerName": "[parameters('vmName')]",
          "adminUsername": "[parameters('adminUsername')]",
          "adminPassword": "[parameters('adminPassword')]"
        },
        "storageProfile": {
          "osDisk": {
            "name": "[parameters('osDiskName')]",
            "vhd": {
              "uri": "[concat(parameters('storageAccountBlobEndpoint'),variables('vmStorageAccountContainerName'),'/', parameters('osDiskName'),'.vhd')]"
            },
            "caching": "ReadWrite",
            "createOption": "FromImage"
          }
        },
        "networkProfile": {
          "networkInterfaces": [
            {
              "id": "[resourceId('Microsoft.Network/networkInterfaces', parameters('nicName'))]"
            }
          ]
        },
        "diagnosticsProfile": {
          "bootDiagnostics": {
             "enabled": false
          }
        }
      }
    }
  ]
}`
