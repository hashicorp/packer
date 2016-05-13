// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

// See https://github.com/Azure/azure-quickstart-templates for a extensive list of templates.

const Linux = `{
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
    "imagePublisher": {
      "type": "string"
    },
    "imageOffer": {
      "type": "string"
    },
    "imageSku": {
      "type": "string"
    },
    "imageVersion": {
      "type": "string"
    },
    "imageUri": {
      "type": "string"
    },
    "osDiskName": {
      "type": "string"
    },
    "sshAuthorizedKey": {
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
          "adminPassword": "[parameters('adminPassword')]",
          "linuxConfiguration": {
            "disablePasswordAuthentication": "false",
            "ssh": {
              "publicKeys": [
                {
                  "path": "[variables('sshKeyPath')]",
                  "keyData": "[parameters('sshAuthorizedKey')]"
                }
              ]
            }
          }
        },
        "storageProfile": {
          "imageReference": {
            "publisher": "[parameters('imagePublisher')]",
            "offer": "[parameters('imageOffer')]",
            "sku": "[parameters('imageSku')]",
            "version": "[parameters('imageVersion')]"
          },
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
             "enabled": "false"
          }
        }
      }
    }
  ]
}`

const LinuxVhd = `{
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
    "imageUri": {
      "type": "string"
    },
    "imagePublisher": {
      "type": "string"
    },
    "imageOffer": {
      "type": "string"
    },
    "imageSku": {
      "type": "string"
    },
    "imageVersion": {
      "type": "string"
    },
    "osDiskName": {
      "type": "string"
    },
    "sshAuthorizedKey": {
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
          "adminPassword": "[parameters('adminPassword')]",
          "linuxConfiguration": {
            "disablePasswordAuthentication": "false",
            "ssh": {
              "publicKeys": [
                {
                  "path": "[variables('sshKeyPath')]",
                  "keyData": "[parameters('sshAuthorizedKey')]"
                }
              ]
            }
          }
        },
        "storageProfile": {
          "osDisk": {
            "name": "osdisk",
            "image": {
              "uri": "[parameters('imageUri')]"
            },
            "vhd": {
              "uri": "[concat(parameters('storageAccountBlobEndpoint'),variables('vmStorageAccountContainerName'),'/', parameters('osDiskName'),'.vhd')]"
            },
            "caching": "ReadWrite",
            "createOption": "FromImage",
            "osType": "linux"
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
             "enabled": "false"
          }
        }
      }
    }
  ]
}`

