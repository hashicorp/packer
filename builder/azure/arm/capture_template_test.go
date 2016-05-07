// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"encoding/json"
	"testing"
)

var captureTemplate01 = `{
    "operationId": "ac1c7c38-a591-41b3-89bd-ea39fceace1b",
    "status": "Succeeded",
    "startTime": "2016-04-04T21:07:25.2900874+00:00",
    "endTime": "2016-04-04T21:07:26.4776321+00:00",
    "properties": {
        "output": {
            "$schema": "http://schema.management.azure.com/schemas/2014-04-01-preview/VM_IP.json",
            "contentVersion": "1.0.0.0",
            "parameters": {
                "vmName": {
                    "type": "string"
                },
                "vmSize": {
                    "type": "string",
                    "defaultValue": "Standard_A2"
                },
                "adminUserName": {
                    "type": "string"
                },
                "adminPassword": {
                    "type": "securestring"
                },
                "networkInterfaceId": {
                    "type": "string"
                }
            },
            "resources": [
                {
                    "apiVersion": "2015-06-15",
                    "properties": {
                        "hardwareProfile": {
                            "vmSize": "[parameters('vmSize')]"
                        },
                        "storageProfile": {
                            "osDisk": {
                                "osType": "Linux",
                                "name": "packer-osDisk.32118633-6dc9-449f-83b6-a7d2983bec14.vhd",
                                "createOption": "FromImage",
                                "image": {
                                    "uri": "http://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.32118633-6dc9-449f-83b6-a7d2983bec14.vhd"
                                },
                                "vhd": {
                                    "uri": "http://storage.blob.core.windows.net/vmcontainerce1a1b75-f480-47cb-8e6e-55142e4a5f68/osDisk.ce1a1b75-f480-47cb-8e6e-55142e4a5f68.vhd"
                                },
                                "caching": "ReadWrite"
                            }
                        },
                        "osProfile": {
                            "computerName": "[parameters('vmName')]",
                            "adminUsername": "[parameters('adminUsername')]",
                            "adminPassword": "[parameters('adminPassword')]"
                        },
                        "networkProfile": {
                            "networkInterfaces": [
                                {
                                    "id": "[parameters('networkInterfaceId')]"
                                }
                            ]
                        },
                        "diagnosticsProfile": {
                            "bootDiagnostics": {
                                "enabled": false
                            }
                        },
                        "provisioningState": 0
                    },
                    "name": "[parameters('vmName')]",
                    "type": "Microsoft.Compute/virtualMachines",
                    "location": "southcentralus"
                }
            ]
        }
    }
}`

var captureTemplate02 = `{
    "operationId": "ac1c7c38-a591-41b3-89bd-ea39fceace1b",
    "status": "Succeeded",
    "startTime": "2016-04-04T21:07:25.2900874+00:00",
    "endTime": "2016-04-04T21:07:26.4776321+00:00"
}`

