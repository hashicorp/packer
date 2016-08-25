package template

import (
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"strings"
)

const (
	jsonPrefix = ""
	jsonIndent = "  "

	resourceKeyVaults         = "Microsoft.KeyVault/vaults"
	resourceNetworkInterfaces = "Microsoft.Network/networkInterfaces"
	resourcePublicIPAddresses = "Microsoft.Network/publicIPAddresses"
	resourceVirtualMachine    = "Microsoft.Compute/virtualMachines"
	resourceVirtualNetworks   = "Microsoft.Network/virtualNetworks"

	variableSshKeyPath = "sshKeyPath"
)

type TemplateBuilder struct {
	template *Template
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
	return nil
}

func (s *TemplateBuilder) SetMarketPlaceImage(publisher, offer, sku, version string) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.ImageReference = &compute.ImageReference{
		Publisher: to.StringPtr(publisher),
		Offer:     to.StringPtr(offer),
		Sku:       to.StringPtr(sku),
		Version:   to.StringPtr(version),
	}

	return nil
}

func (s *TemplateBuilder) SetImageUrl(imageUrl string, osType compute.OperatingSystemTypes) error {
	resource, err := s.getResourceByType(resourceVirtualMachine)
	if err != nil {
		return err
	}

	profile := resource.Properties.StorageProfile
	profile.OsDisk.OsType = osType
	profile.OsDisk.Image = &compute.VirtualHardDisk{
		URI: to.StringPtr(imageUrl),
	}

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

	(*resource.Properties.IPConfigurations)[0].Properties.PublicIPAddress = nil

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

func (s *TemplateBuilder) deleteResourceDependency(resource *Resource, predicate func(string) bool) {
	deps := make([]string, 0)

	for _, dep := range *resource.DependsOn {
		if !predicate(dep) {
			deps = append(deps, dep)
		}
	}

	*resource.DependsOn = deps
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
    "osDiskName": {
      "type": "string"
    },
    "storageAccountBlobEndpoint": {
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
    "apiVersion": "2015-06-15",
    "location": "[resourceGroup().location]",
    "nicName": "packerNic",
    "publicIPAddressName": "packerPublicIP",
    "publicIPAddressType": "Dynamic",
    "sshKeyPath": "[concat('/home/',parameters('adminUsername'),'/.ssh/authorized_keys')]",
    "subnetName": "packerSubnet",
    "subnetAddressPrefix": "10.0.0.0/24",
    "subnetRef": "[concat(variables('vnetID'),'/subnets/',variables('subnetName'))]",
    "virtualNetworkName": "packerNetwork",
    "virtualNetworkResourceGroup": "[resourceGroup().name]",
    "vmStorageAccountContainerName": "images",
    "vnetID": "[resourceId(variables('virtualNetworkResourceGroup'), 'Microsoft.Network/virtualNetworks', variables('virtualNetworkName'))]"
  },
  "resources": [
    {
      "apiVersion": "[variables('apiVersion')]",
      "type": "Microsoft.Network/publicIPAddresses",
      "name": "[variables('publicIPAddressName')]",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[parameters('dnsNameForPublicIP')]"
        }
      }
    },
    {
      "apiVersion": "[variables('apiVersion')]",
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
      "apiVersion": "[variables('apiVersion')]",
      "type": "Microsoft.Network/networkInterfaces",
      "name": "[variables('nicName')]",
      "location": "[variables('location')]",
      "dependsOn": [
        "[concat('Microsoft.Network/publicIPAddresses/', variables('publicIPAddressName'))]",
        "[concat('Microsoft.Network/virtualNetworks/', variables('virtualNetworkName'))]"
      ],
      "properties": {
        "ipConfigurations": [
          {
            "name": "ipconfig",
            "properties": {
              "privateIPAllocationMethod": "Dynamic",
              "publicIPAddress": {
                "id": "[resourceId('Microsoft.Network/publicIPAddresses', variables('publicIPAddressName'))]"
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
        "[concat('Microsoft.Network/networkInterfaces/', variables('nicName'))]"
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
            "name": "osdisk",
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
              "id": "[resourceId('Microsoft.Network/networkInterfaces', variables('nicName'))]"
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
