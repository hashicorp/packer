// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

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