func TestCaptureParseJson(t *testing.T) {
	var operation CaptureOperation
	err := json.Unmarshal([]byte(captureTemplate01), &operation)
	if err != nil {
		t.Fatalf("failed to the sample capture operation: %s", err)
	}

	testSubject := operation.Properties.Output
	if testSubject.Schema != "http://schema.management.azure.com/schemas/2014-04-01-preview/VM_IP.json" {
		t.Errorf("Schema's value was unexpected: %s", testSubject.Schema)
	}
	if testSubject.ContentVersion != "1.0.0.0" {
		t.Errorf("ContentVersion's value was unexpected: %s", testSubject.ContentVersion)
	}

	// == Parameters ====================================
	if len(testSubject.Parameters) != 5 {
		t.Fatalf("expected parameters to have 5 keys, but got %d", len(testSubject.Parameters))
	}
	if _, ok := testSubject.Parameters["vmName"]; !ok {
		t.Errorf("Parameters['vmName'] was an expected parameters, but it did not exist")
	}
	if testSubject.Parameters["vmName"].Type != "string" {
		t.Errorf("Parameters['vmName'].Type == 'string', but got '%s'", testSubject.Parameters["vmName"].Type)
	}
	if _, ok := testSubject.Parameters["vmSize"]; !ok {
		t.Errorf("Parameters['vmSize'] was an expected parameters, but it did not exist")
	}
	if testSubject.Parameters["vmSize"].Type != "string" {
		t.Errorf("Parameters['vmSize'].Type == 'string', but got '%s'", testSubject.Parameters["vmSize"])
	}
	if testSubject.Parameters["vmSize"].DefaultValue != "Standard_A2" {
		t.Errorf("Parameters['vmSize'].DefaultValue == 'string', but got '%s'", testSubject.Parameters["vmSize"].DefaultValue)
	}

	// == Resources =====================================
	if len(testSubject.Resources) != 1 {
		t.Fatalf("expected resources to have length 1, but got %d", len(testSubject.Resources))
	}
	if testSubject.Resources[0].Name != "[parameters('vmName')]" {
		t.Errorf("Resources[0].Name's value was unexpected: %s", testSubject.Resources[0].Name)
	}
	if testSubject.Resources[0].Type != "Microsoft.Compute/virtualMachines" {
		t.Errorf("Resources[0].Type's value was unexpected: %s", testSubject.Resources[0].Type)
	}
	if testSubject.Resources[0].Location != "southcentralus" {
		t.Errorf("Resources[0].Location's value was unexpected: %s", testSubject.Resources[0].Location)
	}

	// == Resources/Properties =====================================
	if testSubject.Resources[0].Properties.ProvisioningState != 0 {
		t.Errorf("Resources[0].Properties.ProvisioningState's value was unexpected: %d", testSubject.Resources[0].Properties.ProvisioningState)
	}

	// == Resources/Properties/HardwareProfile ======================
	hardwareProfile := testSubject.Resources[0].Properties.HardwareProfile
	if hardwareProfile.VMSize != "[parameters('vmSize')]" {
		t.Errorf("Resources[0].Properties.HardwareProfile.VMSize's value was unexpected: %s", hardwareProfile.VMSize)
	}

	// == Resources/Properties/StorageProfile/OSDisk ================
	osDisk := testSubject.Resources[0].Properties.StorageProfile.OSDisk
	if osDisk.OSType != "Linux" {
		t.Errorf("Resources[0].Properties.StorageProfile.OSDisk.OSDisk's value was unexpected: %s", osDisk.OSType)
	}
	if osDisk.Name != "packer-osDisk.32118633-6dc9-449f-83b6-a7d2983bec14.vhd" {
		t.Errorf("Resources[0].Properties.StorageProfile.OSDisk.Name's value was unexpected: %s", osDisk.Name)
	}
	if osDisk.CreateOption != "FromImage" {
		t.Errorf("Resources[0].Properties.StorageProfile.OSDisk.CreateOption's value was unexpected: %s", osDisk.CreateOption)
	}
	if osDisk.Image.Uri != "http://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.32118633-6dc9-449f-83b6-a7d2983bec14.vhd" {
		t.Errorf("Resources[0].Properties.StorageProfile.OSDisk.Image.Uri's value was unexpected: %s", osDisk.Image.Uri)
	}
	if osDisk.Vhd.Uri != "http://storage.blob.core.windows.net/vmcontainerce1a1b75-f480-47cb-8e6e-55142e4a5f68/osDisk.ce1a1b75-f480-47cb-8e6e-55142e4a5f68.vhd" {
		t.Errorf("Resources[0].Properties.StorageProfile.OSDisk.Vhd.Uri's value was unexpected: %s", osDisk.Vhd.Uri)
	}
	if osDisk.Caching != "ReadWrite" {
		t.Errorf("Resources[0].Properties.StorageProfile.OSDisk.Caching's value was unexpected: %s", osDisk.Caching)
	}

	// == Resources/Properties/OSProfile ============================
	osProfile := testSubject.Resources[0].Properties.OSProfile
	if osProfile.AdminPassword != "[parameters('adminPassword')]" {
		t.Errorf("Resources[0].Properties.OSProfile.AdminPassword's value was unexpected: %s", osProfile.AdminPassword)
	}
	if osProfile.AdminUsername != "[parameters('adminUsername')]" {
		t.Errorf("Resources[0].Properties.OSProfile.AdminUsername's value was unexpected: %s", osProfile.AdminUsername)
	}
	if osProfile.ComputerName != "[parameters('vmName')]" {
		t.Errorf("Resources[0].Properties.OSProfile.ComputerName's value was unexpected: %s", osProfile.ComputerName)
	}

	// == Resources/Properties/NetworkProfile =======================
	networkProfile := testSubject.Resources[0].Properties.NetworkProfile
	if len(networkProfile.NetworkInterfaces) != 1 {
		t.Errorf("Count of Resources[0].Properties.NetworkProfile.NetworkInterfaces was expected to be 1, but go %d", len(networkProfile.NetworkInterfaces))
	}
	if networkProfile.NetworkInterfaces[0].Id != "[parameters('networkInterfaceId')]" {
		t.Errorf("Resources[0].Properties.NetworkProfile.NetworkInterfaces[0].Id's value was unexpected: %s", networkProfile.NetworkInterfaces[0].Id)
	}

	// == Resources/Properties/DiagnosticsProfile ===================
	diagnosticsProfile := testSubject.Resources[0].Properties.DiagnosticsProfile
	if diagnosticsProfile.BootDiagnostics.Enabled != false {
		t.Errorf("Resources[0].Properties.DiagnosticsProfile.BootDiagnostics.Enabled's value was unexpected: %t", diagnosticsProfile.BootDiagnostics.Enabled)
	}
}

func TestCaptureEmptyOperationJson(t *testing.T) {
	var operation CaptureOperation
	err := json.Unmarshal([]byte(captureTemplate02), &operation)
	if err != nil {
		t.Fatalf("failed to the sample capture operation: %s", err)
	}

	if operation.Properties != nil {
		t.Errorf("JSON contained no properties, but value was not nil: %+v", operation.Properties)
	}
}
