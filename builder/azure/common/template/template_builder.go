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

	resourceVirtualMachine = "Microsoft.Compute/virtualMachines"
	resourceKeyVaults      = "Microsoft.KeyVault/vaults"

	variableSshKeyPath = "sshKeyPath"
)

type TemplateBuilder struct {
	template *Template
}

func NewTemplateBuilder() (*TemplateBuilder, error) {
	var t Template

	err := json.Unmarshal([]byte(basicTemplate), &t)
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
				compute.SSHPublicKey{
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
		compute.VaultSecretGroup{
			SourceVault: &compute.SubResource{
				ID: to.StringPtr(s.toResourceID(resourceKeyVaults, keyVaultName)),
			},
			VaultCertificates: &[]compute.VaultCertificate{
				compute.VaultCertificate{
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
				compute.WinRMListener{
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

func (s *TemplateBuilder) toKeyVaultID(name string) string {
	return s.toResourceID(resourceKeyVaults, name)
}

func (s *TemplateBuilder) toResourceID(id, name string) string {
	return fmt.Sprintf("[resourceId(resourceGroup().name, '%s', '%s')]", id, name)
}

func (s *TemplateBuilder) toVariable(name string) string {
	return fmt.Sprintf("[variables('%s')]", name)
}

const basicTemplate = `{
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
    "vmStorageAccountContainerName": "images",
    "vnetID": "[resourceId('Microsoft.Network/virtualNetworks', variables('virtualNetworkName'))]"
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
