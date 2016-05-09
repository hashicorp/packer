// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

type CaptureTemplateParameter struct {
	Type         string `json:"type"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

type CaptureHardwareProfile struct {
	VMSize string `json:"vmSize"`
}

type CaptureUri struct {
	Uri string `json:"uri"`
}

type CaptureDisk struct {
	OSType       string     `json:"osType"`
	Name         string     `json:"name"`
	Image        CaptureUri `json:"image"`
	Vhd          CaptureUri `json:"vhd"`
	CreateOption string     `json:"createOption"`
	Caching      string     `json:"caching"`
}

type CaptureStorageProfile struct {
	OSDisk CaptureDisk `json:"osDisk"`
}

type CaptureOSProfile struct {
	ComputerName  string `json:"computerName"`
	AdminUsername string `json:"adminUsername"`
	AdminPassword string `json:"adminPassword"`
}

type CaptureNetworkInterface struct {
	Id string `json:"id"`
}

type CaptureNetworkProfile struct {
	NetworkInterfaces []CaptureNetworkInterface `json:"networkInterfaces"`
}

type CaptureBootDiagnostics struct {
	Enabled bool `json:"enabled"`
}

type CaptureDiagnosticProfile struct {
	BootDiagnostics CaptureBootDiagnostics `json:"bootDiagnostics"`
}

type CaptureProperties struct {
	HardwareProfile    CaptureHardwareProfile   `json:"hardwareProfile"`
	StorageProfile     CaptureStorageProfile    `json:"storageProfile"`
	OSProfile          CaptureOSProfile         `json:"osProfile"`
	NetworkProfile     CaptureNetworkProfile    `json:"networkProfile"`
	DiagnosticsProfile CaptureDiagnosticProfile `json:"diagnosticsProfile"`
	ProvisioningState  int                      `json:"provisioningState"`
}

type CaptureResources struct {
	ApiVersion string            `json:"apiVersion"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Location   string            `json:"location"`
	Properties CaptureProperties `json:"properties"`
}

type CaptureTemplate struct {
	Schema         string                              `json:"$schema"`
	ContentVersion string                              `json:"contentVersion"`
	Parameters     map[string]CaptureTemplateParameter `json:"parameters"`
	Resources      []CaptureResources                  `json:"resources"`
}

type CaptureOperationProperties struct {
	Output *CaptureTemplate `json:"output"`
}

type CaptureOperation struct {
	OperationId string                      `json:"operationId"`
	Status      string                      `json:"status"`
	Properties  *CaptureOperationProperties `json:"properties"`
}
