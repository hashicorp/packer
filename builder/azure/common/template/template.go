package template

import (
	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	//"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
)

/////////////////////////////////////////////////
// Template
type Template struct {
	Schema         *string                `json:"$schema"`
	ContentVersion *string                `json:"contentVersion"`
	Parameters     *map[string]Parameters `json:"parameters"`
	Variables      *map[string]string     `json:"variables"`
	Resources      *[]Resource            `json:"resources"`
}

/////////////////////////////////////////////////
// Template > Parameters
type Parameters struct {
	Type         *string `json:"type"`
	DefaultValue *string `json:"defaultValue,omitempty"`
}

/////////////////////////////////////////////////
// Template > Resource
type Resource struct {
	ApiVersion *string     `json:"apiVersion"`
	Name       *string     `json:"name"`
	Type       *string     `json:"type"`
	Location   *string     `json:"location"`
	DependsOn  *[]string   `json:"dependsOn,omitempty"`
	Properties *Properties `json:"properties,omitempty"`
}

/////////////////////////////////////////////////
// Template > Resource > Properties
type Properties struct {
	AddressSpace            *network.AddressSpace               `json:"addressSpace,omitempty"`
	DiagnosticsProfile      *compute.DiagnosticsProfile         `json:"diagnosticsProfile,omitempty"`
	DNSSettings             *network.PublicIPAddressDNSSettings `json:"dnsSettings,omitempty"`
	HardwareProfile         *compute.HardwareProfile            `json:"hardwareProfile,omitempty"`
	IPConfigurations        *[]network.IPConfiguration          `json:"ipConfigurations,omitempty"`
	NetworkProfile          *compute.NetworkProfile             `json:"networkProfile,omitempty"`
	OsProfile               *compute.OSProfile                  `json:"osProfile,omitempty"`
	PublicIPAllocatedMethod *network.IPAllocationMethod         `json:"publicIPAllocationMethod,omitempty"`
	StorageProfile          *compute.StorageProfile             `json:"storageProfile,omitempty"`
	Subnets                 *[]network.Subnet                   `json:"subnets,omitempty"`
}