// Template to deploy a KeyVault.
//
// NOTE: the parameters for the KeyVault template are identical to Windows
// template.  Keeping these values in sync simplifies the code at the expense
// of template bloat.  This bloat may be addressed in the future.
const KeyVault = `{
  "$schema": "http://schema.management.azure.com/schemas/2014-04-01-preview/deploymentTemplate.json",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "adminUserName": {
      "type": "string"
    },
    "adminPassword": {
      "type": "securestring"
    },
    "dnsNameForPublicIP": {
      "type": "string"
    },
    "imageOffer": {
      "type": "string"
    },
    "imagePublisher": {
      "type": "string"
    },
    "imageSku": {
      "type": "string"
    },
    "imageVersion": {
      "type": "string"
    },
    "keyVaultName": {
      "type": "string"
    },
    "keyVaultSecretValue": {
      "type": "securestring"
    },
    "objectId": {
     "type": "string"
    },
    "osDiskName": {
      "type": "string"
    },
    "storageAccountBlobEndpoint": {
      "type": "string"
    },
    "tenantId": {
      "type": "string"
    },
    "vmName": {
      "type": "string"
    },
    "vmSize": {
      "type": "string"
    },
    "winRMCertificateUrl": {
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

const Windows = `{
  "$schema": "http://schema.management.azure.com/schemas/2014-04-01-preview/deploymentTemplate.json",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "adminUserName": {
      "type": "string"
    },
    "adminPassword": {
      "type": "securestring"
    },
    "dnsNameForPublicIP": {
      "type": "string"
    },
    "imageOffer": {
      "type": "string"
    },
    "imagePublisher": {
      "type": "string"
    },
    "imageSku": {
      "type": "string"
    },
    "imageVersion": {
      "type": "string"
    },
    "imageUri": {
      "type": "string"
    },
    "keyVaultName": {
      "type": "string"
    },
    "keyVaultSecretValue": {
      "type": "securestring"
    },
    "objectId": {
     "type": "string"
    },
    "osDiskName": {
      "type": "string"
    },
    "storageAccountBlobEndpoint": {
      "type": "string"
    },
    "tenantId": {
      "type": "string"
    },
    "vmName": {
      "type": "string"
    },
    "vmSize": {
      "type": "string"
    },
    "winRMCertificateUrl": {
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
          "adminPassword": "[parameters('adminPassword')]",
          "secrets": [
            {
              "sourceVault": {
                "id": "[resourceId(resourceGroup().name, 'Microsoft.KeyVault/vaults', parameters('keyVaultName'))]"
              },
              "vaultCertificates": [
                {
                  "certificateUrl": "[parameters('winRMCertificateUrl')]",
                  "certificateStore": "My"
                }
              ]
            }
          ],
	  "windowsConfiguration": {
            "provisionVMAgent": "true",
            "winRM": {
              "listeners": [{
                  "protocol": "https",
                  "certificateUrl": "[parameters('winRMCertificateUrl')]"
                }
              ]
            },
            "enableAutomaticUpdates": "true"
          }
        },
        "storageProfile": {
          "imageReference": {
            "publisher": "[parameters('imagePublisher')]",
            "offer": "[parameters('imageOffer')]",
            "sku": "[parameters('imageSku')]",
	    "version": "[parameters('imageVersion')]"
          },
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
            "enabled": "false"
          }
        }
      }
    }
  ]
}`

const WindowsVhd = `{
  "$schema": "http://schema.management.azure.com/schemas/2014-04-01-preview/deploymentTemplate.json",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "adminUserName": {
      "type": "string"
    },
    "adminPassword": {
      "type": "securestring"
    },
    "dnsNameForPublicIP": {
      "type": "string"
    },
    "imageOffer": {
      "type": "string"
    },
    "imagePublisher": {
      "type": "string"
    },
    "imageSku": {
      "type": "string"
    },
    "imageVersion": {
      "type": "string"
    },
    "imageUri": {
      "type": "string"
    },
    "keyVaultName": {
      "type": "string"
    },
    "keyVaultSecretValue": {
      "type": "securestring"
    },
    "objectId": {
     "type": "string"
    },
    "osDiskName": {
      "type": "string"
    },
    "storageAccountBlobEndpoint": {
      "type": "string"
    },
    "tenantId": {
      "type": "string"
    },
    "vmName": {
      "type": "string"
    },
    "vmSize": {
      "type": "string"
    },
    "winRMCertificateUrl": {
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
          "adminPassword": "[parameters('adminPassword')]",
          "secrets": [
            {
              "sourceVault": {
                "id": "[resourceId(resourceGroup().name, 'Microsoft.KeyVault/vaults', parameters('keyVaultName'))]"
              },
              "vaultCertificates": [
                {
                  "certificateUrl": "[parameters('winRMCertificateUrl')]",
                  "certificateStore": "My"
                }
              ]
            }
          ],
	  "windowsConfiguration": {
            "provisionVMAgent": "true",
            "winRM": {
              "listeners": [{
                  "protocol": "https",
                  "certificateUrl": "[parameters('winRMCertificateUrl')]"
                }
              ]
            },
            "enableAutomaticUpdates": "true"
          }
        },
        "storageProfile": {
          "osDisk": {
            "name": "osdisk",
            "image": {
              "uri": "[parameters('imageUri')]"
            },
            "vhd": {
              "uri": "[concat(parameters('storageAccountBlobEndpoint'),variables('vmStorageAccountContainerName'),'/', parameters('osDiskName'),'.vhd')]"
            },
            "caching": "ReadWrite",
            "createOption": "FromImage"
            "osType": "windows"
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
            "enabled": "false"
          }
        }
      }
    }
  ]
}`
